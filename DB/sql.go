package db

import (
	"context"
	"database/sql"
	"time"
)

// ReadQuerier is a container for the SQL queries that are meant to be executed against Secondary database list
// Queries and operations such as Get, QueryRow, GetContext, Statement and so on
//
//	some Functionalities are meant to be executed on both read and write for example
//		Prepare(query string) (*sql.Stmt, error)
//		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
type ReadQuerier interface {
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// WriteQuerier is a container for the SQL queries that are meant to be executed against Primary databases list
// Queries and operations such as transactions, Execute Update, Statement and so on
//
//	some Functionalities are meant to be executed on both read and write for example
//		Prepare(query string) (*sql.Stmt, error)
//		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
type WriteQuerier interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// DBConn is the list of connection related actions such as Connecting, Disconnecting
// Ping and so on
type DBConn interface {
	Close() error
	Ping() error
	PingContext(ctx context.Context) error
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
}

// Database is the interface for an object with Separate reader and writer objects inside of it
type Database interface {
	DBConn
	ReadQuerier
	WriteQuerier
	GetReader() ReadQuerier
	GetWriter() WriteQuerier
}
