// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jsonUnmarshalWithNumbers(t *testing.T) {
	b := strings.NewReader(`{"number":1.15}`)
	m := map[string]interface{}{}
	assert.NoError(t, jsonUnmarshalWithNumbers(b, &m))
	assert.Equal(t, map[string]interface{}{
		"number": json.Number("1.15"),
	}, m)
}
