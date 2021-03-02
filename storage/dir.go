// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/util"
)

type Dir struct {
	Chroot string
}

func (d *Dir) ModuleDir(module string) string {
	return filepath.Join(d.Chroot, util.RemoveVersionSuffix(module))
}

func (d *Dir) LatestVersion(module string, major uint) (string, error) {
	versions, err := d.ListVersions(module, &major)
	if err != nil {
		return "", err
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
	return latest.TagString(), nil
}

func (d *Dir) ListModules() ([]string, error) {
	return d.listModules("")
}

func (d *Dir) listModules(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(d.ModuleDir(dir))
	if err != nil {
		return nil, err
	}

	list := []string(nil)
	containsModules := false
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			modules, err := d.listModules(path.Join(dir, dirEntry.Name()))
			if err != nil {
				return nil, err
			}
			list = append(list, modules...)
			continue
		}
		if !containsModules && strings.HasSuffix(dirEntry.Name(), ".info") {
			containsModules = true
			list = append(list, dir)
		}
	}
	sort.Strings(list)
	return list, nil
}

func (d *Dir) ListVersions(module string, major *uint) ([]string, error) {
	dir := d.ModuleDir(module)
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	versions := make(map[string]struct{})
	locked := make(map[string]struct{})
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		switch {
		case strings.HasSuffix(name, ".info"):
			versions[name[:len(name)-5]] = struct{}{}
		case strings.HasSuffix(name, ".lock"):
			locked[name[:len(name)-5]] = struct{}{}
		}
	}

	result := make([]string, 0, len(versions))
	for version := range versions {
		if _, ok := locked[version]; !ok {
			if v, err := util.ParseTagVersion(version); err != nil {
				logger.Type("storage.ListVersions").Err(err).Warn("unexpected version format")
			} else if major == nil || v.Major == *major {
				result = append(result, version)
			}
		}
	}
	return result, nil
}

func (d *Dir) Open(module, version, suffix string) (io.ReadCloser, error) {
	if ok, err := d.IsLocked(module, version); ok {
		return nil, ErrCurrentlyLocked
	} else if err != nil {
		return nil, err
	}
	dir := d.ModuleDir(module)
	file := filepath.Join(dir, version+"."+suffix)
	return os.Open(file)
}

func (d *Dir) HasVersion(module, version string) (bool, error) {
	if ok, err := d.IsLocked(module, version); ok {
		return false, ErrCurrentlyLocked
	} else if err != nil {
		return false, err
	}
	dir := d.ModuleDir(module)
	return checkFile(filepath.Join(dir, version+".info"))
}

func (d *Dir) StoredModules() ([]StoredModuleInfo, error) {
	info := []StoredModuleInfo(nil)
	modules, err := d.ListModules()
	if err != nil {
		return nil, err
	}
	for _, module := range modules {
		moduleInfo := StoredModuleInfo{
			Name: module,
		}
		versions, err := d.ListVersions(module, nil)
		if err != nil {
			return nil, err
		}
		for _, version := range versions {
			size, downloaded, err := d.moduleVersionInfo(module, version)
			if IsCurrentlyLocked(err) {
				moduleInfo.Versions = append(moduleInfo.Versions,
					StoredModuleVersionInfo{
						Version:    version,
						Downloaded: time.Time{},
						Size:       0,
						Locked:     true,
					},
				)
				continue
			}
			if err != nil {
				return nil, err
			}
			moduleInfo.Versions = append(moduleInfo.Versions,
				StoredModuleVersionInfo{
					Version:    version,
					Downloaded: downloaded,
					Size:       size,
					Locked:     false,
				},
			)
			moduleInfo.TotalSize += size
		}
		info = append(info, moduleInfo)
	}

	for i := range info {
		sort.Sort(sortModuleVersionsDesc(info[i]))
	}

	return info, nil
}

func (d *Dir) IsLocked(module, version string) (bool, error) {
	dir := d.ModuleDir(module)
	return checkFile(filepath.Join(dir, version+".lock"))
}

func (d *Dir) moduleVersionInfo(module, version string) (size int64, downloaded time.Time, err error) {
	if locked, err := d.IsLocked(module, version); err != nil {
		return 0, time.Time{}, err
	} else if locked {
		return 0, time.Time{}, ErrCurrentlyLocked
	}
	file := filepath.Join(d.ModuleDir(module), version+".zip")
	if s, err := os.Stat(file); err != nil {
		return 0, time.Time{}, err
	} else {
		return s.Size(), s.ModTime(), nil
	}
}

func checkFile(file string) (bool, error) {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
