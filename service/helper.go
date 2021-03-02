// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package service

import (
	"net/http"
)

var contentTypes = map[string]string{
	"info": "application/json; charset=UTF-8",
	"json": "application/json; charset=UTF-8",
	"mod":  "text/plain; charset=UTF-8",
	"text": "text/plain; charset=UTF-8",
	"zip":  "application/zip",
}

func setContentType(w http.ResponseWriter, key string) {
	w.Header().Set("Content-Type", contentTypes[key])
}
