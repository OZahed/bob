package postgres

import (
	"database/sql"
	"fmt"

	"github.com/OZahed/bob/db"
	"github.com/XSAM/otelsql"
	"github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// driverName is the name of the postgres driver.
// Every driver must have a unique name based on the database or underlying database connection.
const driverName = "postgres"

// Options is the options for connecting to the database.
// See https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
type DbOptions struct {
	// Host is the host name or IP address of the database server.
	host string

	// Port is the port number of the database server. default: 5432
	port string

	// Username is the username to use when connecting to the database.
	username string

	// Password is the password to use when connecting to the database.
	password string

	// Database is the name of the database to connect to.
	database string

	// SSLMode is the SSL mode to use when connecting to the database.
	sslMode string

	// SSLCert is the path to the SSL certificate to use when connecting to the database.
	sslCert string

	monitoring MonitoringOpts
}

type MonitoringOpts struct {
	// Enabled is the flag to enable monitoring.
	enabled bool

	// Tracing is the flag to enable tracing.
	tracing bool
}

// DbOpt is option pattern implementation
type DbOpt func(*DbOptions)

// New returns a new instance of a postgres database.
func NewDb(opts ...DbOpt) (db.Database, error) {
	// Parse database url
	o := &DbOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return openDB(url, opts.Monitoring)
}

// NewFromURL returns a new instance of a postgres database from a URL.
func NewFromURL(url string, mtnOpts MonitoringOpts) (db.Database, error) {
	opts, err := pq.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing database url: %w", err)
	}

	return openDB(opts, mtnOpts)
}

// openDB opens a connection to the database.
func openDB(url string, mtnOpts MonitoringOpts) (*sql.DB, error) {
	if !mtnOpts.Enabled {
		conn, err := sql.Open(driverName, url)
		if err != nil {
			return nil, fmt.Errorf("error connecting to database: %w", err)
		}
		return conn, nil
	}

	conn, err := otelsql.Open(driverName, url, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, fmt.Errorf("error connecting to database with OpenTelemetry Support: %w", err)
	}

	if mtnOpts.Tracing {
		err = otelsql.RegisterDBStatsMetrics(conn, otelsql.WithAttributes(
			semconv.DBSystemMySQL,
		))
		if err != nil {
			return nil, fmt.Errorf("error registering database metrics: %w", err)
		}
	}

	return conn, nil
}
