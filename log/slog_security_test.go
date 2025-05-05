package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestLogSecurityScenarios tests various security-related scenarios
func TestLogSecurityScenarios(t *testing.T) {
	t.Run("Given a logger with JSON handler", func(t *testing.T) {
		var buf bytes.Buffer
		handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{})
		logger := slog.New(handler)

		t.Run("When logging sensitive data", func(t *testing.T) {
			t.Run("Then it should properly escape special characters", func(t *testing.T) {
				testCases := []struct {
					name     string
					input    string
					expected string
				}{
					{
						"SQL Injection Attempt",
						"'; DROP TABLE users; --",
						"'; DROP TABLE users; --",
					},
					{
						"XSS Attack Attempt",
						"<script>alert('xss')</script>",
						"<script>alert('xss')</script>",
					},
					{
						"Command Injection",
						"$(rm -rf /)",
						"$(rm -rf /)",
					},
				}

				for _, tc := range testCases {
					t.Run(tc.name, func(t *testing.T) {
						buf.Reset()
						logger.Info(tc.input)

						var result map[string]interface{}
						if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
							t.Errorf("Failed to unmarshal JSON: %v", err)
						}

						if msg, ok := result["msg"].(string); !ok {
							t.Errorf("Expected message to be string")
						} else if msg != tc.expected {
							t.Errorf("Expected message to be '%s', got '%s'", tc.expected, msg)
						}
					})
				}
			})

			t.Run("Then it should handle malformed UTF-8 sequences", func(t *testing.T) {
				malformedUTF8 := []byte{0xFF, 0xFE, 0xFD}
				buf.Reset()
				logger.Info(string(malformedUTF8))

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}

				if msg, ok := result["msg"].(string); !ok {
					t.Errorf("Expected message to be string")
				} else if !utf8.ValidString(msg) {
					t.Error("Expected message to be valid UTF-8")
				}
			})
		})

		t.Run("When logging with nested structures", func(t *testing.T) {
			t.Run("Then it should handle circular references", func(t *testing.T) {
				type Circular struct {
					Self *Circular
				}
				circular := &Circular{}
				circular.Self = circular

				buf.Reset()
				logger.Info("circular reference", "circular", circular)

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})

			t.Run("Then it should handle deeply nested structures", func(t *testing.T) {
				nested := make(map[string]interface{})
				current := nested
				for i := 0; i < 100; i++ {
					current["nested"] = make(map[string]interface{})
					current = current["nested"].(map[string]interface{})
				}

				buf.Reset()
				logger.Info("deeply nested", "data", nested)

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})
		})

		t.Run("When logging with large payloads", func(t *testing.T) {
			t.Run("Then it should handle large strings", func(t *testing.T) {
				largeString := strings.Repeat("a", 1024*1024) // 1MB string
				buf.Reset()
				logger.Info(largeString)

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})

			t.Run("Then it should handle large number of attributes", func(t *testing.T) {
				attrs := make([]interface{}, 0, 2000) // Ensure capacity for key-value pairs
				for i := 0; i < 1000; i++ {
					attrs = append(attrs, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
				}

				buf.Reset()
				logger.Info("many attributes", attrs...)

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})
		})

		t.Run("When logging with special characters", func(t *testing.T) {
			t.Run("Then it should handle control characters", func(t *testing.T) {
				controlChars := make([]byte, 32)
				for i := 0; i < 32; i++ {
					controlChars[i] = byte(i)
				}

				buf.Reset()
				logger.Info(string(controlChars))

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})

			t.Run("Then it should handle Unicode private use area", func(t *testing.T) {
				privateUseChars := []rune{
					0xE000,   // Start of private use area
					0xF8FF,   // End of private use area
					0x100000, // Supplementary private use area
					0x10FFFD, // End of supplementary private use area
				}

				buf.Reset()
				logger.Info(string(privateUseChars))

				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			})
		})

		t.Run("When logging with concurrent access", func(t *testing.T) {
			t.Run("Then it should handle concurrent writes", func(t *testing.T) {
				concurrentWrites := 100
				done := make(chan bool)

				for i := 0; i < concurrentWrites; i++ {
					go func(id int) {
						logger.Info("concurrent write", "id", id)
						done <- true
					}(i)
				}

				for i := 0; i < concurrentWrites; i++ {
					<-done
				}
			})
		})
	})
}
