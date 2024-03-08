package db

import (
	"context"
	"database/sql"
)

// Database is a subset of the sql.DB interface.
// It is used to wrap the sql.DB and sql.Tx types with extra functionality.
// This interface is a combination of database management methods and methods
// which works with actual data.
type Database interface {
	Close() error
	Ping() error
	PingContext(ctx context.Context) error
	// Embed the Queryer interface for data fetching and manipulating methods.
	Queryer
}

// Queryer is a subset of the sql.DB interface but only for methods which works with actual data.
// It is used to wrap the sql.DB and sql.Tx types with extra functionality.
type Queryer interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// TODO: Implement Stmt interface compatible Prepare and PrepareContext methods.
	// Prepare and PrepareContext are not included in the Queryer interface
	// because they return a Stmt interface which is not a Queryer compatible with sqlx and sql interface
	// These methods should run on all physical instance of the database.
	// Prepare(query string) (Stmt, error)
	// PrepareContext(ctx context.Context, query string) (Stmt, error)
}

type DatabaseX interface {
	Database
	QueryerX
}

type QueryerX interface {
	Queryer
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}
