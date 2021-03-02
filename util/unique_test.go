// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateUniqueID(t *testing.T) {
	assert.Len(t, GenerateUniqueID(), 16)
}
