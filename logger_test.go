package logger

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// testLoggerCall represents different levels of nesting for logging calls
type testLoggerCall struct {
	name     string
	logFunc  func(logger *loggerWrapper)
	expected struct {
		funcPattern string
		filePattern string
	}
}

// level1Logging demonstrates direct logging
func level1Logging(logger *loggerWrapper) {
	logger.Info("Level 1 logging")
}

// level2Logging demonstrates one level of nesting
func level2Logging(logger *loggerWrapper) {
	nestedLogging(logger)
}

// level3Logging demonstrates two levels of nesting
func level3Logging(logger *loggerWrapper) {
	level2Logging(logger)
}

// nestedLogging is a helper function for nested calls
func nestedLogging(logger *loggerWrapper) {
	logger.Info("Nested logging")
}

func TestWithRuntimeContext(t *testing.T) {
	tests := []testLoggerCall{
		{
			name:    "direct_logging",
			logFunc: level1Logging,
			expected: struct {
				funcPattern string
				filePattern string
			}{
				funcPattern: `func: go-logger.level1Logging`,
				filePattern: `src: logger_test.go:\d+`,
			},
		},
		{
			name:    "one_level_nested",
			logFunc: level2Logging,
			expected: struct {
				funcPattern string
				filePattern string
			}{
				funcPattern: `func: go-logger.nestedLogging`,
				filePattern: `src: logger_test.go:\d+`,
			},
		},
		{
			name:    "two_levels_nested",
			logFunc: level3Logging,
			expected: struct {
				funcPattern string
				filePattern string
			}{
				funcPattern: `func: go-logger.nestedLogging`,
				filePattern: `src: logger_test.go:\d+`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger, err := NewLogger(
				WithRuntimeContext(),
				WithOutput(&buf),
			)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			tt.logFunc(logger)

			output := buf.String()
			strippedOutput := stripANSI(output)
			funcMatched, err := regexp.MatchString(tt.expected.funcPattern, strippedOutput)
			if err != nil {
				t.Fatalf("Invalid function pattern: %v", err)
			}
			if !funcMatched {
				t.Errorf("Function pattern mismatch\nexpected pattern: %s\ngot: %s",
					tt.expected.funcPattern, strippedOutput)
			}

			fileMatched, err := regexp.MatchString(tt.expected.filePattern, strippedOutput)
			if err != nil {
				t.Fatalf("Invalid file pattern: %v", err)
			}
			if !fileMatched {
				t.Errorf("File pattern mismatch\nexpected pattern: %s\ngot: %s",
					tt.expected.filePattern, strippedOutput)
			}

			if strings.Contains(strippedOutput, "logrus") || strings.Contains(strippedOutput, "createNewLogger") {
				t.Error("Log contains internal logger package calls")
			}
		})
	}
}

// TestWithRuntimeContextConcurrent tests the runtime context in concurrent scenarios
func TestWithRuntimeContextConcurrent(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewLogger(
		WithRuntimeContext(),
		WithOutput(&buf),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Run multiple goroutines to ensure thread safety
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			level3Logging(logger)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the output contains the correct caller information
	output := buf.String()
	strippedOutput := stripANSI(output)
	if strings.Contains(strippedOutput, "logrus") || strings.Contains(strippedOutput, "createNewLogger") {
		t.Error("Log contains internal logger package calls in concurrent scenario")
	}
}

// TestWithRuntimeContextCustomization verifies that the runtime context formatting can be customized
func TestWithRuntimeContextCustomization(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewLogger(
		WithRuntimeContext(),
		WithOutput(&buf),
		WithMultipleFields("service", "test-service"),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Test message")

	output := buf.String()
	strippedOutput := stripANSI(output)
	if !strings.Contains(strippedOutput, "service=test-service") {
		t.Errorf("Custom fields not present in runtime context output\nOutput: %q\nSearching for: %q",
			strippedOutput, "service=test-service")
	}
}

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
		test    func(*testing.T, *loggerWrapper)
	}{
		{
			name: "basic singleton creation",
			opts: []Option{},
			test: func(t *testing.T, l *loggerWrapper) {
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
			test: func(t *testing.T, l *loggerWrapper) {
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
			test: func(t *testing.T, l *loggerWrapper) {
				// Test concurrent access to singleton
				const numGoroutines = 10
				var wg sync.WaitGroup
				instances := make([]*loggerWrapper, numGoroutines)
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
			test: func(t *testing.T, l *loggerWrapper) {
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

func TestRuntimeContextHook(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		message   string
		logFunc   func(*loggerWrapper, string)
		wantHook  bool
		wantField map[string]*regexp.Regexp
	}{
		{
			name:     "debug level should add runtime context",
			logLevel: "debug",
			message:  "debug message",
			logFunc:  func(l *loggerWrapper, msg string) { l.Debug(msg) },
			wantHook: true,
			wantField: map[string]*regexp.Regexp{
				"func": regexp.MustCompile(`go-logger\.TestRuntimeContextHook\.func\d+`),
				"src":  regexp.MustCompile(`logger_test\.go:\d+`),
			},
		},
		{
			name:     "trace level should add runtime context",
			logLevel: "trace",
			message:  "trace message",
			logFunc:  func(l *loggerWrapper, msg string) { l.Trace(msg) },
			wantHook: true,
			wantField: map[string]*regexp.Regexp{
				"func": regexp.MustCompile(`go-logger\.TestRuntimeContextHook\.func\d+`),
				"src":  regexp.MustCompile(`logger_test\.go:\d+`),
			},
		},
		{
			name:      "info level should not add runtime context",
			logLevel:  "info",
			message:   "info message",
			logFunc:   func(l *loggerWrapper, msg string) { l.Info(msg) },
			wantHook:  false,
			wantField: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger, err := NewLogger(
				WithLevel(tt.logLevel),
				WithOutput(&buf),
			)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			tt.logFunc(logger, tt.message)
			output := buf.String()
			strippedOutput := stripANSI(output)

			// Verify runtime context fields
			if tt.wantHook {
				for field, pattern := range tt.wantField {
					if !pattern.MatchString(strippedOutput) {
						t.Errorf("Missing or incorrect %s\nwant pattern: %q\ngot output: %q",
							field, pattern.String(), strippedOutput)
					}
				}
			} else {
				// Verify no runtime context was added
				if strings.Contains(strippedOutput, "func:") || strings.Contains(strippedOutput, "src:") {
					t.Errorf("Runtime context was added when it shouldn't be\noutput: %q", strippedOutput)
				}
			}

			// Verify message was logged
			if !strings.Contains(strippedOutput, tt.message) {
				t.Errorf("Log message missing\nwant: %q\ngot: %q", tt.message, strippedOutput)
			}
		})
	}
}

// helper function to strip ANSI color codes
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
