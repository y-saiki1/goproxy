// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package client

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_formatTime(t *testing.T) {
	assert.Equal(t, "2022-07-08 04:05:09", formatTime(time.Date(2022, 7, 8, 4, 5, 9, 100, time.UTC)))
	assert.Equal(t, "2022-07-08 17:05:09", formatTime(time.Date(2022, 7, 8, 17, 5, 9, 100, time.UTC)))
}

func Test_formatSize(t *testing.T) {
	assert.Equal(t, template.HTML("-1&nbsp;&nbsp;&nbsp;&nbsp;B&nbsp;&nbsp;"), formatSize(-1))
	assert.Equal(t, template.HTML("0&nbsp;&nbsp;&nbsp;&nbsp;B&nbsp;&nbsp;"), formatSize(0))
	assert.Equal(t, template.HTML("1023&nbsp;&nbsp;&nbsp;&nbsp;B&nbsp;&nbsp;"), formatSize(1023))
	assert.Equal(t, template.HTML("1.00&nbsp;kiB"), formatSize(1024))
	assert.Equal(t, template.HTML("1024.00&nbsp;kiB"), formatSize(1024*1024-1))
	assert.Equal(t, template.HTML("1.00&nbsp;MiB"), formatSize(1024*1024))
	assert.Equal(t, template.HTML("100.00&nbsp;MiB"), formatSize(100*1024*1024))
}

func Test_isZeroOrEven(t *testing.T) {
	assert.True(t, isZeroOrEven(0))
	assert.False(t, isZeroOrEven(1))
	assert.True(t, isZeroOrEven(2))
}

func Test_configuredModuleType(t *testing.T) {
	assert.Equal(t, "configured-module-package", configuredModuleType([]string{""}))
	assert.Equal(t, "configured-module-disabled", configuredModuleType([]string{"", "type", "null"}))
}
