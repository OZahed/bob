// Package db provides some useful SQL database functionalities like Load balancing on master slave configuration
// Wrapping sql.DB to sqlx.DB and it makes an interface on SQL to make it easier to plugin the sql engine
//
// usage example:
//
//	leader,err := otelsql.Open(diver, conString)
//	if err != nil {
//		// do something
//	}
//
//	// set max idle connections and max connections
//	leaderX,err := db.WrapSQLX(leader) // optional
//	if err != nil {
//		// ...
//	}
//
//	var followerDBs []sql.DB // or sqlx.DB
//	for _, slaveConString := range slaveConnectionStrings {
//		follower, err := otelsql.Open(driver, conString)
//		if err != nil {
//			// ...
//		}
//
//		//optional
//		followerX,err := db.WrapSQLX(follower)
//		if err != nil {
//			// ...
//		}
//
//		followerDBs = append(followerDBs, followerX) // or follower
//	}
//
//	// when ever the loadBalancedDb is used it will automatically does the load balancing on the sql statement
//	loadBalancedDb := db.NewBalancedDB(leaderX, ...followerDBs)
package db

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/OZahed/bob/internal/helper"
)

// DB is a logical database with multiple underlying physical databases
// forming a single master multiple slaves topology.
// Reads and writes are automatically directed to the correct physical db.
type DB struct {
	pdbs  []Database  // Physical databases
	xpdbs []DatabaseX // Physical databases with sqlx extensions

	count  uint64 // Monotonically incrementing counter on each query pdbs
	countX uint64 // Monotonically incrementing counter on each query for xpdbs
}

// NewBalancedDB gets Database or DatabaseX interface, DatabaseX is a super set on Database Interface
func NewBalancedDB(master Database, slaves ...Database) Database {
	db := new(DB)

	// check is salves are compatible with DatabaseX interface
	for _, slave := range slaves {
		if sx, ok := slave.(DatabaseX); ok {
			db.xpdbs = append(db.xpdbs, sx)
		}
	}

	db.pdbs = append([]Database{master}, slaves...)

	return db
}

// Close closes all physical databases concurrently after releasing master,
// releasing any open resources.
func (db *DB) Close() error {
	// release master first
	if err := db.master().Close(); err != nil {
		return err
	}

	return helper.Scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Close()
	})
}

// Begin starts a transaction on the master. The isolation level is dependent on the driver.
func (db *DB) Begin() (*sql.Tx, error) {
	return db.master().Begin()
}

// BeginTx starts a transaction with the provided context on the master.
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.master().BeginTx(ctx, opts)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the master as the underlying physical db.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.master().Exec(query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the master as the underlying physical db.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.master().ExecContext(ctx, query, args...)
}

// Ping verifies if a connection to each physical database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	return helper.Scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Ping()
	})
}

// PingContext verifies if a connection to each physical database is still
// alive, establishing a connection if necessary.
func (db *DB) PingContext(ctx context.Context) error {
	return helper.Scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].PingContext(ctx)
	})
}

// TODO: Implement Prepare and PrepareContext
// Prepare creates a prepared statement for later queries or executions
// on each physical database, concurrently.
// func (db *DB) Prepare(query string) (Stmt, error) {
// 	stmts := make([]*sql.Stmt, len(db.pdbs))

// 	err := helper.Scatter(len(db.pdbs), func(i int) (err error) {
// 		stmts[i], err = db.pdbs[i].Prepare(query)
// 		return err
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &stmt{db: db, stmts: stmts}, nil
// }

// PrepareContext creates a prepared statement for later queries or executions
// on each physical database, concurrently.
//
// The provided context is used for the preparation of the statement, not for
// the execution of the statement.
// func (db *DB) PrepareContext(ctx context.Context, query string) (Stmt, error) {
// 	stmts := make([]*sql.Stmt, len(db.pdbs))

// 	err := helper.Scatter(len(db.pdbs), func(i int) (err error) {
// 		stmts[i], err = db.pdbs[i].PrepareContext(ctx, query)
// 		return err
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &stmt{db: db, stmts: stmts}, nil
// }

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// Query uses a slave as the physical db.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.slave().Query(query, args...)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// QueryContext uses a slave as the physical db.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.slave().QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRow uses a slave as the physical db.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.slave().QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRowContext uses a slave as the physical db.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.slave().QueryRowContext(ctx, query, args...)
}

// Get
func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.slaveX().Get(dest, query, args...)
}

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.slaveX().Select(dest, query, args...)
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool for each underlying physical db.
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns then the
// new MaxIdleConns will be reduced to match the MaxOpenConns limit
// If n <= 0, no idle connections are retained.
func (db *DB) SetMaxIdleConns(n int) {
	for i := range db.pdbs {
		db.pdbs[i].SetMaxIdleConns(n)
	}
}

// SetMaxOpenConns sets the maximum number of open connections
// to each physical database.
// If MaxIdleConns is greater than 0 and the new MaxOpenConns
// is less than MaxIdleConns, then MaxIdleConns will be reduced to match
// the new MaxOpenConns limit. If n <= 0, then there is no limit on the number
// of open connections. The default is 0 (unlimited).
func (db *DB) SetMaxOpenConns(n int) {
	for i := range db.pdbs {
		db.pdbs[i].SetMaxOpenConns(n)
	}
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
// Expired connections may be closed lazily before reuse.
// If d <= 0, connections are reused forever.
func (db *DB) SetConnMaxLifetime(d time.Duration) {
	for i := range db.pdbs {
		db.pdbs[i].SetConnMaxLifetime(d)
	}
}

// master returns the master physical database
func (db *DB) master() Database {
	return db.pdbs[0]
}

// slave returns one of the physical databases which is a slave
func (db *DB) slave() Database {
	return db.pdbs[db.acquireSlave(len(db.pdbs))]
}

func (db *DB) slaveX() DatabaseX {
	return db.xpdbs[db.acquireSlaveX(len(db.xpdbs))]
}

func (db *DB) acquireSlaveX(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&db.countX, 1) % uint64(n-1)))
}

func (db *DB) acquireSlave(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&db.count, 1) % uint64(n-1)))
}
