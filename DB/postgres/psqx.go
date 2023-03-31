package postgres

import (
	"context"
	"database/sql"
	"sync/atomic"

	db "github.com/OZahed/scratch/DB"
	"github.com/jmoiron/sqlx"
)

var (
	_ db.ReadQuerierX  = (*Reader)(nil)
	_ db.WriteQuerierX = (*Writer)(nil)
)

// Writer is the writer implementation of ReadQuerierX, hence the ReadQuerier
type Reader struct {
	dbConns []sqlx.DB
	count   *uint32
}

func (r *Reader) getIdx() int {
	return (int(atomic.AddUint32(r.count, 1)) % len(r.dbConns))
}

// Prepare implements db.ReadQuerierX
func (r *Reader) Prepare(query string) (*sql.Stmt, error) {
	return r.dbConns[r.getIdx()].Prepare(query)
}

// PrepareContext implements db.ReadQuerierX
func (r *Reader) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return r.dbConns[r.getIdx()].PrepareContext(ctx, query)
}

// Query implements db.ReadQuerierX
func (r *Reader) Query(query string, args ...any) (*sql.Rows, error) {
	return r.dbConns[r.getIdx()].Query(query, args...)
}

// QueryContext implements db.ReadQuerierX
func (r *Reader) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.dbConns[r.getIdx()].QueryContext(ctx, query, args...)
}

// QueryRow implements db.ReadQuerierX
func (r *Reader) QueryRow(query string, args ...any) *sql.Row {
	return r.dbConns[r.getIdx()].QueryRow(query, args...)
}

// QueryRowContext implements db.ReadQuerierX
func (r *Reader) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return r.dbConns[r.getIdx()].QueryRowContext(ctx, query, args...)
}

// Get implements db.ReadQuerierX
func (r *Reader) Get(dest interface{}, query string, args ...interface{}) error {
	return r.dbConns[r.getIdx()].Get(dest, query, args...)
}

// GetContext implements db.ReadQuerierX
func (r *Reader) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return r.dbConns[r.getIdx()].GetContext(ctx, dest, query, args...)
}

// NamedQuery implements db.ReadQuerierX
func (r *Reader) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return r.dbConns[r.getIdx()].NamedQuery(query, arg)
}

// NamedQueryContext implements db.ReadQuerierX
func (r *Reader) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return r.dbConns[r.getIdx()].NamedQueryContext(ctx, query, arg)
}

// PrepareNamed implements db.ReadQuerierX
func (r *Reader) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return r.dbConns[r.getIdx()].PrepareNamed(query)
}

// PrepareNamedContext implements db.ReadQuerierX
func (r *Reader) PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	return r.dbConns[r.getIdx()].PrepareNamedContext(ctx, query)
}

// Preparex implements db.ReadQuerierX
func (r *Reader) Preparex(query string) (*sqlx.Stmt, error) {
	return r.dbConns[r.getIdx()].Preparex(query)
}

// PreparexContext implements db.ReadQuerierX
func (r *Reader) PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error) {
	return r.dbConns[r.getIdx()].PreparexContext(ctx, query)
}

// QueryRowx implements db.ReadQuerierX
func (r *Reader) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return r.dbConns[r.getIdx()].QueryRowx(query, args...)
}

// QueryRowxContext implements db.ReadQuerierX
func (r *Reader) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return r.dbConns[r.getIdx()].QueryRowxContext(ctx, query, args...)
}

// Queryx implements db.ReadQuerierX
func (r *Reader) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return r.dbConns[r.getIdx()].Queryx(query, args...)
}

// QueryxContext implements db.ReadQuerierX
func (r *Reader) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return r.dbConns[r.getIdx()].QueryxContext(ctx, query, args...)
}

// Select implements db.ReadQuerierX
func (r *Reader) Select(dest interface{}, query string, args ...interface{}) error {
	return r.dbConns[r.getIdx()].Select(dest, query, args...)

}

// SelectContext implements db.ReadQuerierX
func (r *Reader) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return r.dbConns[r.getIdx()].SelectContext(ctx, dest, query, args...)

}

// Writer is the writer implementation of WriteQuerierX, hence the WriteQuerier
type Writer struct {
	dbConn sqlx.DB
}

// Begin implements db.WriteQuerierX
func (w *Writer) Begin() (*sql.Tx, error) {
	return w.dbConn.Begin()
}

// BeginTx implements db.WriteQuerierX
func (w *Writer) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.dbConn.BeginTx(ctx, opts)
}

// Exec implements db.WriteQuerierX
func (w *Writer) Exec(query string, args ...any) (sql.Result, error) {
	return w.dbConn.Exec(query, args...)
}

// ExecContext implements db.WriteQuerierX
func (w *Writer) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.dbConn.ExecContext(ctx, query, args...)
}

// Prepare implements db.WriteQuerierX
func (w *Writer) Prepare(query string) (*sql.Stmt, error) {
	return w.dbConn.Prepare(query)
}

// PrepareContext implements db.WriteQuerierX
func (w *Writer) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.dbConn.PrepareContext(ctx, query)
}

// BeginTxx implements db.WriteQuerierX
func (w *Writer) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return w.dbConn.BeginTxx(ctx, opts)
}

// Beginx implements db.WriteQuerierX
func (w *Writer) Beginx() (*sqlx.Tx, error) {
	return w.dbConn.Beginx()
}

// MustBegin implements db.WriteQuerierX
func (w *Writer) MustBegin() *sqlx.Tx {
	return w.dbConn.MustBegin()
}

// MustBeginTx implements db.WriteQuerierX
func (w *Writer) MustBeginTx(ctx context.Context, opts *sql.TxOptions) *sqlx.Tx {
	return w.dbConn.MustBeginTx(ctx, opts)
}

// MustExec implements db.WriteQuerierX
func (w *Writer) MustExec(query string, args ...interface{}) sql.Result {
	return w.dbConn.MustExec(query, args...)
}

// MustExecContext implements db.WriteQuerierX
func (w *Writer) MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result {
	return w.dbConn.MustExecContext(ctx, query, args...)
}

// NamedExec implements db.WriteQuerierX
func (w *Writer) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return w.dbConn.NamedExec(query, arg)
}

// NamedExecContext implements db.WriteQuerierX
func (w *Writer) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return w.dbConn.NamedExecContext(ctx, query, arg)
}

type Database struct {
	Primaries   *Writer
	Secondaries *Reader
}
