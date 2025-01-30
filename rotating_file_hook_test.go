package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotatingFileHook(t *testing.T) {
	// Create temp directory for test logs
	tmpDir, err := os.MkdirTemp("", "rotating_file_hook_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		config        *RotatingFileConfig
		logEntries    []logEntry
		expectedError error
		validation    func(t *testing.T, tmpDir string, filename string)
	}{
		{
			name: "default configuration",
			config: &RotatingFileConfig{
				Filename: filepath.Join(tmpDir, "default.log"),
			},
			logEntries: []logEntry{
				{level: logrus.InfoLevel, message: "test message"},
			},
			validation: func(t *testing.T, tmpDir string, filename string) {
				assert.FileExists(t, filename)
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.Contains(t, string(content), "test message")
			},
		},
		{
			name: "custom levels",
			config: &RotatingFileConfig{
				Filename: filepath.Join(tmpDir, "levels.log"),
				Levels:   []logrus.Level{logrus.ErrorLevel},
			},
			logEntries: []logEntry{
				{level: logrus.InfoLevel, message: "info message"},
				{level: logrus.ErrorLevel, message: "error message"},
			},
			validation: func(t *testing.T, tmpDir string, filename string) {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.NotContains(t, string(content), "info message")
				assert.Contains(t, string(content), "error message")
			},
		},
		{
			name: "with fields",
			config: &RotatingFileConfig{
				Filename: filepath.Join(tmpDir, "fields.log"),
			},
			logEntries: []logEntry{
				{
					level:   logrus.InfoLevel,
					message: "test with fields",
					fields:  logrus.Fields{"key": "value"},
				},
			},
			validation: func(t *testing.T, tmpDir string, filename string) {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.Contains(t, string(content), "key=value")
			},
		},
		{
			name:   "nil config",
			config: nil,
			logEntries: []logEntry{
				{level: logrus.InfoLevel, message: "test message"},
			},
			validation: func(t *testing.T, tmpDir string, filename string) {
				assert.FileExists(t, filepath.Join("logs", "app.log"))
				// Cleanup default log file
				defer os.RemoveAll("logs")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new hook
			hook, err := newRotatingFileHook(tt.config)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			}
			require.NoError(t, err)
			defer hook.Close()

			// Create test logger
			logger := logrus.New()
			logger.AddHook(hook)
			logger.SetOutput(io.Discard)

			// Write log entries
			for _, entry := range tt.logEntries {
				logWithFields(logger, entry)
			}

			// Run validation
			if tt.validation != nil {
				filename := "logs/app.log"
				if tt.config != nil {
					filename = tt.config.Filename
				}
				tt.validation(t, tmpDir, filename)
			}
		})
	}
}

// Helper structs and functions

type logEntry struct {
	level   logrus.Level
	message string
	fields  logrus.Fields
}

func logWithFields(logger *logrus.Logger, entry logEntry) {
	if entry.fields != nil {
		logger.WithFields(entry.fields).Log(entry.level, entry.message)
	} else {
		logger.Log(entry.level, entry.message)
	}
}

func TestRotatingFileHook_Concurrent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rotating_file_hook_concurrent")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &RotatingFileConfig{
		Filename:   filepath.Join(tmpDir, "concurrent.log"),
		MaxSize:    1,
		MaxBackups: 3,
	}

	hook, err := newRotatingFileHook(config)
	require.NoError(t, err)
	defer hook.Close()

	logger := logrus.New()
	logger.AddHook(hook)
	logger.SetOutput(io.Discard)

	// Test concurrent logging
	concurrentTests := 100
	done := make(chan bool)

	for i := 0; i < concurrentTests; i++ {
		go func(num int) {
			logger.WithFields(logrus.Fields{
				"goroutine": num,
				"timestamp": time.Now().UnixNano(),
			}).Info("Concurrent log message")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrentTests; i++ {
		<-done
	}

	// Verify log file exists and contains data
	assert.FileExists(t, config.Filename)
	content, err := os.ReadFile(config.Filename)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Concurrent log message")
}

func TestRotatingFileHook_Levels(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rotating_file_hook_levels")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		filename      string
		levels        []logrus.Level
		logLevel      logrus.Level
		message       string
		shouldContain bool
	}{
		{
			name:          "log at configured level",
			filename:      "configured_level",
			levels:        []logrus.Level{logrus.InfoLevel},
			logLevel:      logrus.InfoLevel,
			message:       "should be logged",
			shouldContain: true,
		},
		{
			name:          "log below configured level",
			filename:      "below_configured_level",
			levels:        []logrus.Level{logrus.ErrorLevel},
			logLevel:      logrus.WarnLevel,
			message:       "should not be logged",
			shouldContain: false,
		},
		{
			name:          "multiple levels",
			filename:      "multiple_levels",
			levels:        []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel},
			logLevel:      logrus.WarnLevel,
			message:       "should be logged",
			shouldContain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile := filepath.Join(tmpDir, fmt.Sprintf("%s.log", tt.filename))

			config := &RotatingFileConfig{
				Filename: logFile,
				Levels:   tt.levels,
			}

			hook, err := newRotatingFileHook(config)
			require.NoError(t, err)
			defer hook.Close()

			logger := logrus.New()
			logger.AddHook(hook)
			logger.SetOutput(io.Discard) // Prevent output to stderr

			// Log the message
			logger.Log(tt.logLevel, tt.message)

			// Add small delay to ensure writing completes
			time.Sleep(10 * time.Millisecond)

			if tt.shouldContain {
				// Verify file contents
				content, err := os.ReadFile(logFile)
				require.NoError(t, err, "Failed to read log file")
				assert.Contains(t, string(content), tt.message,
					"Log file should contain message for level %v", tt.logLevel)
			} else {
				// assert no file exists
				assert.NoFileExists(t, logFile)
			}
		})
	}
}
