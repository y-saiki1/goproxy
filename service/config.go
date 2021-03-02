// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/util"
)

type Config struct {
	Addr              string                    `json:"addr"`
	Storage           string                    `json:"storage"`
	LogLevel          string                    `json:"log_level"`
	Modules           []ModuleConfig            `json:"modules"`
	Downloads         map[string]DownloadConfig `json:"downloads"`
	Sources           []map[string]interface{}  `json:"sources"`
	Versions          VersionsConfig            `json:"versions"`
	DefaultGoProxyURL string                    `json:"default_go_proxy_url"`
	DownloadsPrefix   string                    `json:"downloads_prefix"`
}

type ModuleConfig struct {
	Name         string                 `json:"name"`
	Source       *string                `json:"source"`
	SourceParams map[string]interface{} `json:"source_params"`
}

type DownloadConfig struct {
	Mode         string                 `json:"mode"`
	Source       string                 `json:"source"`
	SourceParams map[string]interface{} `json:"source_params"`
}

type VersionsConfig struct {
	Go      util.Version `json:"go"`
	Modules []string     `json:"modules"`
}

func LoadConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}
	defer logger.Type("service.LoadConfig").NoErrClose(f)

	c := &Config{}
	if err := jsonUnmarshalWithNumbers(f, c); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}
	return c, nil
}

func jsonUnmarshalWithNumbers(r io.Reader, v interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}
