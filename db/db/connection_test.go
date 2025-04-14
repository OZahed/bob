package db

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TestGetAttribute(t *testing.T) {
	tests := []struct {
		driverName string
		expected   attribute.KeyValue
	}{

		{"mysql", semconv.DBSystemMySQL},
		{"postgres", semconv.DBSystemPostgreSQL},
		{"sqlite3", semconv.DBSystemSqlite},
		{"mssql", semconv.DBSystemMSSQL},
		{"unknown", semconv.DBSystemOtherSQL},
	}

	for _, tt := range tests {
		attr := getAttribute(tt.driverName)
		if attr != tt.expected {
			t.Errorf("for driver %q, expected %v, got %v", tt.driverName, tt.expected, attr)
		}
	}
}
