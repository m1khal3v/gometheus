package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(driver, dsn string) (*sql.DB, error) {
	return sql.Open(driver, dsn)
}
