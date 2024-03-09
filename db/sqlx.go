package db

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrNotSQLCompatible = errors.New("db is not sql.DB compatible")

type sqlxDB struct {
	*sqlx.DB
}

func WrapSQLX(db Database, driverName string) (DatabaseX, error) {
	if dbc, ok := db.(*sql.DB); ok {
		return &sqlxDB{sqlx.NewDb(dbc, driverName)}, nil
	}

	return nil, ErrNotSQLCompatible
}
