// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package client

import (
	"html/template"
	"net/http"

	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/storage"
	"go.lstv.dev/goproxy/util"
)

var indexTemplate *template.Template

func init() {
	t := template.New("index").Funcs(FuncMap())
	indexTemplate = mustParseTemplate(t.Parse(indexTemplateContent))
}

func mustParseTemplate(t *template.Template, err error) *template.Template {
	if err != nil {
		panic(err)
	}
	return t
}

type Client interface {
	ConfiguredModules() [][]string
	StoredModules() ([]storage.StoredModuleInfo, error)
	ConfiguredDownloads() [][]string
}

func ServeClient(c Client, w http.ResponseWriter, _ *http.Request) {
	storedModulesInfo, storedModulesInfoErr := c.StoredModules()
	values := map[string]any{
		"Version":             util.BuiltinVersion(),
		"ConfiguredModules":   c.ConfiguredModules(),
		"StoredModules":       storedModulesInfo,
		"StoredModulesErr":    storedModulesInfoErr,
		"ConfiguredDownloads": c.ConfiguredDownloads(),
	}
	logger.Type("client.ServeClient").NoErr(indexTemplate.Execute(w, values))
}
