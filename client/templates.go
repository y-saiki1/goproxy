// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package client

import (
	_ "embed"
)

var (
	//go:embed assets/index.html
	indexTemplateContent string
)
