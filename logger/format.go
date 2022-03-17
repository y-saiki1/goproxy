// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type formatter struct{}

func (_ *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	t := formatType(entry)
	b := bytes.Buffer{}
	b.WriteString("{\"timestamp\":\"")
	b.WriteString(formatTime(entry.Time))
	b.WriteString("\",\"level\":\"")
	b.WriteString(formatLevel(entry.Level))
	b.WriteString("\",\"type\":")
	b.Write(stringToJSON(t))
	b.WriteString(",\"message\":")
	b.Write(stringToJSON(entry.Message))
	if err, ok := entry.Data[keyErr]; ok {
		b.WriteString(",\"err\":")
		b.Write(toJSON(err))
	}
	b.WriteString(",\"data\":{")
	if fields, ok := entry.Data[keyFields].([]any); ok {
		if l := len(fields); l > 1 {
			for i := 0; i < l; i += 2 {
				k := fields[i]
				v := fields[i+1]
				if i != 0 {
					b.WriteByte(',')
				}
				b.Write(stringToJSON(fmt.Sprintf("%s", k)))
				b.WriteByte(':')
				b.Write(toJSON(v))
			}
		}
	}
	b.WriteString("}}\r\n")
	return b.Bytes(), nil
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func formatLevel(l logrus.Level) string {
	switch l {
	case logrus.PanicLevel:
		return Level.Panic
	case logrus.FatalLevel:
		return Level.Fatal
	case logrus.ErrorLevel:
		return Level.Error
	case logrus.WarnLevel:
		return Level.Warn
	case logrus.InfoLevel:
		return Level.Info
	case logrus.DebugLevel:
		return Level.Debug
	case logrus.TraceLevel:
		return Level.Trace
	default:
		// unreachable
		return ""
	}
}

func formatType(entry *logrus.Entry) string {
	if s, ok := entry.Data[keyType].(string); ok {
		return s
	}
	return typeUnknown
}

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter((*formatter)(nil))
}
