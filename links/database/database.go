package database

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func ConnectWithContext(ctx context.Context, dsn string, maxConns int) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)

	return db, nil
}
