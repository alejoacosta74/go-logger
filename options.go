package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Option func(*loggerWrapper) error

// WithLevel sets the logging level (trace, debug, info, warn, error, fatal, panic)
func WithLevel(level string) Option {
	return func(l *loggerWrapper) error {
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

// runtimeContextHook implements logrus.Hook
type runtimeContextHook struct {
	skipFrames int // Configurable skip frames
}

// NewRuntimeContextHook creates a new hook with configurable frame skipping
func NewRuntimeContextHook(skipFrames int) *runtimeContextHook {
	return &runtimeContextHook{skipFrames: skipFrames}
}

func (h *runtimeContextHook) Levels() []logrus.Level {
	// Return ALL levels up to Debug (inclusive)
	return []logrus.Level{
		logrus.TraceLevel, // 0
		logrus.DebugLevel, // 1
	}
}

// callerInfo holds the extracted runtime caller information
type callerInfo struct {
	funcName  string
	fileName  string
	line      int
	pkgName   string
	shortFunc string
}

// extractCallerInfo without anonymous function filtering
func extractCallerInfo(skipFrames int) (callerInfo, bool) {
	var info callerInfo
	for i := skipFrames; i < skipFrames+15; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok {
			funcName := runtime.FuncForPC(pc).Name()
			if !strings.Contains(funcName, "logrus") &&
				!strings.Contains(funcName, "runtime.") &&
				!strings.Contains(funcName, "testing.") &&
				!strings.Contains(file, "runtime/") &&
				!strings.Contains(file, "testing/") &&
				!strings.Contains(funcName, "WithRuntimeContext") {

				info.funcName = funcName
				info.fileName = file
				info.line = line

				lastDot := strings.LastIndex(funcName, ".")
				if lastDot != -1 {
					pkgPath := funcName[:lastDot]
					fullFunc := funcName[lastDot+1:]
					pkgParts := strings.Split(pkgPath, "/")
					info.pkgName = pkgParts[len(pkgParts)-1]
					info.shortFunc = fullFunc
					return info, true
				}
			}
		}
	}
	return info, false
}

// Hook implementation
func (h *runtimeContextHook) Fire(entry *logrus.Entry) error {
	if info, ok := extractCallerInfo(h.skipFrames); ok {
		_, fileName := filepath.Split(info.fileName)

		// Force quotes by adding them explicitly
		entry.Data["func"] = fmt.Sprintf("%s.%s", info.pkgName, info.shortFunc)
		entry.Data["src"] = fmt.Sprintf("%s:%d", fileName, info.line)
	}
	return nil
}

// WithRuntimeContext implementation
func WithRuntimeContext() Option {
	return func(l *loggerWrapper) error {
		formatter := &logrus.TextFormatter{
			TimestampFormat:        time.RFC3339,
			FullTimestamp:          true,
			DisableLevelTruncation: true,
			ForceColors:            true,
			PadLevelText:           false,
			DisableColors:          false,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				if info, ok := extractCallerInfo(8); ok {
					_, fileName := filepath.Split(info.fileName)
					formattedFunc := fmt.Sprintf("func: %s.%s -", info.pkgName, info.shortFunc)

					return formattedFunc, fmt.Sprintf(" - src: %s:%d", fileName, info.line)
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
	return func(l *loggerWrapper) error {
		l.Entry.Logger.SetOutput(output)
		return nil
	}
}

// WithNullOutput sets the output destination to io.Discard
func WithNullOutput() Option {
	return func(l *loggerWrapper) error {
		l.Entry.Logger.SetOutput(io.Discard)
		return nil
	}
}

// WithFileOutput sets the output destination to a file
func WithFileOutput(file string) Option {
	return func(l *loggerWrapper) error {
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

	return func(l *loggerWrapper) error {
		f := make(logrus.Fields)
		for i := 0; i < len(fields); i += 2 {
			f[fields[i]] = fields[i+1]
		}
		l.Entry = l.Entry.WithFields(f)
		return nil
	}
}
