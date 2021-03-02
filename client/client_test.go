// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package client

import (
	_ "embed"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"go.lstv.dev/goproxy/storage"

	"github.com/stretchr/testify/assert"
)

//go:embed client_test_expected_index.html
var expectedIndexHTML string

type clientMock struct{}

func (_ *clientMock) ConfiguredModules() [][]string {
	return [][]string{
		{
			"example.com/module",
			"type", "null",
			"fallthrough", "disabled",
		},
		{
			"example.com/module/a",
			"type", "gitlab",
			"url", "https://gitlab.example.com",
			"project_id", "1",
			"dir", "a",
			"tag_prefix", "a-",
			"insecure_tls", "true",
		},
		{
			"example.com/module/b",
			"type", "gitlab",
			"url", "https://gitlab.example.com",
			"project_id", "2",
			"dir", "b",
			"tag_prefix", "b-",
			"insecure_tls", "true",
		},
	}
}

func (_ *clientMock) StoredModules() ([]storage.StoredModuleInfo, error) {
	return []storage.StoredModuleInfo{
		{
			Name: "example.com/module/a",
			Versions: []storage.StoredModuleVersionInfo{
				{
					Version:    "v1.14.0",
					Downloaded: time.Date(2022, 7, 8, 4, 5, 9, 100, time.UTC),
					Size:       5000,
					Locked:     false,
				},
			},
			TotalSize: 5000,
		},
		{
			Name: "example.com/module/b",
			Versions: []storage.StoredModuleVersionInfo{
				{
					Version:    "v1.5.0",
					Downloaded: time.Date(2022, 1, 4, 8, 7, 3, 100, time.UTC),
					Size:       4000,
					Locked:     false,
				},
			},
			TotalSize: 4000,
		},
	}, nil
}

func (_ *clientMock) ConfiguredDownloads() [][]string {
	return [][]string{
		{
			"type", "gitlab",
			"url", "https://gitlab.example.com",
			"project_id", "1",
			"package_name", "example.com/module/a",
			"insecure_tls", "true",
		},
	}
}

func Test_mustParseTemplate(t *testing.T) {
	assert.Panics(t, func() {
		mustParseTemplate(nil, errors.New(""))
	})
}

func Test_ServeClient(t *testing.T) {
	w := httptest.NewRecorder()
	ServeClient((*clientMock)(nil), w, nil)
	assert.Equal(t, expectedIndexHTML, w.Body.String())
}
