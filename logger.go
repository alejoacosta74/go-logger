package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type loggerWrapper struct {
	*logrus.Entry
}

var (
	// Logger is the global logger instance
	Logger     *loggerWrapper
	loggerOnce sync.Once
)

func init() {
	Logger = &loggerWrapper{
		Entry: logrus.NewEntry(logrus.StandardLogger()),
	}
}

func NewLogger(opts ...Option) (*loggerWrapper, error) {
	logger, err := createNewLogger(opts...)
	if err != nil {
		return nil, err
	}
	Logger = logger
	return logger, nil
}

func NewSingletonLogger(opts ...Option) (*loggerWrapper, error) {
	var err error
	loggerOnce.Do(func() {
		var logger *loggerWrapper
		logger, err = createNewLogger(opts...)
		if err == nil {
			Logger = logger
		}
	})
	return Logger, err
}

func createNewLogger(opts ...Option) (*loggerWrapper, error) {
	l := logrus.New()
	formatter := &logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	}
	l.SetFormatter(formatter)

	logger := &loggerWrapper{
		Entry: logrus.NewEntry(l),
	}

	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}
	return logger, nil
}

// package level functions
func Trace(args ...interface{}) {
	Logger.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	Logger.Tracef(format, args...)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	Logger.Panicf(format, args...)
}

func WithField(key string, value interface{}) *loggerWrapper {
	return &loggerWrapper{Entry: Logger.WithField(key, value)}
}

func WithFields(fields ...string) *loggerWrapper {
	if len(fields)%2 != 0 {
		panic("WithFields requires an even number of arguments")
	}

	f := make(logrus.Fields)
	for i := 0; i < len(fields); i += 2 {
		f[fields[i]] = fields[i+1]
	}
	return &loggerWrapper{Entry: Logger.WithFields(f)}
}

func SetLevel(level string) {
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	Logger.Entry.Logger.SetLevel(parsedLevel)
}

// ResetLogger resets the singleton instance (for testing)
func ResetLogger() {
	Logger = nil
	loggerOnce = sync.Once{}
}
