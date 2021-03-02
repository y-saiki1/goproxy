// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

var ignoredKeys = map[string]struct{}{
	"timestamp": {},
	"level":     {},
	"type":      {},
	"message":   {},
	"err":       {},
}

func IgnoredKey(key string) bool {
	_, ok := ignoredKeys[key]
	return ok
}

type Logger interface {
	Type(typ string) Logger
	Ctx(ctx context.Context) Logger
	Err(err error) Logger

	With(pairs ...interface{}) Logger

	Fatal(message string)
	Panic(message string)
	Error(message string)
	Warn(message string)
	Info(message string)
	Debug(message string)
	Trace(message string)

	NoErr(err error)
	NoErrLast(lastErr ...interface{})
	NoErrClose(closer io.Closer)
}

type logger struct {
	typ   string
	err   error
	pairs []interface{}
}

// Type returns new logger with specified type name.
// Type should be "package.TypeName", "package.FuncName" etc.
func (l *logger) Type(typ string) Logger {
	return &logger{
		typ:   typ,
		err:   l.err,
		pairs: l.pairs,
	}
}

// Ctx returns new logger with values from passed context.
func (l *logger) Ctx(ctx context.Context) Logger {
	return l.With(ctxFields(ctx)...)
}

// Err returns new logger with specified error.
func (l *logger) Err(err error) Logger {
	return &logger{
		typ:   l.typ,
		err:   err,
		pairs: l.pairs,
	}
}

// With returns new logger with passed pairs.
// Pairs must be key as string and value as JSON-compatible value.
// If key is not string, whole pair is ignored.
// If key already exists at parent logger, returned logger overwrite this value.
// If count of passed values is odd, last value is ignored.
//
// Ignored keys are also: timestamp, level, type, message, err.
// Key stack will be replaced with stack for error, fatal and panic log level messages.
//
// Format With calls as following:
//
//   logger.With(
//     "number", 7,
//     "slice", []int{7, 8},
//   )
func (l *logger) With(pairs ...interface{}) Logger {
	return &logger{
		typ:   l.typ,
		err:   l.err,
		pairs: appendPairs(l.pairs, pairs),
	}
}

func (l *logger) Fatal(message string) {
	l.withStack().logrus().Fatal(message)
}

func (l *logger) Panic(message string) {
	l.withStack().logrus().Panic(message)
}

func (l *logger) Error(message string) {
	l.withStack().logrus().Error(message)
}

func (l *logger) Warn(message string) {
	l.logrus().Warn(message)
}

func (l *logger) Info(message string) {
	l.logrus().Info(message)
}

func (l *logger) Debug(message string) {
	l.logrus().Debug(message)
}

func (l *logger) Trace(message string) {
	l.logrus().Trace(message)
}

// NoErr logs error is passed value is not nil.
func (l *logger) NoErr(err error) {
	if err != nil {
		l.Err(err).Error("unexpected error")
	}
}

// NoErrLast checks if last passed argument is not nil value of error type and pass this value to NoErr.
// If last passed argument is nil, nothing happen.
// Otherwise, Panic is called.
func (l *logger) NoErrLast(lastErr ...interface{}) {
	lenLastErr := len(lastErr)
	if lenLastErr == 0 {
		l.Panic("invalid NoErrLast arguments: no arguments")
	}
	value := lastErr[lenLastErr-1]
	if value == nil {
		return
	}
	if err, ok := value.(error); ok {
		l.NoErr(err)
		return
	}
	l.With(
		"argument_type", fmt.Sprintf("%T", value),
	).Panic("invalid NoErrLast arguments: last must be error")
}

func (l *logger) NoErrClose(closer io.Closer) {
	if closer != nil {
		l.NoErr(closer.Close())
	}
}

func (l *logger) withStack() *logger {
	return &logger{
		typ: l.typ,
		err: l.err,
		pairs: appendPairs(l.pairs, []interface{}{
			keyStack, string(debug.Stack()),
		}),
	}
}

func (l *logger) logrus() *logrus.Entry {
	return logrus.WithFields(l.fields())
}

func (l *logger) fields() logrus.Fields {
	fields := logrus.Fields{}
	if l.typ == "" {
		fields[keyType] = typeUnknown
	} else {
		fields[keyType] = l.typ
	}
	if l.err != nil {
		fields[keyErr] = l.err
	}
	if len(l.pairs) > 1 {
		fields[keyFields] = l.pairs
	}
	return fields
}

func appendPairs(a, b []interface{}) []interface{} {
	ca := cleanPairs(a)
	cb := cleanPairs(b)
	if cb == nil {
		return ca
	}
	lca := len(ca)
	lcb := len(cb)
	r := make([]interface{}, lca, lca+lcb)
	copy(r, ca)
newPairs:
	for i := 0; i < lcb; i += 2 {
		key, ok := cb[i].(string)
		if !ok || IgnoredKey(key) {
			continue
		}
		value := cb[i+1]
		for j := 0; j < lca; j += 2 {
			if r[j] == key {
				r[j+1] = value
				continue newPairs
			}
		}
		r = append(r, key, value)
	}
	return cleanPairs(r)
}

func cleanPairs(a []interface{}) []interface{} {
	l := len(a)
	if l < 2 {
		return nil
	}
	if l%2 == 1 {
		return a[: l-1 : l-1]
	}
	return a[:l:l]
}

type ctxKeyType struct{}

var ctxKey = ctxKeyType{}

func ctxFields(ctx context.Context) []interface{} {
	pairs, _ := ctx.Value(ctxKey).([]interface{})
	return pairs
}

func ContextWith(ctx context.Context, pairs ...interface{}) context.Context {
	if ctx == nil {
		return nil
	}
	fields := ctxFields(ctx)
	fields = appendPairs(fields, pairs)
	return context.WithValue(ctx, ctxKey, fields)
}

var def = &logger{}

// Default returns default logger.
func Default() Logger {
	return def
}

// Type calls Type on default logger.
func Type(typ string) Logger {
	return def.Type(typ)
}
