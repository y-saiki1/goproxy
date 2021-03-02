// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_logger_fields(t *testing.T) {
	l := &logger{}
	assert.Equal(t, logrus.Fields{
		"type": "unknown",
	}, l.fields())

	l = l.Type("hello").(*logger)
	assert.Equal(t, logrus.Fields{
		"type": "hello",
	}, l.fields())

	err := errors.New("my error")
	l = l.Err(err).(*logger)
	assert.Equal(t, logrus.Fields{
		"type": "hello",
		"err":  err,
	}, l.fields())

	l = l.With("a", "b", "c", "d").(*logger)
	assert.Equal(t, logrus.Fields{
		"type": "hello",
		"err":  err,
		"fields": []interface{}{
			"a", "b",
			"c", "d",
		},
	}, l.fields())
}

func assertEqualSlice(t *testing.T, expected, actual []interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual)
	assert.Equalf(t, cap(actual), len(actual), "expected same capacity and length: %d != %d", cap(actual), len(actual))
}

func Test_appendPairs(t *testing.T) {
	assert.Nil(t, appendPairs(nil, nil))
	assert.Nil(t, appendPairs(nil, []interface{}{}))
	assert.Nil(t, appendPairs(nil, []interface{}{"c"}))
	assert.Nil(t, appendPairs([]interface{}{}, nil))
	assert.Nil(t, appendPairs([]interface{}{}, []interface{}{}))
	assert.Nil(t, appendPairs([]interface{}{"a"}, nil))
	assert.Nil(t, appendPairs([]interface{}{"a"}, []interface{}{}))

	assertEqualSlice(t,
		[]interface{}{"c", "d"},
		appendPairs(nil, []interface{}{"c", "d"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b"},
		appendPairs([]interface{}{"a", "b"}, nil),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b"},
		appendPairs([]interface{}{"a", "b"}, []interface{}{"c"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b", "c", "d"},
		appendPairs([]interface{}{"a", "b", "x", "y", "z"}[:2], []interface{}{"c", "d"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b", "c", "d"},
		appendPairs([]interface{}{"a", "b"}, []interface{}{"c", "d"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b", "c", "d"},
		appendPairs([]interface{}{"a", "b"}, []interface{}{1, 2, "c", "d"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "c"},
		appendPairs([]interface{}{"a", "b"}, []interface{}{1, 2, "a", "c"}),
	)
	assertEqualSlice(t,
		[]interface{}{"a", "b"},
		appendPairs([]interface{}{"a", "b"}, []interface{}{1, 2, "timestamp", "c", "level", "d", "type", "e", "message", "f", "err", "g"}),
	)

	// test append overlapping
	a := []interface{}{"a", "b", "x", "y"}[:2]
	b := appendPairs(a, []interface{}{"c", "d"})
	c := appendPairs(a, []interface{}{"e", "f"})
	d := appendPairs(a, []interface{}{"a", "c"})
	assertEqualSlice(t, []interface{}{"a", "b", "x", "y"}, a[:4])
	assertEqualSlice(t, []interface{}{"a", "b", "c", "d"}, b)
	assertEqualSlice(t, []interface{}{"a", "b", "e", "f"}, c)
	assertEqualSlice(t, []interface{}{"a", "c"}, d)
}

func Test_cleanPairs(t *testing.T) {
	assertEqualSlice(t, []interface{}(nil), cleanPairs([]interface{}{}))
	assertEqualSlice(t, []interface{}(nil), cleanPairs([]interface{}{"a"}))
	assertEqualSlice(t, []interface{}{"a", "b"}, cleanPairs([]interface{}{"a", "b"}))
	a := make([]interface{}, 2, 4)
	a[0] = "a"
	a[1] = "b"
	assertEqualSlice(t, []interface{}{"a", "b"}, cleanPairs(a))
	a = append(a, "c")
	assertEqualSlice(t, []interface{}{"a", "b"}, cleanPairs(a))
}

func Test_ContextWith(t *testing.T) {
	assert.Nil(t, ContextWith(nil))
	assert.Equal(t, context.WithValue(context.Background(), ctxKey, []interface{}{"a", "b"}), ContextWith(context.Background(), "a", "b"))
}

func Test_Default(t *testing.T) {
	assert.Equal(t, def, Default())
}

func Test_Type(t *testing.T) {
	assert.Equal(t, &logger{typ: "name"}, Type("name"))
}

func Test_Panic(t *testing.T) {
	assert.Panics(t, func() {
		Default().Panic("yes, this calls panic")
	})
}

type loggerTester struct {
	bytes.Buffer
}

var (
	removeTimestampPattern = regexp.MustCompile(`{"timestamp":"([^"]*)".*`)
	removeStackPattern     = regexp.MustCompile(`{.*"stack":"([^"]*)".*`)
)

func (l *loggerTester) Assert(t *testing.T, expected string) {
	t.Helper()
	s := l.String()
	if loc := removeTimestampPattern.FindStringSubmatchIndex(s); len(loc) > 0 {
		s = s[:loc[2]] + "<NOW>" + s[loc[3]:]
	}
	if loc := removeStackPattern.FindStringSubmatchIndex(s); len(loc) > 0 {
		s = s[:loc[2]] + "<STACK>" + s[loc[3]:]
	}
	assert.Equal(t, expected, s)
	l.Reset()
}

type errorCloser struct{}

func (_ *errorCloser) Close() error {
	return errors.New("close")
}

func Test_logger(t *testing.T) {
	output := &loggerTester{}
	logrus.SetOutput(output)
	logrus.StandardLogger().ExitFunc = func(_ int) {}
	log := Type("name")

	log.Ctx(context.Background()).Info("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"info\",\"type\":\"name\",\"message\":\"message\",\"data\":{}}\r\n")

	func() {
		defer func() {
			recover()
		}()
		log.Panic("message")
	}()
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"panic\",\"type\":\"name\",\"message\":\"message\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.Fatal("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"fatal\",\"type\":\"name\",\"message\":\"message\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.Error("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"error\",\"type\":\"name\",\"message\":\"message\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.Warn("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"warn\",\"type\":\"name\",\"message\":\"message\",\"data\":{}}\r\n")

	log.Info("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"info\",\"type\":\"name\",\"message\":\"message\",\"data\":{}}\r\n")

	log.Debug("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"debug\",\"type\":\"name\",\"message\":\"message\",\"data\":{}}\r\n")

	log.Trace("message")
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"trace\",\"type\":\"name\",\"message\":\"message\",\"data\":{}}\r\n")

	log.NoErr(errors.New("error"))
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"error\",\"type\":\"name\",\"message\":\"unexpected error\",\"err\":\"error\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	assert.Panics(t, func() {
		log.NoErrLast()
	})
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"panic\",\"type\":\"name\",\"message\":\"invalid NoErrLast arguments: no arguments\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.NoErrLast(nil)
	output.Assert(t, "")

	log.NoErrLast(1, nil)
	output.Assert(t, "")

	assert.Panics(t, func() {
		log.NoErrLast(1)
	})
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"panic\",\"type\":\"name\",\"message\":\"invalid NoErrLast arguments: last must be error\",\"data\":{\"argument_type\":\"int\",\"stack\":\"<STACK>\"}}\r\n")

	log.NoErrLast(errors.New("error"))
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"error\",\"type\":\"name\",\"message\":\"unexpected error\",\"err\":\"error\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.NoErrLast(1, errors.New("error"))
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"error\",\"type\":\"name\",\"message\":\"unexpected error\",\"err\":\"error\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")

	log.NoErrClose((*errorCloser)(nil))
	output.Assert(t, "{\"timestamp\":\"<NOW>\",\"level\":\"error\",\"type\":\"name\",\"message\":\"unexpected error\",\"err\":\"close\",\"data\":{\"stack\":\"<STACK>\"}}\r\n")
}
