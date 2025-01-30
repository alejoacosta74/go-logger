package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// runtimeContextHook implements logrus.Hook
type runtimeContextHook struct {
	skipFrames int // Configurable skip frames
}

// NewRuntimeContextHook creates a new hook with configurable frame skipping
func NewRuntimeContextHook(skipFrames int) *runtimeContextHook {
	return &runtimeContextHook{skipFrames: skipFrames}
}

func (h *runtimeContextHook) Levels() []logrus.Level {
	// Return ALL levels
	return []logrus.Level{
		logrus.TraceLevel, // 6
		logrus.DebugLevel, // 5
		logrus.InfoLevel,  // 4
		logrus.WarnLevel,  // 3
		logrus.ErrorLevel, // 2
		logrus.FatalLevel, // 1
		logrus.PanicLevel, // 0
	}
}

// Hook implementation
func (h *runtimeContextHook) Fire(entry *logrus.Entry) error {
	if info, ok := extractCallerInfo(h.skipFrames); ok {

		funcText := fmt.Sprintf("%s.%s", info.pkgName, info.shortFunc)
		srcText := fmt.Sprintf("%s:%d", info.fileName, info.line)

		entry.Data["func"] = funcText
		entry.Data["src"] = srcText
	}
	return nil
}
