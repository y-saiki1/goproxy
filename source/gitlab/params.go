// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.lstv.dev/goproxy/util"
)

type params struct {
	module     string
	projectID  int64
	dir        string // excluding starting and ending slash, e.g. "hello/world"
	tagPrefix  string
	versionDir bool
}

func newParams(module string, p map[string]any) (*params, error) {
	if p == nil {
		return nil, errors.New("newGitlabParams: expected project_id")
	}
	projectIDNumber, ok := p["project_id"].(json.Number)
	if !ok {
		return nil, fmt.Errorf("newGitlabParams: expected project_id as json.Number instead of %T", p["project_id"])
	}
	projectID, err := projectIDNumber.Int64()
	if err != nil {
		return nil, fmt.Errorf("newGitlabParams: invalid project_id: %w", err)
	}
	dir, _ := p["dir"].(string)
	tagPrefix, _ := p["tag_prefix"].(string)
	versionDir, _ := p["version_dir"].(bool)
	return &params{
		module:     module,
		projectID:  projectID,
		dir:        util.UnifyDir(dir),
		tagPrefix:  tagPrefix,
		versionDir: versionDir,
	}, nil
}
