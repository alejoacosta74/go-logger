package logger

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type loggerWrapper struct {
	*logrus.Entry
}

var (
	// Logger is the global logger instance
	Logger     *loggerWrapper
	loggerOnce sync.Once
	// customFormatter is the custom formatter for the logger
	customFormatter *ColorFormatter = &ColorFormatter{
		TextFormatter: logrus.TextFormatter{
			DisableTimestamp:          false,
			ForceColors:               true,  // Force colors even if not TTY
			EnvironmentOverrideColors: true,  // Allow environment to override color settings
			DisableColors:             false, // Ensure colors aren't disabled
		},
	}
)

func init() {
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry.Logger.SetFormatter(customFormatter)
	Logger = &loggerWrapper{
		Entry: entry,
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
	l.SetFormatter(customFormatter)

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
	if parsedLevel == logrus.DebugLevel || parsedLevel == logrus.TraceLevel {
		// add the runtime context hook
		Logger.Entry.Logger.AddHook(NewRuntimeContextHook(3))
	}
}

// ResetLogger resets the singleton instance (for testing)
func ResetLogger() {
	Logger = nil
	loggerOnce = sync.Once{}
}

type ColorFormatter struct {
	logrus.TextFormatter
}

func (f *ColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	var b bytes.Buffer

	// Colors for log components
	timestampColor := color.New(color.FgCyan)
	levelColor := color.New(color.FgWhite)
	messageColor := color.New(color.Reset) // Default terminal color

	// Determine level color
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = color.New(color.FgBlue)
	case logrus.InfoLevel:
		levelColor = color.New(color.FgGreen)
	case logrus.WarnLevel:
		levelColor = color.New(color.FgYellow)
	case logrus.ErrorLevel:
		levelColor = color.New(color.FgRed)
	case logrus.FatalLevel, logrus.PanicLevel:
		levelColor = color.New(color.BgRed, color.FgWhite)
	}

	// Format timestamp, level, and message
	timestamp := timestampColor.Sprint(entry.Time.Format(time.StampMilli))
	level := levelColor.Sprintf("[%s]", entry.Level.String())
	message := messageColor.Sprintf(entry.Message)

	// Write main log line
	b.WriteString(fmt.Sprintf("%s %s %s", timestamp, level, message))

	for key, value := range entry.Data {
		if key == "func" || key == "src" {
			fieldColor := color.New(color.FgCyan)
			fieldKey := fieldColor.Sprint(key)
			fieldValue := fmt.Sprintf("%s", value)
			b.WriteString(fmt.Sprintf("\t%s: %s", fieldKey, fieldValue))
		} else {
			b.WriteString(fmt.Sprintf("\t%s: %v", key, value))
		}
	}

	b.WriteString("\n")

	return b.Bytes(), nil
}
