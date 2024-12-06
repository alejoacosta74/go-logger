package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestWithLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		messages []struct {
			level   string
			message string
		}
		wantErr      bool
		wantMessages []string // messages that should appear in output
	}{
		{
			name:  "valid level - info",
			level: "info",
			messages: []struct {
				level   string
				message string
			}{
				{"debug", "debug message"},
				{"info", "info message"},
				{"warn", "warn message"},
			},
			wantErr: false,
			wantMessages: []string{
				"info message",
				"warn message",
			},
		},
		{
			name:  "valid level - debug",
			level: "debug",
			messages: []struct {
				level   string
				message string
			}{
				{"debug", "debug message"},
				{"info", "info message"},
				{"warn", "warn message"},
			},
			wantErr: false,
			wantMessages: []string{
				"debug message",
				"info message",
				"warn message",
			},
		},
		{
			name:     "invalid level",
			level:    "invalid",
			wantErr:  true,
			messages: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewLogger(
				WithLevel(tt.level),
				WithOutput(&buf),
				// WithTestFormatter(), // Use test formatter to avoid ANSI codes
			)

			if tt.wantErr {
				if err == nil {
					t.Error("WithLevel() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Log all messages
			for _, msg := range tt.messages {
				switch msg.level {
				case "debug":
					logger.Debug(msg.message)
				case "info":
					logger.Info(msg.message)
				case "warn":
					logger.Warn(msg.message)
				}
			}

			output := buf.String()

			// Check that expected messages appear in output
			for _, wantMsg := range tt.wantMessages {
				if !strings.Contains(output, wantMsg) {
					t.Errorf("Output missing expected message %q", wantMsg)
				}
			}

			// Check that messages below level don't appear
			// For example, if level is info, debug messages shouldn't appear
			for _, msg := range tt.messages {
				if !shouldMessageAppear(tt.level, msg.level) && strings.Contains(output, msg.message) {
					t.Errorf("Output contains message %q that should be filtered by level %s", msg.message, tt.level)
				}
			}
		})
	}
}

// shouldMessageAppear returns true if a message with the given level should appear
// when the logger is set to the given level
func shouldMessageAppear(loggerLevel, messageLevel string) bool {
	levels := map[string]int{
		"panic": 0,
		"fatal": 1,
		"error": 2,
		"warn":  3,
		"info":  4,
		"debug": 5,
		"trace": 6,
	}

	loggerLevelNum, ok1 := levels[loggerLevel]
	messageLevelNum, ok2 := levels[messageLevel]

	if !ok1 || !ok2 {
		return false
	}

	return messageLevelNum <= loggerLevelNum
}

func TestNewSingletonLogger(t *testing.T) {
	t.Cleanup(func() {
		ResetLogger()
	})

	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
		test    func(*testing.T, *Logger)
	}{
		{
			name: "basic singleton creation",
			opts: []Option{},
			test: func(t *testing.T, l *Logger) {
				// Get another instance and verify it's the same
				l2, err := NewSingletonLogger()
				if err != nil {
					t.Fatalf("Failed to get second instance: %v", err)
				}
				if l != l2 {
					t.Error("Second instance is not the same as first")
				}
			},
		},
		{
			name: "singleton with options",
			opts: []Option{
				WithLevel("debug"),
				WithMultipleFields("service", "test"),
			},
			test: func(t *testing.T, l *Logger) {
				if l.Entry.Logger.GetLevel() != logrus.DebugLevel {
					t.Error("Level option not applied")
				}
				if l.Entry.Data["service"] != "test" {
					t.Error("Fields option not applied")
				}
			},
		},
		{
			name: "invalid option",
			opts: []Option{
				WithLevel("invalid"),
			},
			wantErr: true,
			test:    nil,
		},
		{
			name: "concurrent access",
			opts: []Option{},
			test: func(t *testing.T, l *Logger) {
				// Test concurrent access to singleton
				const numGoroutines = 10
				var wg sync.WaitGroup
				instances := make([]*Logger, numGoroutines)
				errs := make([]error, numGoroutines)

				wg.Add(numGoroutines)
				for i := 0; i < numGoroutines; i++ {
					go func(idx int) {
						defer wg.Done()
						instance, err := NewSingletonLogger()
						instances[idx] = instance
						errs[idx] = err
					}(i)
				}
				wg.Wait()

				// Verify all instances are the same and no errors occurred
				for i := 0; i < numGoroutines; i++ {
					if errs[i] != nil {
						t.Errorf("Goroutine %d got error: %v", i, errs[i])
					}
					if instances[i] != l {
						t.Errorf("Goroutine %d got different instance", i)
					}
				}
			},
		},
		{
			name: "option modification after creation",
			opts: []Option{},
			test: func(t *testing.T, l *Logger) {
				// Get first instance
				original := l.Entry.Logger.GetLevel()

				// Try to modify with new instance
				_, err := NewSingletonLogger(WithLevel("debug"))
				if err != nil {
					t.Fatalf("Failed to get second instance: %v", err)
				}

				// Verify level was not changed
				if l.Entry.Logger.GetLevel() != original {
					t.Error("Singleton instance was modified by subsequent calls")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetLogger()
			logger, err := NewSingletonLogger(tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSingletonLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if logger != nil {
					t.Error("Logger should be nil when error occurs")
				}
				return
			}

			if tt.test != nil {
				tt.test(t, logger)
			}
		})
	}
}

// TestSingletonLoggerOutput tests the actual logging functionality
func TestSingletonLoggerOutput(t *testing.T) {

	var buf bytes.Buffer
	logger, err := NewSingletonLogger(
		WithOutput(&buf),
		WithLevel("debug"),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test different log levels
	testCases := []struct {
		level   string
		logFunc func(args ...interface{})
		message string
	}{
		{"debug", logger.Debug, "debug message"},
		{"info", logger.Info, "info message"},
		{"warn", logger.Warn, "warn message"},
		{"error", logger.Error, "error message"},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)
			output := buf.String()
			if !strings.Contains(output, tc.message) {
				t.Errorf("Log output missing message: %s", tc.message)
			}
			if !strings.Contains(output, tc.level) {
				t.Errorf("Log output missing level: %s", tc.level)
			}
		})
	}
}
