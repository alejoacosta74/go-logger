package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultMaxSize    = 100 // 100MB
	DefaultMaxBackups = 3
	DefaultMaxAge     = 28 // 28 days
)

// RotatingFileHook implements logrus.Hook interface with log rotation support
type rotatingFileHook struct {
	config    *lumberjack.Logger
	formatter logrus.Formatter
	levels    []logrus.Level
	mu        sync.Mutex
}

// RotatingFileConfig holds configuration for log rotation
type RotatingFileConfig struct {
	Filename   string
	MaxSize    int  // megabytes
	MaxBackups int  // number of backups
	MaxAge     int  // days
	Compress   bool // compress rotated files
	Levels     []logrus.Level
}

// NewRotatingFileHook creates a new hook with log rotation support
func newRotatingFileHook(cfg *RotatingFileConfig) (*rotatingFileHook, error) {
	if cfg == nil {
		cfg = &RotatingFileConfig{}
	}
	if cfg.Filename == "" {
		cfg.Filename = "logs/app.log"
	}

	// Set default values if not specified
	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultMaxSize
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = DefaultMaxBackups
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = DefaultMaxAge
	}
	if len(cfg.Levels) == 0 {
		cfg.Levels = logrus.AllLevels
	}

	hook := &rotatingFileHook{
		config: &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		},
		formatter: &logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		},
		levels: cfg.Levels,
	}

	return hook, nil
}

// Fire writes the log entry to the file
func (h *rotatingFileHook) Fire(entry *logrus.Entry) error {
	// First check if entry.Level is within the threshold levels
	if !h.shouldLog(entry.Level) {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.config.Write(line)
	return err
}

// shouldLog checks if the log level should be logged
func (h *rotatingFileHook) shouldLog(level logrus.Level) bool {
	if len(h.levels) == 0 {
		return true
	}

	minLevel := logrus.PanicLevel
	for _, l := range h.levels {
		if l < minLevel {
			minLevel = l
		}
	}

	return level >= minLevel
}

// Levels returns the levels this hook should be fired for
func (h *rotatingFileHook) Levels() []logrus.Level {
	return h.levels
}

// Close implements io.Closer
func (h *rotatingFileHook) Close() error {
	return h.config.Close()
}
