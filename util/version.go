// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	ZeroVersion    = "0.0.0"
	ZeroTagVersion = "v0.0.0"

	digitsPattern          = `\d+`
	nonDigitPattern        = `[A-Za-z\-]`
	identPattern           = `[0-9A-Za-z\-]`
	numIdentPattern        = `0|[1-9]\d*`
	alphanumIdentPattern   = `(?:` + identPattern + `*` + nonDigitPattern + identPattern + `*)`
	versionCorePattern     = `(` + numIdentPattern + `)\.(` + numIdentPattern + `)\.(` + numIdentPattern + `)`
	preReleasePattern      = preReleaseIdentPattern + `(?:\.` + preReleaseIdentPattern + `)*`
	preReleaseIdentPattern = `(?:` + alphanumIdentPattern + `|(?:` + numIdentPattern + `))`
	buildPattern           = buildIdentPattern + `(?:\.` + buildIdentPattern + `)*`
	buildIdentPattern      = `(?:` + alphanumIdentPattern + `|` + digitsPattern + `)`
	semverPattern          = `^` + versionCorePattern + `(?:\-(` + preReleasePattern + `))?(?:\+(` + buildPattern + `))?$`
)

var semverRegexp = regexp.MustCompile(semverPattern)

type InvalidVersionFormatError string

func (e InvalidVersionFormatError) Error() string {
	return "invalid version format: " + string(e)
}

type Version struct {
	Major, Minor, Patch uint
	PreRelease, Build   string
}

func (v Version) Compare(ver Version) int {
	if ver.Major > v.Major {
		return -1
	} else if ver.Major < v.Major {
		return 1
	}
	if ver.Minor > v.Minor {
		return -1
	} else if ver.Minor < v.Minor {
		return 1
	}
	if ver.Patch > v.Patch {
		return -1
	} else if ver.Patch < v.Patch {
		return 1
	}
	return strings.Compare(ver.PreRelease, v.PreRelease)
}

func (v Version) Latest(ver Version) Version {
	if v.Compare(ver) == -1 {
		return ver
	}
	return v
}

func (v Version) Valid() bool {
	return semverRegexp.MatchString(v.String())
}

func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Version) UnmarshalText(data []byte) error {
	version, err := ParseVersion(string(data))
	if err != nil {
		version, err = ParseTagVersion(string(data))
		if err != nil {
			return err
		}
	}
	*v = version
	return nil
}

func (v Version) TagString() string {
	return "v" + v.String()
}

func (v Version) String() string {
	sb := strings.Builder{}
	sb.WriteString(strconv.FormatUint(uint64(v.Major), 10))
	sb.WriteByte('.')
	sb.WriteString(strconv.FormatUint(uint64(v.Minor), 10))
	sb.WriteByte('.')
	sb.WriteString(strconv.FormatUint(uint64(v.Patch), 10))
	if v.PreRelease != "" {
		sb.WriteByte('-')
		sb.WriteString(v.PreRelease)
	}
	if v.Build != "" {
		sb.WriteByte('+')
		sb.WriteString(v.Build)
	}
	return sb.String()
}

func ParseTagVersion(v string) (Version, error) {
	if !strings.HasPrefix(v, "v") {
		return Version{}, InvalidVersionFormatError(v)
	}
	return ParseVersion(v[1:])
}

func ParseVersion(v string) (Version, error) {
	parts := semverRegexp.FindStringSubmatch(v)
	if len(parts) == 0 {
		return Version{}, InvalidVersionFormatError(v)
	}
	major, err := strconv.Atoi(parts[1])
	if err != nil || major < 0 {
		return Version{}, InvalidVersionFormatError(v)
	}
	minor, err := strconv.Atoi(parts[2])
	if err != nil || minor < 0 {
		return Version{}, InvalidVersionFormatError(v)
	}
	patch, err := strconv.Atoi(parts[3])
	if err != nil || patch < 0 {
		return Version{}, InvalidVersionFormatError(v)
	}
	return Version{
		Major:      uint(major),
		Minor:      uint(minor),
		Patch:      uint(patch),
		PreRelease: parts[4],
		Build:      parts[5],
	}, nil
}

// CompareVersions compares passed versions.
// Error is returned if passed versions are not valid.
func CompareVersions(a, b string) (int, error) {
	av, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	bv, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	return av.Compare(bv), nil
}

// CompareTagVersions compares passed tag versions.
// Error is returned if passed tag versions are not valid.
func CompareTagVersions(a, b string) (int, error) {
	av, err := ParseTagVersion(a)
	if err != nil {
		return 0, err
	}
	bv, err := ParseTagVersion(b)
	if err != nil {
		return 0, err
	}
	return av.Compare(bv), nil
}

func LatestVersion(a, b string) (Version, error) {
	av, err := ParseVersion(a)
	if err != nil {
		return Version{}, err
	}
	bv, err := ParseVersion(b)
	if err != nil {
		return Version{}, err
	}
	return av.Latest(bv), nil
}

func LatestTagVersion(a, b string) (Version, error) {
	av, err := ParseTagVersion(a)
	if err != nil {
		return Version{}, err
	}
	bv, err := ParseTagVersion(b)
	if err != nil {
		return Version{}, err
	}
	return av.Latest(bv), nil
}
