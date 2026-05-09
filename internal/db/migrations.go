package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func RunMigrations(conn *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}
	return goose.Up(conn, dir)
}
