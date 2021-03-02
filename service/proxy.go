// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"go.lstv.dev/goproxy/client"
	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/source"
	"go.lstv.dev/goproxy/storage"
	"go.lstv.dev/goproxy/util"
)

const DefaultDownloadsPathPrefix = "/dl"

type GoProxy struct {
	log                 logger.Logger
	server              http.Server
	versions            VersionsConfig
	defaultGoProxyURL   string // exclude ending slash
	downloadsPathPrefix string // include starting slash, exclude ending slash
	modules             map[string]source.Source
	downloads           map[string]source.Downloads
	sources             map[string]source.Source
	files               storage.Dir
}

func NewGoProxy(config *Config) (*GoProxy, error) {
	log := logger.Type("service.GoProxy")

	// configuring default go proxy url
	if config.DefaultGoProxyURL == "" {
		log.Fatal("missing default_go_proxy_url configuration")
	}
	defaultGoProxyURL := config.DefaultGoProxyURL
	if _, err := url.Parse(defaultGoProxyURL); err != nil {
		return nil, fmt.Errorf("invalid default_go_proxy_url: %w", err)
	}
	if strings.HasSuffix(defaultGoProxyURL, "/") {
		return nil, errors.New("invalid default_go_proxy_url: unexpected ending slash")
	}
	log.With(
		"default_go_proxy_url", defaultGoProxyURL,
	).Info("configured default go proxy url")

	// configuring downloads path prefix
	downloadsPathPrefix := DefaultDownloadsPathPrefix
	if config.DownloadsPrefix != "" {
		downloadsPathPrefix = "/" + url.PathEscape(config.DownloadsPrefix)
	}
	log.With(
		"downloads_path_prefix", downloadsPathPrefix,
	).Info("configured downloads path prefix")

	// create new GoProxy
	p := &GoProxy{
		log: log,
		server: http.Server{
			Addr: config.Addr,
		},
		versions:            config.Versions,
		defaultGoProxyURL:   defaultGoProxyURL,
		downloadsPathPrefix: downloadsPathPrefix,
		modules:             map[string]source.Source{},
		downloads:           map[string]source.Downloads{},
		sources:             map[string]source.Source{},
		files: storage.Dir{
			Chroot: config.Storage,
		},
	}
	p.server.Handler = p
	if err := p.loadSources(config); err != nil {
		return nil, err
	}
	if err := p.loadModules(config); err != nil {
		return nil, err
	}
	if err := p.loadDownloads(config); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *GoProxy) loadSources(config *Config) error {
	for i, s := range config.Sources {
		name, ok := s["name"].(string)
		if !ok {
			return fmt.Errorf("invalid source [%d]: expected name as string", i)
		}
		if _, ok := p.sources[name]; ok {
			return fmt.Errorf("invalid source [%d]: name already used", i)
		}
		typ, ok := s["type"].(string)
		if !ok {
			return fmt.Errorf("invalid source [%d]: expected type as string", i)
		}
		delete(s, "name")
		delete(s, "type")
		s, err := source.New(typ, s)
		if err != nil {
			return fmt.Errorf("invalid source [%d]: unable to create source: %w", i, err)
		}
		p.log.With(
			"name", name,
			"type", typ,
		).Info("added source")
		p.sources[name] = s
	}
	return nil
}

func (p *GoProxy) loadModules(config *Config) error {
	for i, m := range config.Modules {
		if _, ok := p.modules[m.Name]; ok {
			return fmt.Errorf("invalid module [%d]: name already used", i)
		}
		if m.Source == nil {
			// added disabled module
			p.log.With(
				"name", m.Name,
				"source", nil,
			).Info("added module")
			p.modules[m.Name] = nil
			continue
		}
		s, ok := p.sources[*m.Source]
		if !ok {
			return fmt.Errorf("invalid module [%d]: invalid source %q", i, *m.Source)
		}
		ps, err := s.Parametrize(m.Name, m.SourceParams)
		if err != nil {
			return fmt.Errorf("invalid module [%d]: unable to parametrize source: %w", i, err)
		}
		p.log.With(
			"name", m.Name,
			"source", *m.Source,
		).Info("added module")
		p.modules[m.Name] = ps
	}
	return nil
}

func (p *GoProxy) loadDownloads(config *Config) error {
	for name, d := range config.Downloads {
		s, ok := p.sources[d.Source]
		if !ok {
			return fmt.Errorf("invalid downloads [%s]: invalid source %q", name, d.Source)
		}
		ds, err := s.ParametrizeDownloads(name, d.Mode, d.SourceParams)
		if err != nil {
			return fmt.Errorf("invalid downloads [%s]: unable to parametrize source: %w", name, err)
		}
		p.log.With(
			"name", name,
			"source", d.Source,
		).Info("added downloads")
		p.downloads[name] = ds
	}
	return nil
}

func (p *GoProxy) Start() error {
	p.log.With(
		"addr", p.server.Addr,
		"version", util.BuiltinVersion(),
	).Info("start")
	return p.server.ListenAndServe()
}

func (p *GoProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		client.ServeClient(p, w, req)
		return
	case "/favicon.ico":
		http.NotFound(w, req)
		return
	case "/healthz":
		w.WriteHeader(http.StatusNoContent)
		return
	case "/versions.json":
		p.Versions(w, req)
		return
	}

	ctx := logger.ContextWith(req.Context(),
		"request_id", util.GenerateUniqueID(),
	)
	// check downloads prefix
	if path := req.URL.Path; strings.HasPrefix(path, p.downloadsPathPrefix) {
		p.serveDownload(ctx, w, path[len(p.downloadsPathPrefix):], req.URL.Query())
		return
	}
	// parse url
	module, version, action, err := util.ParseURL(req.URL.Path)
	if err != nil {
		p.log.Ctx(ctx).Err(err).With(
			"url", req.URL.Path,
		).Debug("unknown url")
	}
	// if module is not configured, fallthrough to default go proxy
	s, ok := p.modules[util.RemoveVersionSuffix(module)]
	if err != nil || !ok {
		defaultGoProxyURL := p.defaultGoProxyURL + req.URL.Path
		p.log.Ctx(ctx).With(
			"url", defaultGoProxyURL,
		).Debug("redirect")
		http.Redirect(w, req, defaultGoProxyURL, http.StatusTemporaryRedirect)
		return
	}
	// if fallthrough is disabled
	if s == nil {
		p.log.Ctx(ctx).With(
			"url", req.URL.String(),
		).Debug("fallthrough disabled")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// handle only GET method
	if req.Method != http.MethodGet {
		p.log.Ctx(ctx).Debug("expected GET method")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// handle list action
	if action == "list" {
		if err := p.serveList(ctx, w, module, s); err != nil {
			p.log.Ctx(ctx).Err(err).With(
				"module", module,
			).Debug("unable to list module versions")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// get latest version
	if version == "latest" {
		latest, err := p.latestVersion(ctx, module, s)
		if err != nil {
			p.log.Ctx(ctx).Err(err).With(
				"module", module,
			).Debug("unable to get latest version")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		p.log.Ctx(ctx).With(
			"module", module,
			"version", version,
		).Debug("translate latest to version")
		version = latest
	}
	// if there is no stored version, download module
	if ok, err := p.files.HasVersion(module, version); !ok {
		if err != nil {
			p.log.Ctx(ctx).Err(err).Debug("unable to check module version")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := s.DownloadModule(ctx, p.files.Chroot, version); err != nil {
			p.log.Ctx(ctx).Err(err).Debug("unable to download module")
			if source.IsVersionNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// serve stored module
	if err := p.serve(ctx, w, module, version, action); err != nil {
		p.log.Ctx(ctx).Err(err).Debug("unable to serve response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *GoProxy) ModuleNames() []string {
	names := make([]string, 0, len(p.modules))
	for n := range p.modules {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (p *GoProxy) ConfiguredModules() [][]string {
	names := p.ModuleNames()
	modules := make([][]string, 0, len(names))
	for _, n := range names {
		c := []string{n}
		if m := p.modules[n]; m != nil {
			c = append(c, m.ConfigPreview()...)
		} else {
			c = append(c,
				"type", "null",
				"fallthrough", "disabled",
			)
		}
		modules = append(modules, c)
	}
	return modules
}

func (p *GoProxy) StoredModules() ([]storage.StoredModuleInfo, error) {
	return p.files.StoredModules()
}

func (p *GoProxy) DownloadNames() []string {
	names := make([]string, 0, len(p.downloads))
	for n := range p.downloads {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (p *GoProxy) ConfiguredDownloads() [][]string {
	names := p.DownloadNames()
	downloads := make([][]string, 0, len(names))
	for _, n := range names {
		c := []string{n}
		if d := p.downloads[n]; d != nil {
			c = append(c, d.ConfigPreview()...)
		}
		downloads = append(downloads, c)
	}
	return downloads
}

func (p *GoProxy) Versions(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := p.log.With(
		"func", "Versions",
	)

	latestVersions := map[string]interface{}{}
	content := map[string]interface{}{
		"go_version":      p.versions.Go.String(),
		"latest_versions": latestVersions,
	}
	for moduleWithoutVersionSuffix, s := range p.modules {
		if s == nil {
			continue
		}
		module, version, err := p.latestMajorVersion(ctx, moduleWithoutVersionSuffix, s)
		if err != nil {
			log.Err(err).With(
				"module", moduleWithoutVersionSuffix,
			).Error("unable to get module major version")
			continue
		}
		latestVersions[module] = version
	}
	for _, module := range p.versions.Modules {
		version, err := p.latestVersionFromDefaultProxy(ctx, module)
		if err != nil {
			log.Err(err).With(
				"module", module,
			).Error("unable to get module version")
		}
		latestVersions[module] = version
	}

	setContentType(w, "json")
	if err := json.NewEncoder(w).Encode(content); err != nil {
		log.Err(err).Error("unable to encode versions")
		return
	}
}

func (p *GoProxy) serveList(ctx context.Context, w http.ResponseWriter, module string, s source.Source) error {
	log := p.log.Ctx(ctx).With(
		"func", "serveList",
	)
	versions, err := s.ListVersions(ctx, util.VersionSuffix(module))
	if err != nil {
		return err
	}
	major := util.VersionSuffix(module)
	storedVersions, err := p.files.ListVersions(module, &major)
	if err != nil {
		log.Ctx(ctx).Err(err).With(
			"module", module,
		).Info("unable to get list of stored versions")
	}
	versions = util.MergeVersions(versions, storedVersions)
	log.Ctx(ctx).With(
		"module", module,
		"action", "list",
	).Info("serve")
	setContentType(w, "text")
	for _, v := range versions {
		log.NoErrLast(w.Write([]byte(v)))
		log.NoErrLast(w.Write([]byte("\r\n")))
	}
	return nil
}

func (p *GoProxy) latestVersion(ctx context.Context, module string, s source.Source) (string, error) {
	major := util.VersionSuffix(module)
	version, err := s.LatestVersion(ctx, major)
	if err != nil {
		return "", err
	}
	storedVersion, err := p.files.LatestVersion(module, major)
	if err != nil {
		p.log.Ctx(ctx).Err(err).With(
			"module", module,
		).Info("unable to get latest stored version")
		return version, nil
	}
	latestVersion, err := util.LatestTagVersion(version, storedVersion)
	if err != nil {
		return "", fmt.Errorf("latestVersion: unable to compare %q and %q: %w", version, storedVersion, err)
	}
	return latestVersion.TagString(), nil
}

func (p *GoProxy) latestMajorVersion(ctx context.Context, module string, s source.Source) (moduleWithVersionSuffix, version string, err error) {
	moduleWithVersionSuffix = module
	version, err = p.latestVersion(ctx, module, s)
	if err != nil {
		return "", "", err
	}
	if version == util.ZeroTagVersion {
		return "", "", fmt.Errorf("unable to get first version of module %q", module)
	}
	for i := uint(2); ; i++ {
		m := util.SetVersionSuffix(module, i)
		v, err := p.latestVersion(ctx, m, s)
		if err != nil || v == util.ZeroTagVersion {
			break
		}
		moduleWithVersionSuffix = m
		version = v
	}
	return moduleWithVersionSuffix, version, nil
}

func (p *GoProxy) serve(ctx context.Context, w http.ResponseWriter, module, version, action string) error {
	log := p.log.Ctx(ctx).With(
		"module", module,
		"version", version,
		"action", action,
	)
	switch action {
	case "info", "mod", "zip":
		log.Trace("serve")
		return p.serveFile(ctx, w, module, version, action)
	default:
		log.Trace("unknown action")
		return nil
	}
}

func (p *GoProxy) serveFile(_ context.Context, w http.ResponseWriter, module, version, suffix string) error {
	f, err := p.files.Open(module, version, suffix)
	if err != nil {
		return fmt.Errorf("unable to open %q file for module %q at version %q: %w", suffix, module, version, err)
	}
	defer p.log.NoErrClose(f)
	setContentType(w, suffix)
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("unable to read %q file for module %q at version %q: %w", suffix, module, version, err)
	}
	return nil
}

func (p *GoProxy) serveDownload(ctx context.Context, w http.ResponseWriter, relativePath string, query url.Values) {
	if relativePath == "/versions.json" {
		p.serveDownloadVersions(ctx, w, query.Get("filter"))
		return
	}

	log := p.log.Ctx(ctx).With(
		"download_path", relativePath,
	)

	parts := strings.Split(relativePath, "/")
	switch len(parts) {
	case 3:
		parts = append(parts, "")
	case 4:
		// no-op
	default:
		log.Debug("invalid url")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if parts[0] != "" {
		log.Debug("expected / after download prefix")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log = log.With(
		"name", parts[1],
		"version", parts[2],
		"arch", parts[3],
	)

	ds, ok := p.downloads[parts[1]]
	if !ok {
		log.Debug("invalid download name")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	v := util.Version{}
	if err := (error)(nil); parts[2] == "latest" {
		v, err = ds.LatestDownloadVersion(ctx)
		if err != nil {
			log.Err(err).Debug("invalid download version")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log = log.With(
			"latest_version", v,
		)
	} else {
		v, err = util.ParseVersion(parts[2])
		if err != nil {
			log.Err(err).Debug("invalid download version")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	ds.WriteDownload(ctx, w, v, parts[3])
	log.Trace("download finished")
}

func (p *GoProxy) serveDownloadVersions(ctx context.Context, w http.ResponseWriter, filter string) {
	log := p.log.Ctx(ctx)
	result := struct {
		LatestVersions map[string]util.Version `json:"latest_versions"`
	}{
		LatestVersions: map[string]util.Version{},
	}
	for name, download := range p.downloads {
		if filter != "" && name != filter {
			continue
		}
		v, err := download.LatestDownloadVersion(ctx)
		if err != nil {
			log.Err(err).With(
				"name", name,
			).Debug("invalid download version")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result.LatestVersions[name] = v
	}
	setContentType(w, "json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Err(err).Warn("download versions encoding failed")
	}
}

func (p *GoProxy) latestVersionFromDefaultProxy(ctx context.Context, module string) (util.Version, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.defaultGoProxyURL+"/"+module+"/@latest", http.NoBody)
	if err != nil {
		return util.Version{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return util.Version{}, err
	}
	defer p.log.Ctx(ctx).NoErrClose(resp.Body)
	v := struct {
		Version string `json:"Version"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return util.Version{}, err
	}
	return util.ParseTagVersion(v.Version)
}
