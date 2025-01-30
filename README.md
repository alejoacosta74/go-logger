# Go Logger Library

A flexible and feature-rich logging library for Go applications, built on top of [logrus](https://github.com/sirupsen/logrus). This library provides enhanced logging capabilities with runtime context, customizable outputs, log levels, and field support.

## Features

- A global logger instance ready to use
- Multiple logging levels (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- Runtime context capturing (function names, file locations)
- Singleton logger support
- Customizable outputs (console, file, null)
- Field-based logging
- Thread-safe operation
- ANSI color support
- Flexible configuration through options pattern

## Installation

```bash
go get github.com/alejoacosta74/go-logger
```

## Basic Usage

- Create a new logger with default settings
```go
package main

import log "github.com/alejoacosta74/go-logger"

func main() {
	// Create a new logger with default settings
	log, err := log.NewLogger()
	if err != nil {
		panic(err)
	}
	// Basic logging
	log.Info("Hello, World!")
	log.Debug("This is a debug message")
	log.Error("Something went wrong")
}
```

- Or use the global logger instance
```go
log.Info("Hello, World!")
log.Debug("This is a debug message")
log.Error("Something went wrong")
```
## Advanced Features

### Setting Log Level

- Set the log level to info on the global logger instance
```go
log.SetLevel("info")
```

- Set the log level to info on a new logger instance
```go
log, err := log.NewLogger(log.WithLevel("info"))
if err != nil {
	panic(err)
}
```

### Adding Runtime Context

Runtime context adds function name and file location to log entries:

```go
logger, := log.NewLogger(
	log.WithRuntimeContext(),
)
logger.Debug("This message will include runtime context")
// Output: level=debug func=main.myFunction src=main.go:25 msg="This message will include runtime context"
```

### Custom Output Destinations

```go
// Write to a file
logger, := log.NewLogger(
	log.WithFileOutput("app.log"),
)
// Write to custom writer
var buf bytes.Buffer
logger, := log.NewLogger(
	log.WithOutput(&buf),
)
// Disable output (for testing)
logger, := log.NewLogger(
	log.WithNullOutput(),
)
```

### Add logging to a file

The file will be rotated when the max size is reached.

```go
log.AddFileOutputHook("app.log", &log.RotatingFileConfig{
	MaxSize:    10,
	MaxBackups: 3,
	MaxAge:     7,
	Compress:   true,
})
```

### Using Fields

```go
// Single field
logger.WithField("user_id", "123").Info("User logged in")
// Multiple fields
logger, := logger.NewLogger(
	log.WithMultipleFields(
		"service", "auth",
		"environment", "production",
	),
)
// Or add fields to existing logger
logger.WithFields(
"request_id", "abc-123",
"user_id", "456",
).Info("Request processed")
```
### Singleton Logger

```go
// Create or get singleton instance
logger, err := log.NewSingletonLogger(
	log.WithLevel("info"),
	log.WithRuntimeContext(),
)
// Use package-level functions
logger.Info("Using singleton logger")
// or
log.Debug("Debug message")
```

### Formatting Options

```go
log.Infof("Hello %s", "World")
log.Debugf("Count: %d", 42)
log.Errorf("Failed to process: %v", err)
```

## Available Log Levels

- `Trace`: Most verbose level
- `Debug`: Debugging information
- `Info`: General operational information
- `Warn`: Warning messages
- `Error`: Error messages
- `Fatal`: Fatal errors (calls os.Exit(1))
- `Panic`: Panic messages (calls panic())

## Thread Safety

The logger is safe for concurrent use. All logging operations are thread-safe, and the singleton pattern implementation ensures safe initialization in concurrent environments.

## Testing

The library includes comprehensive test coverage. Run tests with:
```bash
go test ./...
```

