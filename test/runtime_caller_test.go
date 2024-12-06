package test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/alejoacosta74/go-logger"
)

// testLoggerCall represents different levels of nesting for logging calls
type testLoggerCall struct {
	name     string
	logFunc  func(logger *logger.Logger)
	expected struct {
		funcPattern string
		filePattern string
	}
}

// level1Logging demonstrates direct logging
func level1Logging(logger *logger.Logger) {
	logger.Info("Level 1 logging")
}

// level2Logging demonstrates one level of nesting
func level2Logging(logger *logger.Logger) {
	nestedLogging(logger)
}

// level3Logging demonstrates two levels of nesting
func level3Logging(logger *logger.Logger) {
	level2Logging(logger)
}

// nestedLogging is a helper function for nested calls
func nestedLogging(logger *logger.Logger) {
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
				funcPattern: `func: test.level1Logging`,
				filePattern: `src: test/runtime_caller_test.go:\d+`,
			},
		},
		{
			name:    "one_level_nested",
			logFunc: level2Logging,
			expected: struct {
				funcPattern string
				filePattern string
			}{
				funcPattern: `func: test.nestedLogging`,
				filePattern: `src: test/runtime_caller_test.go:\d+`,
			},
		},
		{
			name:    "two_levels_nested",
			logFunc: level3Logging,
			expected: struct {
				funcPattern string
				filePattern string
			}{
				funcPattern: `func: test.nestedLogging`,
				filePattern: `src: test/runtime_caller_test.go:\d+`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger, err := logger.NewLogger(
				logger.WithRuntimeContext(),
				logger.WithOutput(&buf),
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
	logger, err := logger.NewLogger(
		logger.WithRuntimeContext(),
		logger.WithOutput(&buf),
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

func TestRuntimeContextHook(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		message   string
		logFunc   func(*logger.Logger, string)
		wantHook  bool
		wantField map[string]*regexp.Regexp
	}{
		{
			name:     "debug level should add runtime context",
			logLevel: "debug",
			message:  "debug message",
			logFunc:  func(l *logger.Logger, msg string) { l.Debug(msg) },
			wantHook: true,
			wantField: map[string]*regexp.Regexp{
				"func": regexp.MustCompile(`test.TestRuntimeContextHook.func\d+`),
				"src":  regexp.MustCompile(`test/runtime_caller_test.go:\d+`),
			},
		},
		{
			name:     "trace level should add runtime context",
			logLevel: "trace",
			message:  "trace message",
			logFunc:  func(l *logger.Logger, msg string) { l.Trace(msg) },
			wantHook: true,
			wantField: map[string]*regexp.Regexp{
				"func": regexp.MustCompile(`test.TestRuntimeContextHook.func\d+`),
				"src":  regexp.MustCompile(`test/runtime_caller_test.go:\d+`),
			},
		},
		{
			name:      "info level should not add runtime context",
			logLevel:  "info",
			message:   "info message",
			logFunc:   func(l *logger.Logger, msg string) { l.Info(msg) },
			wantHook:  false,
			wantField: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger, err := logger.NewLogger(
				logger.WithLevel(tt.logLevel),
				logger.WithOutput(&buf),
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

// TestWithRuntimeContextCustomization verifies that the runtime context formatting can be customized
func TestWithRuntimeContextCustomization(t *testing.T) {
	var buf bytes.Buffer
	logger, err := logger.NewLogger(
		logger.WithRuntimeContext(),
		logger.WithOutput(&buf),
		logger.WithMultipleFields("service", "test-service"),
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

// helper function to strip ANSI color codes
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
