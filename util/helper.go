// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.lstv.dev/goproxy/logger"
)

var versionSuffixPattern = regexp.MustCompile(`/v(?:[2-9]|[1-9][0-9]+)$`)

func VersionSuffix(module string) (major uint) {
	if indexes := versionSuffixPattern.FindStringSubmatchIndex(module); len(indexes) != 0 {
		i, err := strconv.ParseUint(module[indexes[0]+2:], 10, 64)
		if err != nil {
			logger.Type("helper.VersionSuffix").Err(err).With(
				"module", module,
			).Error("invalid major version")
			return 1
		}
		return uint(i)
	}
	return 1
}

func SetVersionSuffix(module string, major uint) (moduleWithVersionSuffix string) {
	m := RemoveVersionSuffix(module)
	if major <= 1 {
		return m
	}
	return fmt.Sprintf("%s/v%d", m, major)
}

func RemoveVersionSuffix(module string) (moduleWithoutVersionSuffix string) {
	if indexes := versionSuffixPattern.FindStringSubmatchIndex(module); len(indexes) != 0 {
		return module[:indexes[0]]
	}
	return module
}

func ParseURL(path string) (module, version, action string, err error) {
	if len(path) == 0 {
		return "", "", "", errors.New("expected / at [0]")
	}
	s := path[1:]
	if strings.HasSuffix(s, "/@latest") {
		return s[:len(s)-8], "latest", "info", nil
	}
	v := strings.Index(s, "/@v/")
	if v < 0 {
		return "", "", "", errors.New("expected @v")
	}
	module = s[:v]
	s = s[v+4:]
	if s == "list" {
		return module, "", "list", nil
	}
	switch {
	case strings.HasSuffix(s, ".info"):
		return module, s[:len(s)-5], "info", nil
	case strings.HasSuffix(s, ".mod"):
		return module, s[:len(s)-4], "mod", nil
	case strings.HasSuffix(s, ".zip"):
		return module, s[:len(s)-4], "zip", nil
	default:
		return "", "", "", errors.New("expected suffix .info, .mod or .zip")
	}
}

func MergeVersions(a, b []string) []string {
	if len(b) == 0 {
		return a
	}
	m := make(map[string]struct{}, len(a))
	for _, v := range a {
		m[v] = struct{}{}
	}
	for _, v := range b {
		m[v] = struct{}{}
	}
	versions := make([]string, 0, len(m))
	for v := range m {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	return versions
}

// TrimName remove first level of directory and parameter dir.
// If current name is directory, returns empty string.
// If current dir is not empty and name is not part of dir, returns empty string.
// If returned string is not empty, it starts with slash.
func TrimName(dir, name string) string {
	if IsDir(name) {
		return ""
	}
	s := TrimFirstDir(name)
	if s == "" {
		return ""
	}
	if dir == "" {
		return "/" + s
	}
	if i := strings.IndexByte(s, '/'); i > 0 && s[:i] == dir {
		return s[len(dir):]
	}
	return ""
}

func TrimFirstDir(name string) string {
	if i := strings.IndexByte(name, '/'); i > 0 {
		return name[i+1:]
	}
	return ""
}

func IsDir(name string) bool {
	l := len(name)
	return l > 0 && name[len(name)-1] == '/'
}

// UnifyDir returns dir excluding starting and ending slash.
// If dir is empty or "/", returns empty string.
func UnifyDir(dir string) string {
	return strings.Trim(dir, "/")
}

func VersionDir(version string) string {
	if v, err := ParseTagVersion(version); err != nil {
		logger.Type("helper.VersionDir").Err(err).Panic("unable to parse version")
	} else if v.Major > 1 {
		return fmt.Sprintf("/v%d", v.Major)
	}
	return ""
}
