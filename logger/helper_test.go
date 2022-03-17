// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dummyLogTextFormatter struct{}

func (_ dummyLogTextFormatter) MarshalText() ([]byte, error) {
	return []byte("text"), nil
}

func (_ dummyLogTextFormatter) String() string {
	return "string"
}

type dummyLogJSONFormatter struct{}

func (_ dummyLogJSONFormatter) MarshalText() ([]byte, error) {
	return []byte("text"), nil
}

func (_ dummyLogJSONFormatter) MarshalJSON() ([]byte, error) {
	return []byte(`"json"`), nil
}

func (_ dummyLogJSONFormatter) String() string {
	return "string"
}

type utilDate struct {
	year  int
	month int
	day   int
}

func (d utilDate) Date() (year int, month time.Month, day int) {
	return d.year + 1, time.Month(d.month + 1), d.day + 1
}

func (d utilDate) LogFormat() string {
	return d.String()
}

func (d utilDate) String() string {
	year, month, day := d.Date()
	return fmt.Sprintf(`%04d-%02d-%02d`, year, month, day)
}

type panicMarshaller struct{}

func (_ *panicMarshaller) MarshalJSON() ([]byte, error) {
	panic("panicMarshaller")
}

func Test_toJSON(t *testing.T) {
	vbool := true
	pbool := &vbool
	vint := 7
	pint := &vint
	vstring := "x"
	pstring := &vstring
	vstruct := struct {
		X int
	}{
		X: 2,
	}
	pstruct := &vstruct
	varray := [2]int{2, 7}
	parray := &varray
	vslice := []int{2, 7}
	pslice := &vslice
	vmap := map[string]any{
		"x": 2,
		"y": struct {
			Z int `json:"z"`
		}{
			Z: 7,
		},
	}
	pmap := vmap
	verr := errors.New("err")
	perr := &verr
	pnilint := (*int)(nil)
	var date *utilDate
	assert.Equal(t, json.RawMessage(`null`), toJSON(nil))
	assert.Equal(t, json.RawMessage(`null`), toJSON(pnilint))
	assert.Equal(t, json.RawMessage(`null`), toJSON(&pnilint))
	assert.Equal(t, json.RawMessage(`true`), toJSON(vbool))
	assert.Equal(t, json.RawMessage(`true`), toJSON(pbool))
	assert.Equal(t, json.RawMessage(`true`), toJSON(&pbool))
	assert.Equal(t, json.RawMessage(`7`), toJSON(vint))
	assert.Equal(t, json.RawMessage(`7`), toJSON(pint))
	assert.Equal(t, json.RawMessage(`7`), toJSON(&pint))
	assert.Equal(t, json.RawMessage(`"x"`), toJSON(vstring))
	assert.Equal(t, json.RawMessage(`"x"`), toJSON(pstring))
	assert.Equal(t, json.RawMessage(`"x"`), toJSON(&pstring))
	assert.Equal(t, json.RawMessage(`{"X":2}`), toJSON(vstruct))
	assert.Equal(t, json.RawMessage(`{"X":2}`), toJSON(pstruct))
	assert.Equal(t, json.RawMessage(`{"X":2}`), toJSON(&pstruct))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(varray))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(parray))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(&parray))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(vslice))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(pslice))
	assert.Equal(t, json.RawMessage(`[2,7]`), toJSON(&pslice))
	assert.Equal(t, json.RawMessage(`{"x":2,"y":{"z":7}}`), toJSON(vmap))
	assert.Equal(t, json.RawMessage(`{"x":2,"y":{"z":7}}`), toJSON(pmap))
	assert.Equal(t, json.RawMessage(`{"x":2,"y":{"z":7}}`), toJSON(&pmap))
	assert.Equal(t, json.RawMessage(`"err"`), toJSON(verr))
	assert.Equal(t, json.RawMessage(`"err"`), toJSON(perr))
	assert.Equal(t, json.RawMessage(`"err"`), toJSON(&perr))
	assert.Equal(t, json.RawMessage(`"text"`), toJSON(dummyLogTextFormatter{}))
	assert.Equal(t, json.RawMessage(`"text"`), toJSON(&dummyLogTextFormatter{}))
	assert.Equal(t, json.RawMessage(`"json"`), toJSON(dummyLogJSONFormatter{}))
	assert.Equal(t, json.RawMessage(`"json"`), toJSON(&dummyLogJSONFormatter{}))
	assert.Equal(t, json.RawMessage(`null`), toJSON(date))
	assert.Equal(t, json.RawMessage(`"%!panic(panicMarshaller)"`), toJSON(&panicMarshaller{}))
}
