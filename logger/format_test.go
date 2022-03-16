// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func assertFormatter(t *testing.T,
	expected string,
	time time.Time,
	level logrus.Level,
	typ string,
	message string,
	err error,
	with ...any,
) {
	t.Helper()
	f := logrus.Fields{
		keyType:   typ,
		keyFields: with,
	}
	if err != nil {
		f[keyErr] = err
	}
	b, err := (*formatter)(nil).Format(&logrus.Entry{
		Logger:  nil,
		Data:    f,
		Time:    time,
		Level:   level,
		Caller:  nil,
		Message: message,
		Buffer:  nil,
		Context: nil,
	})
	assert.Equal(t, expected, string(b))
	assert.NoError(t, err)
}

func Test_formatter(t *testing.T) {
	now := time.Date(2022, 9, 7, 4, 20, 45, 15, time.UTC)
	assertFormatter(t,
		"{\"timestamp\":\"2022-09-07T04:20:45.000000015Z\",\"level\":\"info\",\"type\":\"test\",\"message\":\"example of message\",\"data\":{}}\r\n",
		now,
		logrus.InfoLevel,
		"test",
		"example of message",
		nil,
	)
	assertFormatter(t,
		"{\"timestamp\":\"2022-09-07T04:20:45.000000015Z\",\"level\":\"info\",\"type\":\"test\",\"message\":\"example of message\",\"err\":\"error message\",\"data\":{\"hello\":\"world\",\"another\":\"message\"}}\r\n",
		now,
		logrus.InfoLevel,
		"test",
		"example of message",
		errors.New("error message"),
		"hello", "world",
		"another", "message",
	)
}

func Test_formatTime(t *testing.T) {
	assert.Equal(t, "2022-09-07T04:20:45.000000015Z", formatTime(time.Date(2022, 9, 7, 4, 20, 45, 15, time.UTC)))
}

func Test_formatLevel(t *testing.T) {
	assert.Equal(t, Level.Panic, formatLevel(logrus.PanicLevel))
	assert.Equal(t, Level.Fatal, formatLevel(logrus.FatalLevel))
	assert.Equal(t, Level.Error, formatLevel(logrus.ErrorLevel))
	assert.Equal(t, Level.Warn, formatLevel(logrus.WarnLevel))
	assert.Equal(t, Level.Info, formatLevel(logrus.InfoLevel))
	assert.Equal(t, Level.Debug, formatLevel(logrus.DebugLevel))
	assert.Equal(t, Level.Trace, formatLevel(logrus.TraceLevel))
	assert.Equal(t, "", formatLevel(logrus.Level(1000)))
}

func Test_formatType(t *testing.T) {
	assert.Equal(t, "name", formatType(&logrus.Entry{Data: logrus.Fields{keyType: "name"}}))
	assert.Equal(t, typeUnknown, formatType(&logrus.Entry{Data: logrus.Fields{}}))
}
