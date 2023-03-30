package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// ReadQuerierX is a container for the SQL queries that are meant to be executed against Secondary database list
// Queries and operations such as Get, QueryRow, GetContext, Statement and so on
//
//	some Functionalities are meant to be executed on both read and write for example
//		Prepare(query string) (*sql.Stmt, error)
//		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
type ReadQuerierX interface {
	ReadQuerier

	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MapperFunc(mf func(string) string)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// WriteQuerierx is a container for the SQL queries that are meant to be executed against Primary databases list
// Queries and operations such as transactions, Execute Update, Statement and so on
//
//	some Functionalities are meant to be executed on both read and write for example
//		Prepare(query string) (*sql.Stmt, error)
//		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
type WriteQuerierX interface {
	WriteQuerier

	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
	Beginx() (*sqlx.Tx, error)
	MustBegin() *sqlx.Tx
	MustBeginTx(ctx context.Context, opts *sql.TxOptions) *sqlx.Tx
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type DatabaseX interface {
	DBConn
	ReadQuerierX
	WriteQuerierX
	GetReaderX() ReadQuerierX
	GetWriterX() WriteQuerierX
}
