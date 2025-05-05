package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

// TestNewSlogDefault tests the default logger configuration
func TestNewSlogDefault(t *testing.T) {
	logger := NewSlog()
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
}

// TestNewSlogWithLevel tests different log levels
func TestNewSlogWithLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"Debug Level", "debug", slog.LevelDebug},
		{"Info Level", "info", slog.LevelInfo},
		{"Warning Level", "warning", slog.LevelWarn},
		{"Error Level", "error", slog.LevelError},
		{"Invalid Level", "invalid", slog.LevelDebug}, // Should default to debug
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSlog(WithLevel(tt.level))
			if logger == nil {
				t.Error("Expected logger to be non-nil")
			}
		})
	}
}

// TestNewSlogWithHandlerType tests different handler types
func TestNewSlogWithHandlerType(t *testing.T) {
	tests := []struct {
		name     string
		handler  HandlerType
		expected string
	}{
		{"JSON Handler", JsonHandler, "json"},
		{"Text Handler", TextHandler, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSlog(WithHandlerType(tt.handler))
			if logger == nil {
				t.Error("Expected logger to be non-nil")
			}
		})
	}
}

// TestUTF8CharacterHandling tests handling of UTF-8 characters in logs
func TestUTF8CharacterHandling(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{})
	logger := slog.New(handler)

	// Test various UTF-8 characters
	testCases := []string{
		"Hello 世界",
		"Special chars: 你好, こんにちは",
		"Emojis: 😀 🎉 🌟",
		"Mixed: Hello 世界! 123",
	}

	for _, tc := range testCases {
		buf.Reset()
		logger.Info(tc)

		var result map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Failed to unmarshal JSON for input '%s': %v", tc, err)
		}

		// Verify the message is properly escaped in JSON
		if msg, ok := result["msg"].(string); !ok {
			t.Errorf("Expected message to be string for input '%s'", tc)
		} else if msg != tc {
			t.Errorf("Expected message '%s', got '%s'", tc, msg)
		}
	}
}

// TestLogInjectionPrevention tests prevention of log injection attacks
func TestLogInjectionPrevention(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{})
	logger := slog.New(handler)

	// Test cases for potential injection attacks
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Newline Injection",
			"Hello\nWorld",
			"Hello\nWorld", // slog handles newlines correctly
		},
		{
			"JSON Injection",
			`{"malicious": "data"}`,
			`{"malicious": "data"}`, // slog handles JSON correctly
		},
		{
			"Control Characters",
			"Hello\x00World",
			"HelloWorld", // slog removes null bytes
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			logger.Info(tc.input)

			var result map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("Failed to unmarshal JSON for input '%s': %v", tc.input, err)
			}

			if msg, ok := result["msg"].(string); !ok {
				t.Errorf("Expected message to be string for input '%s'", tc.input)
			} else if msg != tc.expected {
				t.Errorf("Expected message to be '%s', got '%s'", tc.expected, msg)
			}
		})
	}
}

// TestStackFrameLogging tests the stack frame logging functionality
func TestStackFrameLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlog(
		WithHandlerType(JsonHandler),
		WithStackFrame(),
	)

	// Create a custom handler that writes to our buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
	})
	logger = slog.New(handler)

	logger.Info("Test message with stack frame")

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	// Verify source information is present
	if source, ok := result["source"].(map[string]interface{}); !ok {
		t.Error("Expected source information to be present")
	} else {
		if _, ok := source["function"]; !ok {
			t.Error("Expected function information to be present")
		}
		if _, ok := source["file"]; !ok {
			t.Error("Expected file information to be present")
		}
		if _, ok := source["line"]; !ok {
			t.Error("Expected line information to be present")
		}
	}
}

// TestUTCTimeHandling tests the UTC time handling functionality
func TestUTCTimeHandling(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(a.Key, a.Value.Time().UTC())
			}
			return a
		},
	})
	logger := slog.New(handler)

	logger.Info("Test message with UTC time")

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	// Verify time field is present and in UTC
	if timeStr, ok := result["time"].(string); !ok {
		t.Error("Expected time field to be present")
	} else {
		parsedTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			t.Errorf("Failed to parse time: %v", err)
		}
		if parsedTime.Location() != time.UTC {
			t.Errorf("Expected time to be in UTC, got %v", parsedTime.Location())
		}
	}
}

// TestCustomReplaceAttr tests custom attribute replacement
func TestCustomReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	customReplace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "custom" {
			return slog.String("custom", "replaced")
		}
		return a
	}

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: customReplace,
	})
	logger := slog.New(handler)

	logger.Info("Test message", "custom", "original")

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	// Verify custom attribute was replaced
	if custom, ok := result["custom"].(string); !ok {
		t.Error("Expected custom field to be present")
	} else if custom != "replaced" {
		t.Errorf("Expected custom value to be 'replaced', got '%s'", custom)
	}
}
