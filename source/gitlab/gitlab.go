// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gitlab

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/source"
	"go.lstv.dev/goproxy/util"
)

const Type = "gitlab"

func init() {
	source.Register(Type, New)
}

type Source struct {
	log         logger.Logger
	url         string
	auth        string
	insecureTLS bool
	client      *http.Client
	params      *params
}

func New(config map[string]interface{}) (source.Source, error) {
	if config == nil {
		return nil, errors.New("gitlab.New: expected url and auth")
	}
	url, ok := config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("gitlab.New: expected url as string instead of %T", config["url"])
	}
	auth, ok := config["auth"].(string)
	if !ok {
		return nil, fmt.Errorf("gitlab.New: expected auth as string instead of %T", config["auth"])
	}
	allowInsecureTLS, _ := config["allow_insecure_tls"].(bool)
	g := &Source{
		log: logger.Type("gitlab.Source").With(
			"url", url,
		),
		url:         url,
		auth:        auth,
		insecureTLS: allowInsecureTLS,
		client:      &http.Client{},
	}
	if allowInsecureTLS {
		g.allowInsecureTLS()
	}
	return g, nil
}

func (s *Source) apiURL(relativePath string) string {
	const apiSuffix = "api/v4/"
	l := len(s.url)
	if l == 0 {
		return ""
	}
	if s.url[l-1] == '/' {
		return s.url + apiSuffix + relativePath
	}
	return s.url + "/" + apiSuffix + relativePath
}

func (s *Source) allowInsecureTLS() {
	s.log.Info("allowed insecure tls")
	s.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func (s *Source) Parametrize(module string, params map[string]interface{}) (source.Source, error) {
	p, err := newParams(module, params)
	if err != nil {
		return nil, err
	}
	return &Source{
		log: s.log.With(
			"module", p.module,
			"project_id", p.projectID,
			"dir", p.dir,
			"tag_prefix", p.tagPrefix,
			"version_dir", p.versionDir,
		),
		url:         s.url,
		auth:        s.auth,
		insecureTLS: s.insecureTLS,
		client:      s.client,
		params:      p,
	}, nil
}

func (s *Source) ConfigPreview() (pairs []string) {
	return []string{
		"type", "gitlab",
		"url", s.url,
		"project_id", strconv.FormatInt(s.params.projectID, 10),
		"dir", s.params.dir,
		"tag_prefix", s.params.tagPrefix,
		"insecure_tls", strconv.FormatBool(s.insecureTLS),
	}
}

func (s *Source) ListVersions(ctx context.Context, major uint) ([]string, error) {
	log := s.log.Ctx(ctx).With(
		"func", "ListVersions",
	)
	if s.params == nil {
		log.Error("not parametrized source")
		return nil, source.ErrNotParametrized
	}
	url := s.apiURL(fmt.Sprintf("projects/%d/repository/tags?search=^%sv",
		s.params.projectID,
		s.params.tagPrefix,
	))
	resp, err := s.doGetRequest(ctx, url)
	if err != nil {
		log.Err(err).Debug("request failed")
		return nil, fmt.Errorf("ListVersions: request failed: %w", err)
	}
	defer s.log.NoErrClose(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.With(
			"status_code", resp.StatusCode,
		).Debug("request failed: unexpected status code")
		return nil, fmt.Errorf("ListVersions: request failed: status code %d", resp.StatusCode)
	}
	content := []struct {
		Name string `json:"name"`
	}(nil)
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		log.Err(err).Debug("invalid response")
		return nil, fmt.Errorf("ListVersions: invalid response: %w", err)
	}
	versions := []string(nil)
	tagPrefixLength := len(s.params.tagPrefix)
	for _, t := range content {
		version := t.Name[tagPrefixLength:]
		if v, err := util.ParseTagVersion(version); err != nil {
			if isK8S(version) {
				continue
			}
			log.Err(err).Debug("invalid tag version")
		} else if v.Major == major || (v.Major == 0 && major == 1) {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

func (s *Source) LatestVersion(ctx context.Context, major uint) (string, error) {
	log := s.log.Ctx(ctx).With(
		"func", "LatestVersion",
	)
	versions, err := s.ListVersions(ctx, major)
	if err != nil {
		log.Err(err).Debug("list versions failed")
		return "", fmt.Errorf("LatestVersion: %w", err)
	}
	latest := util.Version{}
	latestStable := util.Version{}
	for _, version := range versions {
		v, err := util.ParseTagVersion(version)
		if err != nil {
			return "", fmt.Errorf("LatestVersion: %w", err)
		}
		latest = latest.Latest(v)
		if v.PreRelease == "" {
			latestStable = latestStable.Latest(v)
		}
	}
	if latestStable != (util.Version{}) {
		latest = latestStable
	}
	log.With(
		"latest_version", latest,
	).Debug("latest version")
	return latest.TagString(), nil
}

func (s *Source) DownloadModule(ctx context.Context, dir, version string) error {
	c := logger.ContextWith(ctx,
		"dir", dir,
		"version", version,
	)
	log := s.log.Ctx(c).With(
		"func", "DownloadModule",
	)

	if s.params == nil {
		log.Error("not parametrized source")
		return source.ErrNotParametrized
	}

	commit, timestamp, err := s.findCommit(c, version)
	if err != nil {
		log.Err(err).Debug("failed get commit for version")
		return err
	}

	if err := os.MkdirAll(filepath.Join(dir, s.params.module), 0755); err != nil {
		log.Err(err).Debug("unable to create directories")
	}

	lockPath := filepath.Join(dir, s.params.module, version+".lock")
	lockContent := []byte(time.Now().String())
	if err := os.WriteFile(lockPath, lockContent, 0755); err != nil {
		log.Err(err).Debug("unable to create lock file")

		if b, _ := os.ReadFile(lockPath); bytes.Compare(b, lockContent) == 0 {
			s.log.NoErr(os.Remove(lockPath))
		}

		return fmt.Errorf("unable to create lock file: %w", err)
	}
	defer func() {
		s.log.NoErr(os.Remove(lockPath))
	}()

	tmpPath := filepath.Join(dir, s.params.module, version+".tmp")
	defer func() {
		s.log.NoErr(os.Remove(tmpPath))
	}()
	if err := s.fetchArchive(ctx, tmpPath, commit); err != nil {
		log.Err(err).Debug("unable to get archive")
		return err
	}
	if err := s.saveModule(c, dir, version, timestamp, tmpPath); err != nil {
		log.Err(err).Debug("unable to save module")
	}
	return nil
}

func (s *Source) ParametrizeDownloads(name, mode string, params map[string]interface{}) (source.Downloads, error) {
	if mode != "generic-packages" {
		return nil, fmt.Errorf("ParametrizeDownloads: invalid mode %q", mode)
	}
	projectIDNumber, ok := params["project_id"].(json.Number)
	if !ok {
		return nil, fmt.Errorf("ParametrizeDownloads: expected project_id as json.Number instead of %T", params["project_id"])
	}
	projectID, err := projectIDNumber.Int64()
	if err != nil {
		return nil, fmt.Errorf("ParametrizeDownloads: invalid project_id %w", err)
	}
	packageName := name // default package name
	if packageNameInterface, ok := params["package_name"]; ok {
		// packageName must be variable from outer scope
		if packageName, ok = packageNameInterface.(string); !ok {
			return nil, fmt.Errorf("ParametrizeDownloads: expected package_name as string instead of %T", packageNameInterface)
		}
	}
	disableArchitecture, _ := params["disable_architecture"].(bool)
	fileExtension, _ := params["file_extension"].(string)
	return &downloads{
		Source:              s,
		name:                name,
		projectID:           projectID,
		packageName:         packageName,
		disableArchitecture: disableArchitecture,
		fileExtension:       fileExtension,
	}, nil
}

func (s *Source) doGetRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", s.auth)
	return s.client.Do(req)
}

func (s *Source) findCommit(ctx context.Context, version string) (commit, timestamp string, err error) {
	tag := s.params.tagPrefix + version
	url := s.apiURL(fmt.Sprintf("projects/%d/repository/tags/%s",
		s.params.projectID,
		tag,
	))
	resp, err := s.doGetRequest(ctx, url)
	if err != nil {
		return "", "", fmt.Errorf("findCommit: request failed: %w", err)
	}
	defer s.log.NoErrClose(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return "", "", fmt.Errorf("findCommit: %w", newTagNotFoundError(tag))
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("findCommit: request failed: status code %d", resp.StatusCode)
	}
	obj := &struct {
		Commit struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
		} `json:"commit"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(obj); err != nil {
		return "", "", fmt.Errorf("findCommit: invalid response: %w", err)
	}
	return obj.Commit.ID, obj.Commit.CreatedAt, nil
}

func (s *Source) fetchArchive(ctx context.Context, file, commit string) error {
	url := s.apiURL(fmt.Sprintf("/projects/%d/repository/archive.zip?sha=%s",
		s.params.projectID,
		commit,
	))
	resp, err := s.doGetRequest(ctx, url)
	if err != nil {
		return fmt.Errorf("fetchArchive: request failed: %w", err)
	}
	defer s.log.NoErrClose(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("fetchArchive: commit not found %q", commit)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetchArchive: request failed: status code %d", resp.StatusCode)
	}
	if err := s.saveArchive(file, resp.Body); err != nil {
		return fmt.Errorf("fetchArchive: unable to create file: %w", err)
	}
	return nil
}

func (s *Source) saveArchive(file string, r io.Reader) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer s.log.NoErrClose(f)
	_, err = io.Copy(f, r)
	return err
}

func (s *Source) saveModule(ctx context.Context, dir, version, timestamp, archivePath string) error {
	log := s.log.Ctx(ctx).With(
		"func", "saveModule",
	)
	log.With(
		"timestamp", timestamp,
		"archive_path", archivePath,
	).Debug("called")

	infoPath := filepath.Join(dir, s.params.module, version+".info")
	modPath := filepath.Join(dir, s.params.module, version+".mod")
	zipPath := filepath.Join(dir, s.params.module, version+".zip")

	done := false
	defer func() {
		if !done {
			log.Trace("not done, removing incomplete files")
			log.NoErr(os.Remove(infoPath))
			log.NoErr(os.Remove(modPath))
			log.NoErr(os.Remove(zipPath))
		}
	}()

	infoFile, err := os.Create(infoPath)
	if err != nil {
		log.With(
			"file", infoFile,
		).Error("unable to create info file")
		return fmt.Errorf("saveModule: unable to create info file: %w", err)
	}
	defer log.NoErrClose(infoFile)

	modFile, err := os.Create(modPath)
	if err != nil {
		log.With(
			"file", modFile,
		).Error("unable to create mod file")
		return fmt.Errorf("saveModule: unable to create mod file: %w", err)
	}
	defer log.NoErrClose(modFile)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		log.With(
			"file", zipFile,
		).Error("unable to create zip file")
		return fmt.Errorf("saveModule: unable to create zip file: %w", err)
	}
	defer log.NoErrClose(zipFile)

	if err := s.writeInfo(infoFile, version, timestamp); err != nil {
		log.With(
			"file", infoFile,
		).Error("unable to write info file")
		return fmt.Errorf("saveModule: unable to write info file: %w", err)
	}
	if err := s.writeZip(ctx, zipFile, modFile, version, archivePath); err != nil {
		return fmt.Errorf("saveModule: unable to write zip file: %w", err)
	}

	done = true
	return nil
}

func (s *Source) writeInfo(w io.Writer, version, timestamp string) error {
	info := struct {
		Version string // version string
		Time    string // commit time
	}{
		Version: version,
		Time:    timestamp,
	}
	return json.NewEncoder(w).Encode(info)
}

func (s *Source) writeZip(ctx context.Context, zipW, modW io.Writer, version, archivePath string) error {
	log := s.log.Ctx(ctx).With(
		"func", "writeZip",
	)

	dir := s.params.dir
	if s.params.versionDir {
		dir += util.VersionDir(version)
	}
	log.With(
		"dir", dir,
	).Trace("use content of dir")

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	w := zip.NewWriter(zipW)
	defer log.NoErrClose(w)
	for _, f := range r.File {
		name := util.TrimName(dir, f.Name)
		l := log.With(
			"name", name,
			"original_name", f.Name,
		)
		if name == "" {
			l.Trace("file skipped")
			continue
		}
		if name == "/go.mod" {
			l.Trace("found go.mod")
			if err := s.writeMod(modW, f); err != nil {
				return err
			}
		}
		name = s.params.module + util.VersionDir(version) + "@" + version + name
		if err := s.writeZipFile(f, name, w); err != nil {
			return err
		}
		l.With(
			"full_name", name,
		).Trace("file written to zip")
	}
	return nil
}

func (s *Source) writeZipFile(f *zip.File, name string, w *zip.Writer) error {
	fh := f.FileHeader
	fh.Name = name
	fw, err := w.CreateHeader(&fh)
	if err != nil {
		return err
	}
	fr, err := f.Open()
	if err != nil {
		return err
	}
	defer s.log.NoErrClose(fr)
	_, err = io.Copy(fw, fr)
	return err
}

func (s *Source) writeMod(w io.Writer, f *zip.File) error {
	fr, err := f.Open()
	if err != nil {
		return err
	}
	defer s.log.NoErrClose(fr)
	_, err = io.Copy(w, fr)
	return err
}
