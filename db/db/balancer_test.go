package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// createMockDB initializes a new sql.DB and its sqlmock instance.
func createMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

// slowDB is a wrapper around *sql.DB that simulates a delay on ExecContext.
type slowDB struct {
	*sql.DB
}

func (s *slowDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	time.Sleep(15 * time.Millisecond)
	return s.DB.ExecContext(ctx, query, args...)
}

func TestExecUsesMaster(t *testing.T) {
	masterDB, masterMock := createMockDB(t)
	slaveDB, _ := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	// Expect an Exec on the master.
	masterMock.
		ExpectExec(regexp.QuoteMeta("INSERT INTO test VALUES (1)")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a balanced DB with the master and slave.
	// (Assumes NewBalancedDB accepts interfaces compatible with *sql.DB.)
	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	_, err := balanced.Exec("INSERT INTO test VALUES (1)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := masterMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestQueryUsesSlave(t *testing.T) {
	masterDB, _ := createMockDB(t)
	slaveDB, slaveMock := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	// Set expectation on the slave for a Query.
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	slaveMock.
		ExpectQuery(regexp.QuoteMeta("SELECT * FROM test")).
		WillReturnRows(rows)

	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	res, err := balanced.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res.Close()

	if err := slaveMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSlowQueryLoggingSimulation(t *testing.T) {
	// For simulating a slow query, wrap the master with slowDB.
	realMaster, masterMock := createMockDB(t)
	master := &slowDB{realMaster}
	slaveDB, _ := createMockDB(t)
	defer master.DB.Close()
	defer slaveDB.Close()

	threshold := 10 * time.Millisecond
	// In a real test, you'd capture log output here.
	balanced := NewBalancedDB(threshold, nil, master, slaveDB)

	// Expect Exec on the master.
	masterMock.
		ExpectExec(regexp.QuoteMeta("UPDATE test SET col = 2")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := balanced.Exec("UPDATE test SET col = 2")
	if err != nil {
		t.Fatalf("unexpected error during Exec: %v", err)
	}

	if err := masterMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}

	// In a full test you might capture log output to verify that slow query logging occurred.
}

func TestQueryContextUsesSlave(t *testing.T) {
	masterDB, _ := createMockDB(t)
	slaveDB, slaveMock := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	slaveMock.
		ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE name = ?")).
		WithArgs("Alice").
		WillReturnRows(rows)

	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	res, err := balanced.QueryContext(context.Background(), "SELECT id FROM users WHERE name = ?", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Close()

	if err := slaveMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestExecContextDoesNotUsesSlave(t *testing.T) {
	masterDB, masterMock := createMockDB(t)
	slaveDB, slave1Mock := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	masterMock.
		ExpectExec(regexp.QuoteMeta("INSERT INTO users (name) VALUES (?)")).
		WithArgs("Alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	_, err := balanced.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := masterMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}

	if err := slave1Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestQueryRowUsesSlave(t *testing.T) {
	expectedCount := 5
	masterDB, _ := createMockDB(t)
	slaveDB, slaveMock := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)
	slaveMock.
		ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM products")).
		WillReturnRows(rows)

	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	var count int
	err := balanced.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		t.Fatalf("unexpected error during Scan: %v", err)
	}

	if count != expectedCount {
		t.Errorf("expected count 5, got %d", count)
	}

	if err := slaveMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestQueryRowContextUsesSlave(t *testing.T) {
	masterDB, _ := createMockDB(t)
	slaveDB, slaveMock := createMockDB(t)
	defer masterDB.Close()
	defer slaveDB.Close()

	rows := sqlmock.NewRows([]string{"name"}).AddRow("Bob")
	slaveMock.
		ExpectQuery(regexp.QuoteMeta("SELECT name FROM users WHERE id = ?")).
		WithArgs(10).
		WillReturnRows(rows)

	balanced := NewBalancedDB(0, nil, masterDB, slaveDB)
	var name string
	err := balanced.QueryRowContext(context.Background(), "SELECT name FROM users WHERE id = ?", 10).Scan(&name)
	if err != nil {
		t.Fatalf("unexpected error during Scan: %v", err)
	}

	if name != "Bob" {
		t.Errorf("expected name Bob, got %s", name)
	}

	if err := slaveMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMultipleSlavesRoundRobin(t *testing.T) {
	masterDB, masterMock := createMockDB(t)
	slave1DB, slave1Mock := createMockDB(t)
	slave2DB, slave2Mock := createMockDB(t)
	defer masterDB.Close()
	defer slave1DB.Close()
	defer slave2DB.Close()

	query := "SELECT data FROM logs WHERE id = ?"
	// Mock rows for the slaves
	rows1 := sqlmock.NewRows([]string{"data"}).AddRow("log1")
	rows2 := sqlmock.NewRows([]string{"data"}).AddRow("log2")
	rows3 := sqlmock.NewRows([]string{"data"}).AddRow("log3")
	rows4 := sqlmock.NewRows([]string{"data"}).AddRow("log4")

	// Expect queries to alternate between slaves

	slave2Mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnRows(rows1)
	slave1Mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(2).
		WillReturnRows(rows2)

	slave2Mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(3).
		WillReturnRows(rows3)
	slave1Mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(4).
		WillReturnRows(rows4)

	balanced := NewBalancedDB(0, nil, masterDB, slave1DB, slave2DB)

	// Execute 4 queries
	for i := 1; i <= 4; i++ {
		res, err := balanced.Query(query, i)
		if err != nil {
			t.Fatalf("unexpected error on query %d: %v", i, err)
		}

		res.Close()
	}

	masterMock.
		ExpectExec(regexp.QuoteMeta("INSERT INTO logs (data) VALUES (?)")).
		WithArgs("log5").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := balanced.Exec("INSERT INTO logs (data) VALUES (?)", "log5")
	if err != nil {
		t.Fatalf("unexpected error on insert: %v", err)
	}

	if err := slave1Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations for slave1: %s", err)
	}
	if err := slave2Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations for slave2: %s", err)
	}

	if err := masterMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations for master: %s", err)
	}
}
