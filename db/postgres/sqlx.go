package postgres

import (
	"database/sql"
	"errors"

	"github.com/OZahed/bob/db"
	"github.com/jmoiron/sqlx"
)

var ErrNotSQLCompatible = errors.New("db is not sql.DB compatible")

type sqlxDB struct {
	*sqlx.DB
}

func WrapSQLX(db db.Database) (db.DatabaseX, error) {
	if dbc, ok := db.(*sql.DB); ok {
		return &sqlxDB{sqlx.NewDb(dbc, driverName)}, nil
	}

	return nil, ErrNotSQLCompatible
}
