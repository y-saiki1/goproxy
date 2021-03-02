// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package storage

import (
	"errors"
	"time"

	"go.lstv.dev/goproxy/util"
)

var ErrCurrentlyLocked = errors.New("currently locked")

func IsCurrentlyLocked(err error) bool {
	return errors.Is(err, ErrCurrentlyLocked)
}

type StoredModuleInfo struct {
	Name      string
	Versions  []StoredModuleVersionInfo
	TotalSize int64
}

type StoredModuleVersionInfo struct {
	Version    string
	Downloaded time.Time
	Size       int64
	Locked     bool
}

type sortModuleVersionsDesc StoredModuleInfo

func (s sortModuleVersionsDesc) Len() int {
	return len(s.Versions)
}

func (s sortModuleVersionsDesc) Less(i, j int) bool {
	c, _ := util.CompareTagVersions(s.Versions[i].Version, s.Versions[j].Version)
	return c > 0
}

func (s sortModuleVersionsDesc) Swap(i, j int) {
	s.Versions[i], s.Versions[j] = s.Versions[j], s.Versions[i]
}
