package db

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type SQLDriverInstance interface {
	Name() string
	ConnectionString() string
	DBName() string
}

type Config struct {
	Prometheus  bool
	Otel        bool
	Sqlx        bool
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
}

func NewDatabaseConnection(cfg Config, driver SQLDriverInstance) (*sql.DB, error) {
	var dbc *sql.DB
	var err error

	if cfg.Otel {
		dbc, err = otelsql.Open(driver.Name(), driver.ConnectionString(),
			otelsql.WithAttributes(getAttribute(driver.Name())),
			otelsql.WithDBName(driver.DBName()),
		)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Sqlx {
		sqlx.NewDb(dbc, driver.Name())
		otelsql.ReportDBStatsMetrics(dbc, otelsql.WithAttributes(getAttribute(driver.Name())))
	}

	if cfg.MaxIdle > 0 {
		dbc.SetMaxIdleConns(cfg.MaxIdle)
	}

	if cfg.MaxOpen > 0 {
		dbc.SetMaxOpenConns(cfg.MaxOpen)
	}

	if cfg.MaxLifetime > 0 {
		dbc.SetConnMaxLifetime(cfg.MaxLifetime)
	}

	// Add Prometheus metrics
	return dbc, nil
}

func getAttribute(driverName string) attribute.KeyValue {
	switch driverName {
	case "mysql":
		return semconv.DBSystemMySQL
	case "postgres":
		return semconv.DBSystemPostgreSQL
	case "sqlite3":
		return semconv.DBSystemSqlite
	case "mssql":
		return semconv.DBSystemMSSQL
	default:
		return semconv.DBSystemOtherSQL
	}
}
