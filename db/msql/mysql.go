package msql

import (
	//nolint:revive
	"fmt"
	"strings"

	//nolint:revive
	_ "github.com/go-sql-driver/mysql"
)

type MySQLConnStringProvider struct {
	Host         string
	Port         int
	Proto        string
	User         string
	Password     string
	DatabaseName string
}

func (s *MySQLConnStringProvider) Name() string {
	return "mysql"
}

func (s *MySQLConnStringProvider) ConnectionString() string {
	if strings.TrimSpace(s.Proto) == "" {
		s.Proto = "tcp"
	}

	return fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		s.User, s.Password, s.Proto, s.Host, s.Port, s.DatabaseName,
	)
}

func (s *MySQLConnStringProvider) DBName() string {
	return s.DatabaseName
}
