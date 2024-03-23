package psql

import (
	//nolint:revive
	"fmt"

	//nolint:revive
	_ "github.com/lib/pq"
)

type PostgreSQLConnectionStringProvider struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSL          bool
}

func (s *PostgreSQLConnectionStringProvider) Name() string {
	return "postgres"
}

func (s *PostgreSQLConnectionStringProvider) DBName() string {
	return s.DatabaseName
}

func (s *PostgreSQLConnectionStringProvider) ConnectionString() string {
	sslMode := "disable"
	if s.SSL {
		sslMode = "enable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.User, s.Password, s.Host, s.Port, s.DatabaseName, sslMode,
	)
}

func NewPostgreSQLDriverConn() *PostgreSQLConnectionStringProvider {
	panic("not implemented")
}
