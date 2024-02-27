package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/OZahed/bob/db"
	"github.com/XSAM/otelsql"
	"github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// driverName is the name of the postgres driver.
// Every driver must have a unique name based on the database or underlying database connection.
const driverName = "postgres"

// Options is the Options for connecting to the database.
// See https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
type Options struct {
	// Host is the host name or IP address of the database server.
	Host string

	// Port is the port number of the database server. default: 5432
	Port string

	// Username is the username to use when connecting to the database.
	Username string

	// Password is the password to use when connecting to the database.
	Password string

	// Database is the name of the database to connect to.
	Database string

	// SSLMode is the SSL mode to use when connecting to the database.
	SSLMode string

	// SSLCert is the path to the SSL certificate to use when connecting to the database.
	SSLCert string

	Monitoring MonitoringOpts
}

type MonitoringOpts struct {
	// Enabled is the flag to enable monitoring.
	Enabled bool

	// Tracing is the flag to enable tracing.
	Tracing bool
}

type DBOption func(opts *Options) *Options

func WithHost(hostname string) DBOption {
	hostname = strings.TrimSpace(hostname)

	return func(opts *Options) *Options {
		opts.Host = hostname
		return opts
	}
}

func WithPort(port int) DBOption {
	portString := strconv.Itoa(port)

	return func(opts *Options) *Options {
		opts.Port = portString
		return opts
	}
}

func WithUsername(username string) DBOption {
	username = strings.TrimSpace(username)

	return func(opts *Options) *Options {
		opts.Username = username
		return opts
	}
}

func WithPassword(password string) DBOption {
	password = strings.TrimSpace(password)

	return func(opts *Options) *Options {
		opts.Password = password
		return opts
	}
}

func WithSSLMode(sslMode string) DBOption {
	sslMode = strings.TrimSpace(sslMode)

	return func(opts *Options) *Options {
		opts.SSLMode = sslMode
		return opts
	}
}

func WithSSLCert(sslCert string) DBOption {
	sslCert = strings.TrimSpace(sslCert)

	return func(opts *Options) *Options {
		if opts.SSLMode != "disable" {
			opts.SSLCert = sslCert
		} else {
			errors.New("SSL certificate not provided")
		}

		return opts
	}
}

func WithMonitoringOpts(monitoringOpts MonitoringOpts) DBOption {
	return func(opts *Options) *Options {
		opts.Monitoring = monitoringOpts
		return opts
	}
}

// New returns a new instance of a postgres database.
func NewFromOption(dbOptions ...DBOption) (db.Database, error) {
	if len(dbOptions) == 0 {
		return nil, fmt.Errorf("Options not provided")
	}

	opts := optionsBuilder(dbOptions...)

	// Parse database url
	url := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		driverName, opts.Username, opts.Password, opts.Host, opts.Port, opts.Database, opts.SSLMode,
	)

	if opts.SSLMode != "disable" {
		url = fmt.Sprintf("%s&sslrootcert=%s", url, opts.SSLCert)
	}

	return openDB(url, opts.Monitoring)
}

func optionsBuilder(dbOptions ...DBOption) *Options {
	opts := &Options{}

	for _, opt := range dbOptions {
		opt(opts)
	}

	return opts
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
