// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetLevel(t *testing.T) {
	levels := []string{
		"", // default level
		Level.Panic,
		Level.Fatal,
		Level.Error,
		Level.Warn,
		Level.Info,
		Level.Debug,
		Level.Trace,
	}
	for _, l := range levels {
		assert.NoError(t, SetLevel(l))
	}
	assert.EqualError(t, SetLevel("invalid"), "unknown log level \"invalid\", use one of following: panic, fatal, error, warn, info, debug, trace or empty string for default log level")
}
