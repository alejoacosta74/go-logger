package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type Option func(*Logger) error

// WithFormatter sets a custom formatter for the logger
func WithFormatter(formatter logrus.Formatter) Option {
	return func(l *Logger) error {
		l.Entry.Logger.SetFormatter(formatter)
		return nil
	}
}

// WithLevel sets the logging level (trace, debug, info, warn, error, fatal, panic)
func WithLevel(level string) Option {
	return func(l *Logger) error {
		parsedLevel, err := logrus.ParseLevel(level)
		if err != nil {
			return err
		}
		l.Entry.Logger.SetLevel(parsedLevel)

		// Add hook for debug OR trace level
		if parsedLevel == logrus.DebugLevel || parsedLevel == logrus.TraceLevel {
			l.Entry.Logger.AddHook(NewRuntimeContextHook(3))
		}
		return nil
	}
}

// WithRuntimeContext implementation
func WithRuntimeContext() Option {
	return func(l *Logger) error {
		formatter := &logrus.TextFormatter{
			TimestampFormat:        time.RFC3339,
			FullTimestamp:          true,
			DisableLevelTruncation: true,
			ForceColors:            true,
			PadLevelText:           false,
			DisableColors:          false,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				if info, ok := extractCallerInfo(8); ok {
					formattedFunc := fmt.Sprintf("func: %s.%s -", info.pkgName, info.shortFunc)

					return formattedFunc, fmt.Sprintf(" - src: %s:%d", info.fileName, info.line)
				}
				return "", ""
			},
		}

		// Preserve existing fields when setting the formatter
		fields := l.Entry.Data
		l.Entry.Logger.SetFormatter(formatter)
		l.Entry.Logger.SetReportCaller(true)
		if len(fields) > 0 {
			l.Entry = l.Entry.WithFields(fields)
		}
		return nil
	}
}

// WithOutput sets the output destination for the logger
func WithOutput(output io.Writer) Option {
	return func(l *Logger) error {
		l.Entry.Logger.SetOutput(output)
		return nil
	}
}

// WithNullOutput sets the output destination to io.Discard
func WithNullOutput() Option {
	return func(l *Logger) error {
		l.Entry.Logger.SetOutput(io.Discard)
		return nil
	}
}

// WithFileOutput sets the output destination to a file
func WithFileOutput(file string) Option {
	return func(l *Logger) error {
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		l.Entry.Logger.SetOutput(f)
		return nil
	}
}

// WithMultipleFields adds fields to the log entry
func WithMultipleFields(fields ...string) Option {
	if len(fields)%2 != 0 {
		panic("With MultipleFields requires an even number of arguments")
	}

	return func(l *Logger) error {
		f := make(logrus.Fields)
		for i := 0; i < len(fields); i += 2 {
			f[fields[i]] = fields[i+1]
		}
		l.Entry = l.Entry.WithFields(f)
		return nil
	}
}
