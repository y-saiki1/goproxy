// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go.lstv.dev/goproxy/logger"
	"go.lstv.dev/goproxy/service"
	_ "go.lstv.dev/goproxy/source/gitlab"
)

func main() {
	if len(os.Args) < 2 {
		_, file := filepath.Split(os.Args[0])
		fmt.Println("Usage:", file, "<config>")
		return
	}
	c, err := service.LoadConfig(os.Args[1])
	if err != nil {
		logger.Type("main").NoErrLast(fmt.Fprintln(os.Stderr, err))
		os.Exit(1)
	}
	logger.Type("main").NoErr(logger.SetLevel(c.LogLevel))
	p, err := service.NewGoProxy(c)
	if err != nil {
		logger.Type("main").NoErrLast(fmt.Fprintln(os.Stderr, err))
		os.Exit(1)
	}
	if err := p.Start(); err != nil {
		logger.Type("main").Err(err).Error("failed")
	}
}
