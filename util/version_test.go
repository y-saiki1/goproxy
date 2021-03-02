// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var boolToValid = map[bool]string{
	false: "invalid",
	true:  "valid",
}

func assertRegexp(t *testing.T, r *regexp.Regexp, valid bool, v string) {
	assert.Equalf(t, valid, r.MatchString(v),
		"expected %q as %s %q", v, boolToValid[valid], r.String())
}

func Test_versionCorePattern(t *testing.T) {
	r := regexp.MustCompile(`^` + versionCorePattern + `$`)
	for major := 0; major < 12; major++ {
		for minor := 0; minor < 12; minor++ {
			for patch := 0; patch < 12; patch++ {
				v := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				assertRegexp(t, r, true, v)
			}
		}
	}
	invalidVersions := []string{
		"1",
		"1.0",
		"1.0.0.0",
		"01.0.0",
		"1.1.1a",
	}
	for _, v := range invalidVersions {
		assertRegexp(t, r, false, v)
	}
}

func Test_alphanumIdentPattern(t *testing.T) {
	r := regexp.MustCompile(`^` + alphanumIdentPattern + `$`)
	valid := []string{
		"a",
		"A",
		"-",

		"1a",
		"1A",
		"1-",
		"a1",
		"A1",
		"-1",
		"1a1",
		"1A1",
		"1-1",

		"aa",
		"aA",
		"a-",
		"aa",
		"Aa",
		"-a",
		"aaa",
		"aAa",
		"a-a",

		"Aa",
		"AA",
		"A-",
		"aA",
		"AA",
		"-A",
		"AaA",
		"AAA",
		"A-A",

		"-a",
		"-A",
		"--",
		"a-",
		"A-",
		"--",
		"-a-",
		"-A-",
		"---",
	}
	invalid := []string{
		"*",
		"1",
		"11",
	}
	for _, v := range valid {
		assertRegexp(t, r, true, v)
	}
	for _, v := range invalid {
		assertRegexp(t, r, false, v)
	}
}

func Test_ParseVersion(t *testing.T) {
	valid := map[string]Version{
		"0.0.0": {
			Major:      0,
			Minor:      0,
			Patch:      0,
			PreRelease: "",
			Build:      "",
		},
		"0.0.1-alpha": {
			Major:      0,
			Minor:      0,
			Patch:      1,
			PreRelease: "alpha",
			Build:      "",
		},
		"0.0.0+abcd": {
			Major:      0,
			Minor:      0,
			Patch:      0,
			PreRelease: "",
			Build:      "abcd",
		},
		"1.0.0": {
			Major:      1,
			Minor:      0,
			Patch:      0,
			PreRelease: "",
			Build:      "",
		},
	}
	invalid := []string{
		"0",
		"1",
		"1.0",
		"1.0.0.0",
		"10000000000000000000000000000000.1.1",
		"1.10000000000000000000000000000000.1",
		"1.1.10000000000000000000000000000000",
	}
	for in, out := range valid {
		v, err := ParseVersion(in)
		assert.Equal(t, out, v)
		assert.NoError(t, err)
	}
	for _, in := range invalid {
		v, err := ParseVersion(in)
		assert.Empty(t, v)
		assert.Error(t, err)
	}
}

func Test_ParseTagVersion(t *testing.T) {
	valid := map[string]Version{
		"v0.0.0": {
			Major:      0,
			Minor:      0,
			Patch:      0,
			PreRelease: "",
			Build:      "",
		},
		"v1.0.0": {
			Major:      1,
			Minor:      0,
			Patch:      0,
			PreRelease: "",
			Build:      "",
		},
	}
	invalid := []string{
		"1.0.0",
	}
	for in, out := range valid {
		v, err := ParseTagVersion(in)
		assert.Equal(t, out, v)
		assert.NoError(t, err)
	}
	for _, in := range invalid {
		v, err := ParseTagVersion(in)
		assert.Empty(t, v)
		assert.Error(t, err)
	}
}

func Test_Version_Compare(t *testing.T) {
	a := Version{}
	b := Version{}
	assert.Equal(t, 0, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      1,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, 1, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      1,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, -1, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, 1, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, -1, a.Compare(b))
	a = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, 1, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, -1, a.Compare(b))
	a = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      5,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, 1, a.Compare(b))
	a = Version{
		Major:      0,
		Minor:      5,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, -1, a.Compare(b))
	a = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "alfa.1",
		Build:      "",
	}
	b = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, -1, a.Compare(b))
	assert.Equal(t, 1, b.Compare(a))
	b.PreRelease = "alfa.1"
	assert.Equal(t, 0, a.Compare(b))
	assert.Equal(t, 0, b.Compare(a))
}

func Test_Version_Latest(t *testing.T) {
	a := Version{}
	b := Version{}
	assert.Equal(t, a, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      1,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, a, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      1,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, b, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, a, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      1,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, b, a.Latest(b))
	a = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, a, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      0,
		Patch:      5,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, b, a.Latest(b))
	a = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      0,
		Minor:      5,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, a, a.Latest(b))
	a = Version{
		Major:      0,
		Minor:      5,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	b = Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}
	assert.Equal(t, b, a.Latest(b))
}

func Test_Version_Valid(t *testing.T) {
	v := Version{}
	assert.True(t, v.Valid())
	v.PreRelease = "abcd"
	assert.True(t, v.Valid())
	v.Build = "123"
	assert.True(t, v.Valid())
	v.PreRelease = ""
	assert.True(t, v.Valid())
	v.PreRelease = "*"
	assert.False(t, v.Valid())
	v.Build = ""
	assert.False(t, v.Valid())
	v.Build = "*"
	assert.False(t, v.Valid())
	v.PreRelease = "0"
	assert.False(t, v.Valid())
	v.PreRelease = "abcd"
	assert.False(t, v.Valid())
}

func assertVersionMarshalText(t *testing.T, expected string, actual Version) {
	t.Helper()
	b, err := actual.MarshalText()
	assert.Equal(t, string(b), expected)
	assert.NoError(t, err)
}

func Test_Version_MarshalText(t *testing.T) {
	v := Version{}
	assertVersionMarshalText(t, "0.0.0", v)
	v.Major = 1
	assertVersionMarshalText(t, "1.0.0", v)
	v.PreRelease = "abcd"
	assertVersionMarshalText(t, "1.0.0-abcd", v)
	v.Build = "123"
	assertVersionMarshalText(t, "1.0.0-abcd+123", v)
	v.PreRelease = ""
	assertVersionMarshalText(t, "1.0.0+123", v)
}

func Test_Version_UnmarshalText_ParseVersion(t *testing.T) {
	v := Version{}
	require.NoError(t, v.UnmarshalText([]byte("2.1.3-a+b")))
	assert.Equal(t, Version{
		Major:      2,
		Minor:      1,
		Patch:      3,
		PreRelease: "a",
		Build:      "b",
	}, v)
}
func Test_Version_UnmarshalText_ParseTagVersion(t *testing.T) {
	v := Version{}
	require.NoError(t, v.UnmarshalText([]byte("v2.1.3-a.1")))
	assert.Equal(t, Version{
		Major:      2,
		Minor:      1,
		Patch:      3,
		PreRelease: "a.1",
		Build:      "",
	}, v)
}

func Test_Version_UnmarshalText_Fail(t *testing.T) {
	v := Version{}
	require.Error(t, v.UnmarshalText([]byte("")))
	assert.Empty(t, v)
}

func Test_Version_TagString(t *testing.T) {
	v := Version{}
	assert.Equal(t, "v0.0.0", v.TagString())
	v.Major = 1
	assert.Equal(t, "v1.0.0", v.TagString())
	v.PreRelease = "abcd"
	assert.Equal(t, "v1.0.0-abcd", v.TagString())
	v.Build = "123"
	assert.Equal(t, "v1.0.0-abcd+123", v.TagString())
	v.PreRelease = ""
	assert.Equal(t, "v1.0.0+123", v.TagString())
}

func Test_Version_String(t *testing.T) {
	v := Version{}
	assert.Equal(t, "0.0.0", v.String())
	v.Major = 1
	assert.Equal(t, "1.0.0", v.String())
	v.PreRelease = "abcd"
	assert.Equal(t, "1.0.0-abcd", v.String())
	v.Build = "123"
	assert.Equal(t, "1.0.0-abcd+123", v.String())
	v.PreRelease = ""
	assert.Equal(t, "1.0.0+123", v.String())
}

func Test_CompareVersions(t *testing.T) {
	c, err := CompareVersions("0.0.0", "0.0.0")
	assert.Equal(t, 0, c)
	assert.NoError(t, err)
	c, err = CompareVersions("1.0.0", "0.0.0")
	assert.Equal(t, 1, c)
	assert.NoError(t, err)
	c, err = CompareVersions("0.0.0", "1.0.0")
	assert.Equal(t, -1, c)
	assert.NoError(t, err)
	c, err = CompareVersions("1.0", "0.0.0")
	assert.Empty(t, c)
	assert.Error(t, err)
	c, err = CompareVersions("0.0.0", "1.0")
	assert.Empty(t, c)
	assert.Error(t, err)
}

func Test_CompareTagVersions(t *testing.T) {
	c, err := CompareTagVersions("v0.0.0", "v0.0.0")
	assert.Equal(t, 0, c)
	assert.NoError(t, err)
	c, err = CompareTagVersions("v1.0.0", "v0.0.0")
	assert.Equal(t, 1, c)
	assert.NoError(t, err)
	c, err = CompareTagVersions("v0.0.0", "v1.0.0")
	assert.Equal(t, -1, c)
	assert.NoError(t, err)
	c, err = CompareTagVersions("v1.0", "v0.0.0")
	assert.Empty(t, c)
	assert.Error(t, err)
	c, err = CompareTagVersions("v0.0.0", "v1.0")
	assert.Empty(t, c)
	assert.Error(t, err)
}

func Test_LatestVersion(t *testing.T) {
	l, err := LatestVersion("0.0.0", "0.0.0")
	assert.Equal(t, Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestVersion("1.0.0", "0.0.0")
	assert.Equal(t, Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestVersion("0.0.0", "1.0.0")
	assert.Equal(t, Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestVersion("1.0", "0.0.0")
	assert.Empty(t, l)
	assert.Error(t, err)
	l, err = LatestVersion("0.0.0", "1.0")
	assert.Empty(t, l)
	assert.Error(t, err)
	l, err = LatestVersion("0.0", "1.0")
	assert.Empty(t, l)
	assert.Error(t, err)
}

func Test_LatestTagVersion(t *testing.T) {
	l, err := LatestTagVersion("v0.0.0", "v0.0.0")
	assert.Equal(t, Version{
		Major:      0,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestTagVersion("v1.0.0", "v0.0.0")
	assert.Equal(t, Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestTagVersion("v0.0.0", "v1.0.0")
	assert.Equal(t, Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "",
		Build:      "",
	}, l)
	assert.NoError(t, err)
	l, err = LatestTagVersion("v1.0", "v0.0.0")
	assert.Empty(t, l)
	assert.Error(t, err)
	l, err = LatestTagVersion("v0.0.0", "v1.0")
	assert.Empty(t, l)
	assert.Error(t, err)
	l, err = LatestTagVersion("v0.0", "v1.0")
	assert.Empty(t, l)
	assert.Error(t, err)
}
