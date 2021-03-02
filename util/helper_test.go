// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VersionSuffix(t *testing.T) {
	assert.Equal(t, uint(1), VersionSuffix(""))
	for i := 2; i < 100; i++ {
		assert.Equal(t, uint(i), VersionSuffix(fmt.Sprintf("/v%d", i)))
	}
	// logged error
	assert.Equal(t, uint(1), VersionSuffix("/v1000000000000000000000000000000"))
}

func Test_SetVersionSuffix(t *testing.T) {
	assert.Equal(t, "", SetVersionSuffix("/v2", 0))
	assert.Equal(t, "", SetVersionSuffix("/v2", 1))
	assert.Equal(t, "//v444", SetVersionSuffix("/", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x", 444))
	assert.Equal(t, "/x/y/v444", SetVersionSuffix("/x/y", 444))
	assert.Equal(t, "/x/v/v444", SetVersionSuffix("/x/v", 444))
	assert.Equal(t, "/x/v0/v444", SetVersionSuffix("/x/v0", 444))
	assert.Equal(t, "/x/v1/v444", SetVersionSuffix("/x/v1", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v2", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v3", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v4", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v5", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v6", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v7", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v8", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v9", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v10", 444))
	assert.Equal(t, "/x/v444", SetVersionSuffix("/x/v11", 444))
	assert.Equal(t, "/x/v00/v444", SetVersionSuffix("/x/v00", 444))
	assert.Equal(t, "/x/v01/v444", SetVersionSuffix("/x/v01", 444))
}

func Test_RemoveVersionSuffix(t *testing.T) {
	assert.Equal(t, "/", RemoveVersionSuffix("/"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x"))
	assert.Equal(t, "/x/y", RemoveVersionSuffix("/x/y"))
	assert.Equal(t, "/x/v", RemoveVersionSuffix("/x/v"))
	assert.Equal(t, "/x/v0", RemoveVersionSuffix("/x/v0"))
	assert.Equal(t, "/x/v1", RemoveVersionSuffix("/x/v1"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v2"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v3"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v4"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v5"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v6"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v7"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v8"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v9"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v10"))
	assert.Equal(t, "/x", RemoveVersionSuffix("/x/v11"))
	assert.Equal(t, "/x/v00", RemoveVersionSuffix("/x/v00"))
	assert.Equal(t, "/x/v01", RemoveVersionSuffix("/x/v01"))
}

func assertParseURL(t *testing.T,
	expectedModule string,
	expectedVersion string,
	expectedAction string,
	path string) {
	t.Helper()
	module, version, action, err := ParseURL(path)
	assert.Equal(t, expectedModule, module)
	assert.Equal(t, expectedVersion, version)
	assert.Equal(t, expectedAction, action)
	assert.NoError(t, err)
}

func assertParseURLError(t *testing.T, path string) {
	t.Helper()
	module, version, action, err := ParseURL(path)
	assert.Empty(t, module)
	assert.Empty(t, version)
	assert.Empty(t, action)
	assert.Error(t, err)
}

func Test_ParseURL(t *testing.T) {
	assertParseURLError(t, "")
	assertParseURLError(t, "/")
	assertParseURLError(t, "//")
	assertParseURLError(t, "/my/")
	assertParseURLError(t, "/my/@something")
	assertParseURLError(t, "/my/@v")
	assertParseURLError(t, "/my/@v/")
	assertParseURLError(t, "/my/@v/1.0.0.help")
	assertParseURL(t, "my", "", "list", "/my/@v/list")
	assertParseURL(t, "my", "1.0.0", "info", "/my/@v/1.0.0.info")
	assertParseURL(t, "my", "1.0.0", "mod", "/my/@v/1.0.0.mod")
	assertParseURL(t, "my", "1.0.0", "zip", "/my/@v/1.0.0.zip")
	assertParseURL(t, "my", "latest", "info", "/my/@latest")
	assertParseURL(t, "my/project", "", "list", "/my/project/@v/list")
	assertParseURL(t, "my/project", "1.0.0", "info", "/my/project/@v/1.0.0.info")
	assertParseURL(t, "my/project", "1.0.0", "mod", "/my/project/@v/1.0.0.mod")
	assertParseURL(t, "my/project", "1.0.0", "zip", "/my/project/@v/1.0.0.zip")
	assertParseURL(t, "my/project", "latest", "info", "/my/project/@latest")
	assertParseURL(t, "my/project/v2", "1.0.0", "info", "/my/project/v2/@v/1.0.0.info")
	assertParseURL(t, "my/project/v2", "1.0.0", "mod", "/my/project/v2/@v/1.0.0.mod")
	assertParseURL(t, "my/project/v2", "1.0.0", "zip", "/my/project/v2/@v/1.0.0.zip")
	assertParseURL(t, "my/project/v2", "latest", "info", "/my/project/v2/@latest")
}

func Test_MergeVersions(t *testing.T) {
	assert.Equal(t, []string{"1"}, MergeVersions([]string{"1"}, []string{}))
	assert.Equal(t, []string{"1"}, MergeVersions([]string{"1"}, []string{"1"}))
	assert.Equal(t, []string{"1", "2"}, MergeVersions([]string{"1"}, []string{"2"}))
	assert.Equal(t, []string{"1", "2"}, MergeVersions([]string{"2"}, []string{"1"}))
}

func Test_TrimName(t *testing.T) {
	assert.Equal(t, "", TrimName("", ""))
	assert.Equal(t, "", TrimName("", "skip"))
	assert.Equal(t, "", TrimName("", "skip/"))
	assert.Equal(t, "/a", TrimName("", "skip/a"))
	assert.Equal(t, "", TrimName("", "skip/a/"))
	assert.Equal(t, "/a-a", TrimName("", "skip/a-a"))
	assert.Equal(t, "", TrimName("", "skip/a-a/"))
	assert.Equal(t, "/a/b", TrimName("", "skip/a/b"))
	assert.Equal(t, "", TrimName("", "skip/a/b/"))
	assert.Equal(t, "/a/b/c", TrimName("", "skip/a/b/c"))
	assert.Equal(t, "", TrimName("", "skip/a/b/c/"))
	assert.Equal(t, "/a/b/c/file.txt", TrimName("", "skip/a/b/c/file.txt"))
	assert.Equal(t, "/a-a/b/c/file.txt", TrimName("", "skip/a-a/b/c/file.txt"))
	assert.Equal(t, "", TrimName("a", ""))
	assert.Equal(t, "", TrimName("a", "skip"))
	assert.Equal(t, "", TrimName("a", "skip/"))
	assert.Equal(t, "", TrimName("a", "skip/a"))
	assert.Equal(t, "", TrimName("a", "skip/a/"))
	assert.Equal(t, "", TrimName("a", "skip/a-a"))
	assert.Equal(t, "", TrimName("a", "skip/a-a/"))
	assert.Equal(t, "/b", TrimName("a", "skip/a/b"))
	assert.Equal(t, "", TrimName("a", "skip/a/b/"))
	assert.Equal(t, "/b/c", TrimName("a", "skip/a/b/c"))
	assert.Equal(t, "", TrimName("a", "skip/a/b/c/"))
	assert.Equal(t, "/b/c/file.txt", TrimName("a", "skip/a/b/c/file.txt"))
	assert.Equal(t, "", TrimName("a", "skip/a-a/b/c/file.txt"))
}

func Test_VersionDir(t *testing.T) {
	assert.Equal(t, "", VersionDir("v0.0.0"))
	assert.Equal(t, "", VersionDir("v1.0.0"))
	assert.Equal(t, "/v2", VersionDir("v2.0.0"))
	assert.Equal(t, "/v3", VersionDir("v3.0.0"))
	assert.Panics(t, func() {
		VersionDir("")
	})
}

func Test_UnifyDir(t *testing.T) {
	assert.Equal(t, "", UnifyDir("/"))
	assert.Equal(t, "x", UnifyDir("/x"))
	assert.Equal(t, "x", UnifyDir("x/"))
	assert.Equal(t, "x", UnifyDir("/x/"))
	assert.Equal(t, "x/y", UnifyDir("x/y"))
	assert.Equal(t, "x/y", UnifyDir("/x/y"))
	assert.Equal(t, "x/y", UnifyDir("x/y/"))
	assert.Equal(t, "x/y", UnifyDir("/x/y/"))
}
