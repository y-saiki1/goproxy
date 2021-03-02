// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	Level = struct {
		Panic string
		Fatal string
		Error string
		Warn  string
		Info  string
		Debug string
		Trace string
	}{
		Panic: "panic",
		Fatal: "fatal",
		Error: "error",
		Warn:  "warn",
		Info:  "info",
		Debug: "debug",
		Trace: "trace",
	}
)

// SetLevel sets log level.
// Empty string is default log level which is trace.
func SetLevel(level string) error {
	switch level {
	case Level.Panic:
		logrus.SetLevel(logrus.PanicLevel)
		return nil
	case Level.Fatal:
		logrus.SetLevel(logrus.FatalLevel)
		return nil
	case Level.Error:
		logrus.SetLevel(logrus.ErrorLevel)
		return nil
	case Level.Warn:
		logrus.SetLevel(logrus.WarnLevel)
		return nil
	case Level.Info:
		logrus.SetLevel(logrus.InfoLevel)
		return nil
	case Level.Debug:
		logrus.SetLevel(logrus.DebugLevel)
		return nil
	case Level.Trace, "": // also accept empty string as default
		logrus.SetLevel(logrus.TraceLevel)
		return nil
	default:
		return fmt.Errorf("unknown log level %q, use one of following: panic, fatal, error, warn, info, debug, trace or empty string for default log level", level)
	}
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}
