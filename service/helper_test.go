// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package service

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setContentType(t *testing.T) {
	cases := []struct {
		Key    string
		Result string
	}{
		{
			Key:    "info",
			Result: "application/json; charset=UTF-8",
		},
		{
			Key:    "json",
			Result: "application/json; charset=UTF-8",
		},
		{
			Key:    "mod",
			Result: "text/plain; charset=UTF-8",
		},
		{
			Key:    "text",
			Result: "text/plain; charset=UTF-8",
		},
		{
			Key:    "zip",
			Result: "application/zip",
		},
	}

	for _, c := range cases {
		r := httptest.NewRecorder()
		setContentType(r, c.Key)
		assert.Equalf(t, c.Result, r.Header().Get("Content-Type"), "unexpected content type for %s", c.Key)
	}
}
