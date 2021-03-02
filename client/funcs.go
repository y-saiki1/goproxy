// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package client

import (
	"fmt"
	"html/template"
	"time"
)

func FuncMap() map[string]interface{} {
	return map[string]interface{}{
		"formatTime":           formatTime,
		"formatSize":           formatSize,
		"isZeroOrEven":         isZeroOrEven,
		"configuredModuleType": configuredModuleType,
	}
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

func formatSize(size int64) template.HTML {
	if size >= 1024*1024 {
		return divideUnit(size, 1024*1024, "MiB")
	}
	if size >= 1024 {
		return divideUnit(size, 1024, "kiB")
	}
	return template.HTML(fmt.Sprintf("%d&nbsp;&nbsp;&nbsp;&nbsp;B&nbsp;&nbsp;", size))
}

func isZeroOrEven(i int) bool {
	return i%2 == 0
}

func configuredModuleType(configuredModule []string) string {
	s := configuredModule[1:]
	for i := 0; i < len(s); i += 2 {
		if s[i] == "type" && s[i+1] == "null" {
			return "configured-module-disabled"
		}
	}
	return "configured-module-package"
}

func divideUnit(size int64, divider float64, unit string) template.HTML {
	f := float64(size) / divider
	return template.HTML(fmt.Sprintf("%.02f&nbsp;%s", f, unit))
}
