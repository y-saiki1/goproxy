// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	typeUnknown = "unknown"

	keyType   = "type"
	keyErr    = "err"
	keyFields = "fields"
	keyStack  = "stack"
)

var (
	null     = json.RawMessage("null")
	jsonKind = map[reflect.Kind]struct{}{
		reflect.Bool:       {},
		reflect.Int:        {},
		reflect.Int8:       {},
		reflect.Int16:      {},
		reflect.Int32:      {},
		reflect.Int64:      {},
		reflect.Uint:       {},
		reflect.Uint8:      {},
		reflect.Uint16:     {},
		reflect.Uint32:     {},
		reflect.Uint64:     {},
		reflect.Float32:    {},
		reflect.Float64:    {},
		reflect.Complex64:  {},
		reflect.Complex128: {},
		reflect.Array:      {},
		reflect.Map:        {},
		reflect.Slice:      {},
		reflect.String:     {},
		reflect.Struct:     {},
		// Following kinds are not converted to JSON:
		// - reflect.Invalid
		// - reflect.Uintptr
		// - reflect.Chan
		// - reflect.Func
		// - reflect.Interface
		// - reflect.Ptr
		// - reflect.UnsafePointer
	}
)

func toJSON(value any) (jsonValue json.RawMessage) {
	defer func() {
		if r := recover(); r != nil {
			jsonValue = stringToJSON(fmt.Sprintf("%%!panic(%s)", r))
		}
	}()
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Invalid {
		return null
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return null
		}
		v = reflect.Indirect(v)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return null
			}
			v = reflect.Indirect(v)
		}
	}

	if err, ok := value.(error); ok {
		return stringToJSON(err.Error())
	}
	if _, ok := jsonKind[v.Kind()]; ok {
		if b, err := json.Marshal(value); err == nil {
			return b
		}
	}
	return stringToJSON(fmt.Sprintf("%v", v.Interface()))
}

func stringToJSON(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}
