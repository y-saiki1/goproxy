// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"go.lstv.dev/goproxy/util"
)

var (
	ErrNotParametrized = errors.New("source is not parametrized")
	ErrNotRegistered   = errors.New("source type is not registered")

	sourcesMutex sync.Mutex
	sources      = map[string]func(map[string]interface{}) (Source, error){}
)

type VersionNotFoundError struct {
	Err error
}

func NewVersionNotFoundError(err error) *VersionNotFoundError {
	return &VersionNotFoundError{
		Err: err,
	}
}

func (v *VersionNotFoundError) Unwrap() error {
	return v.Err
}

func (v *VersionNotFoundError) Error() string {
	if v == nil || v.Err == nil {
		return "version not found"
	}
	return fmt.Sprintf("version not found: %s", v.Err.Error())
}

func IsVersionNotFound(err error) bool {
	var e *VersionNotFoundError
	return errors.As(err, &e)
}

func Register(name string, builder func(map[string]interface{}) (Source, error)) bool {
	sourcesMutex.Lock()
	defer sourcesMutex.Unlock()
	if _, ok := sources[name]; ok {
		return false
	}
	sources[name] = builder
	return true
}

func New(name string, params map[string]interface{}) (Source, error) {
	if b := builder(name); b != nil {
		return b(params)
	}
	return nil, ErrNotRegistered
}

type Source interface {
	// Parametrize returns new source with specified parameters.
	Parametrize(module string, params map[string]interface{}) (Source, error)

	// ConfigPreview returns key-value pairs of configuration preview.
	ConfigPreview() (pairs []string)

	// ListVersions returns list of all versions at form v1.0.0.
	// For major 1, also major 0 is used.
	ListVersions(ctx context.Context, major uint) ([]string, error)

	// LatestVersion returns latest version with specified major.
	// For major 1, also major 0 is used.
	// Latest stable version is released if exist.
	LatestVersion(ctx context.Context, major uint) (string, error)

	// DownloadModule download module files at specified version to specified directory.
	//
	// For version v1.0.0 and directory /tmp/package are created files:
	//
	//   /tmp/package/v1.0.0.lock (temporary)
	//   /tmp/package/v1.0.0.tmp (temporary)
	//   /tmp/package/v1.0.0.info
	//   /tmp/package/v1.0.0.mod
	//   /tmp/package/v1.0.0.zip
	//
	// For version v2.0.0 and directory /tmp/package are created files:
	//
	//   /tmp/package/v2.0.0.lock (temporary)
	//   /tmp/package/v2.0.0.tmp (temporary)
	//   /tmp/package/v2.0.0.info
	//   /tmp/package/v2.0.0.mod
	//   /tmp/package/v2.0.0.zip
	//
	// Lock file is created first and removed after function is done.
	DownloadModule(ctx context.Context, dir, version string) error

	// ParametrizeDownloads returns new Downloads with specified parameters.
	ParametrizeDownloads(name, mode string, params map[string]interface{}) (Downloads, error)
}

type Downloads interface {
	// ConfigPreview returns key-value pairs of configuration preview.
	ConfigPreview() (pairs []string)

	WriteDownload(ctx context.Context, w http.ResponseWriter, v util.Version, arch string)

	LatestDownloadVersion(ctx context.Context) (latest util.Version, err error)
}

func builder(name string) func(map[string]interface{}) (Source, error) {
	sourcesMutex.Lock()
	defer sourcesMutex.Unlock()
	return sources[name]
}
