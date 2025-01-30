package logger

import (
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

// Type Fields is an alias for logrus.Fields
type Fields = logrus.Fields

var (
	// Log is the global logger instance
	Log *Logger

	loggerOnce sync.Once
	// colorFormatter is a custom colored formatter for debug and trace levels
	colorFormatter *ColorFormatter = &ColorFormatter{
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
	Log = &Logger{
		Entry: entry,
	}
}

func NewLogger(opts ...Option) (*Logger, error) {
	logger, err := createNewLogger(opts...)
	if err != nil {
		return nil, err
	}
	Log = logger
	return logger, nil
}

func NewSingletonLogger(opts ...Option) (*Logger, error) {
	var err error
	loggerOnce.Do(func() {
		var logger *Logger
		logger, err = createNewLogger(opts...)
		if err == nil {
			Log = logger
		}
	})
	return Log, err
}

func createNewLogger(opts ...Option) (*Logger, error) {
	l := logrus.New()

	logger := &Logger{
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

// Trace logs a message at the trace level using the global Log instance.
// This function modifies the global Log's level to trace and accepts variadic arguments
// that will be formatted using fmt.Sprint.
func Trace(args ...interface{}) {
	Log.Trace(args...)
}

// Tracef logs a formatted message at the trace level using the global Log instance.
// This function modifies the global Log's level to trace and accepts a format string
// and variadic arguments that will be formatted using fmt.Sprintf.
func Tracef(format string, args ...interface{}) {
	Log.Tracef(format, args...)
}

// Debug logs a message at the debug level using the global Log instance.
// This function modifies the global Log's level to debug and accepts variadic arguments
// that will be formatted using fmt.Sprint.
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf logs a formatted message at the debug level using the global Log instance.
// This function modifies the global Log's level to debug and accepts a format string
// and variadic arguments that will be formatted using fmt.Sprintf.
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Info logs a message at the info level using the global Log instance.
// This function modifies the global Log's level to info and accepts variadic arguments
// that will be formatted using fmt.Sprint.
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof logs a formatted message at the info level using the global Log instance.
// This function modifies the global Log's level to info and accepts a format string
// and variadic arguments that will be formatted using fmt.Sprintf.
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Warn logs a message at the warn level using the global Log instance.
// This function modifies the global Log's level to warn and accepts variadic arguments
// that will be formatted using fmt.Sprint.
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf logs a formatted message at the warn level using the global Log instance.
// This function modifies the global Log's level to warn and accepts a format string
// and variadic arguments that will be formatted using fmt.Sprintf.
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// Error logs a message at the error level using the global Log instance.
// This function modifies the global Log's level to error and accepts variadic arguments
// that will be formatted using fmt.Sprint.
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf logs a formatted message at the error level using the global Log instance.
// This function modifies the global Log's level to error and accepts a format string
// and variadic arguments that will be formatted using fmt.Sprintf.
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatal logs a message at the fatal level using the global Log instance and then exits.
// This function modifies the global Log's level to fatal, accepts variadic arguments
// that will be formatted using fmt.Sprint, and terminates the program with os.Exit(1).
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Fatalf logs a formatted message at the fatal level using the global Log instance and then exits.
// This function modifies the global Log's level to fatal, accepts a format string and variadic
// arguments that will be formatted using fmt.Sprintf, and terminates the program with os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// Panic logs a message at the panic level using the global Log instance and then panics.
// This function modifies the global Log's level to panic, accepts variadic arguments
// that will be formatted using fmt.Sprint, and calls panic() with the resulting string.
func Panic(args ...interface{}) {
	Log.Panic(args...)
}

// Panicf logs a formatted message at the panic level using the global Log instance and then panics.
// This function modifies the global Log's level to panic, accepts a format string and variadic
// arguments that will be formatted using fmt.Sprintf, and calls panic() with the resulting string.
func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
}

// WithField adds a single field to the logger entry. It takes a key string and a value of any type,
// and returns a new Logger instance with the field added. This is useful for adding contextual
// information to log entries, such as request IDs, user IDs, or any other metadata that helps
// trace and debug issues.
func WithField(key string, value interface{}) *Logger {
	return &Logger{Entry: Log.Entry.WithField(key, value)}

}

// SetOutput sets the output destination for the global logger
func SetOutput(output io.Writer) {
	Log.Entry.Logger.SetOutput(output)
}

// AddFileOutputHook adds a file hook to the global logger
func AddFileOutputHook(filename string, cfg *RotatingFileConfig, levels ...logrus.Level) error {
	if cfg == nil {
		cfg = &RotatingFileConfig{}
	}
	if cfg.Filename == "" {
		cfg.Filename = filename
	}
	cfg.Levels = levels
	hook, err := newRotatingFileHook(cfg)
	if err != nil {
		return err
	}
	Log.Entry.Logger.AddHook(hook)
	return nil
}

// NullOutput sets the logger output to io.Discard, effectively disabling all log output.
// This is useful for testing scenarios where log output needs to be suppressed.
func NullOutput() {
	Log.Entry.Logger.SetOutput(io.Discard)
}

func WithFields(fields ...string) *Logger {
	if len(fields)%2 != 0 {
		panic("WithFields requires an even number of arguments")
	}

	f := make(logrus.Fields)
	for i := 0; i < len(fields); i += 2 {
		f[fields[i]] = fields[i+1]
	}
	return &Logger{Entry: Log.WithFields(f)}
}

func SetLevel(level string) {
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	Log.Entry.Logger.SetLevel(parsedLevel)
	if parsedLevel == logrus.DebugLevel || parsedLevel == logrus.TraceLevel {
		// set the color formatter
		Log.Entry.Logger.SetFormatter(colorFormatter)
		// add the runtime context hook
		Log.Entry.Logger.AddHook(NewRuntimeContextHook(3))
	}
}

// ResetLogger resets the singleton instance (for testing)
func ResetLogger() {
	Log = nil
	loggerOnce = sync.Once{}
}
