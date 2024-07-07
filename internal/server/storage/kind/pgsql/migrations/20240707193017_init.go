package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInit, downInit)
}

func upInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE metric
	(
		name  varchar(16)  not null primary key,
		type  varchar(256) not null,
		value double precision not null
	)
	`)
	return err
}

func downInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE metric")
	return err
}
