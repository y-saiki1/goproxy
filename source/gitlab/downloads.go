// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.lstv.dev/goproxy/util"
)

type downloads struct {
	*Source
	name                string
	projectID           int64
	packageName         string
	disableArchitecture bool
	fileExtension       string
}

func (d *downloads) ConfigPreview() (pairs []string) {
	return []string{
		"type", "gitlab",
		"url", d.url,
		"project_id", strconv.FormatInt(d.projectID, 10),
		"package_name", d.packageName,
		"insecure_tls", strconv.FormatBool(d.insecureTLS),
	}
}

func (d *downloads) WriteDownload(ctx context.Context, w http.ResponseWriter, v util.Version, arch string) {
	log := d.log.Ctx(ctx)
	url := d.apiURL(fmt.Sprintf("projects/%[1]d/packages/generic/%[2]s/%[3]s/%[2]s-%[3]s%[4]s",
		d.projectID,
		d.packageName,
		v,
		d.extension(arch),
	))
	resp, err := d.doGetRequest(ctx, url)
	if err != nil {
		log.Err(err).Warn("download request failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer log.NoErrClose(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.With(
			"status_code", resp.StatusCode,
		).Warn("download request failed: unexpected status code")
		w.WriteHeader(resp.StatusCode)
		return
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Err(err).Warn("download request failed")
	}
}

func (d *downloads) LatestDownloadVersion(ctx context.Context) (latest util.Version, err error) {
	latest = util.Version{}
	for page := 1; ; page++ {
		v, hasNextPage, err := d.latestDownloadVersionPage(ctx, page)
		if err != nil {
			return util.Version{}, err
		}
		latest = v.Latest(latest)
		if !hasNextPage {
			return latest, nil
		}
	}
}

func (d *downloads) latestDownloadVersionPage(ctx context.Context, page int) (latest util.Version, hasNextPage bool, err error) {
	log := d.log.Ctx(ctx)
	// TODO https://gitlab.com/gitlab-org/gitlab/-/issues/290007
	// Use "projects/%d/packages?page=1&per_page=1&order_by=version&sort=desc&package_type=generic&package_name=%s" after fix.
	url := d.apiURL(fmt.Sprintf("projects/%d/packages?page=%d&package_type=generic&package_name=%s",
		d.projectID,
		page,
		d.name,
	))
	resp, err := d.doGetRequest(ctx, url)
	if err != nil {
		log.Err(err).Warn("fetch latest download version failed")
		return util.Version{}, false, err
	}
	defer log.NoErrClose(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.With(
			"status_code", resp.StatusCode,
		).Warn("fetch latest download version failed: unexpected status code")
		return util.Version{}, false, err
	}
	result := ([]struct {
		Version util.Version `json:"version"`
	})(nil)
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Err(err).Warn("parse latest download version failed")
		return util.Version{}, false, err
	}
	if len(result) == 0 {
		err := errors.New("no latest download version")
		log.Err(err).Warn("fetch latest download version failed")
		return util.Version{}, false, err
	}
	latest = util.Version{}
	for _, r := range result {
		latest = r.Version.Latest(latest)
	}
	nextPage, err := strconv.ParseInt(resp.Header.Get("x-next-page"), 10, 64)
	if err != nil {
		// if x-next-page is not valid, there is no next page
		return latest, false, nil
	}
	return latest, int(nextPage) != page, nil
}

func (d *downloads) extension(arch string) string {
	if d.disableArchitecture {
		return d.fileExtension
	}
	return "-" + arch + d.fileExtension
}
