package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/derinil/links/links/account"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountReader struct {
	db         *sqlx.DB
	linkReader *LinkReader
}

func NewAccountReader(db *sqlx.DB) *AccountReader {
	return &AccountReader{
		db:         db,
		linkReader: NewLinkReader(db),
	}
}

func (s *AccountReader) Get(ctx context.Context, cmd *account.GetCmd) (*account.Account, error) {
	b := builder.Select("*").From("accounts")

	if cmd.Handle != "" {
		b = b.Where(squirrel.Eq{"handle": cmd.Handle})
	}

	if cmd.ID != uuid.Nil {
		b = b.Where(squirrel.Eq{"id": cmd.ID})
	}

	q, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var a account.Account
	if err := s.db.GetContext(ctx, &a, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get account by handle: %w", err)
	}

	if !cmd.Shallow {
		ls, err := s.linkReader.ListLinksByAccountID(ctx, a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list links by account id: %w", err)
		}

		a.Links = ls
	}

	if err := a.AfterLoad(); err != nil {
		return nil, fmt.Errorf("failed to run after load on account: %w", err)
	}

	return &a, nil
}

type AccountWriter struct {
	db         *sqlx.DB
	linkWriter *LinkWriter
}

func NewAccountWriter(db *sqlx.DB) *AccountWriter {
	return &AccountWriter{
		db:         db,
		linkWriter: NewLinkWriter(db),
	}
}

func (s *AccountWriter) SaveAccount(ctx context.Context, a *account.Account) error {
	const query = `insert into
		accounts (id, name, handle, password, avi, css, inserted_at, updated_at)
		values (:id, :name, :handle, :password, :avi, :css, :inserted_at, :updated_at)
	on conflict (id) do update set
		name = :name,
		handle = :handle,
		avi = :avi,
		css = :css,
		updated_at = :updated_at`

	if err := a.BeforeSave(); err != nil {
		return fmt.Errorf("failed to run before save on account: %w", err)
	}

	tx := s.db.MustBeginTx(ctx, nil)
	defer tx.Rollback()

	for i := range a.Links {
		if err := s.linkWriter.SaveLinkWithTx(ctx, tx, &a.Links[i]); err != nil {
			return fmt.Errorf("failed to save link: %w", err)
		}
	}

	if _, err := tx.NamedExecContext(ctx, query, a); err != nil {
		return fmt.Errorf("failed to insert account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
