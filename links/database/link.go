package database

import (
	"context"
	"fmt"

	"github.com/derinil/links/links/account"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type LinkReader struct {
	db *sqlx.DB
}

func NewLinkReader(db *sqlx.DB) *LinkReader {
	return &LinkReader{db: db}
}

func (s *LinkReader) ListLinksByAccountID(ctx context.Context, id uuid.UUID) ([]account.Link, error) {
	const query = `select * from links where account_id = $1`

	var ls []account.Link
	if err := s.db.SelectContext(ctx, &ls, query, id); err != nil {
		return nil, fmt.Errorf("failed to select links: %w", err)
	}

	for i := range ls {
		if err := ls[i].AfterLoad(); err != nil {
			return nil, fmt.Errorf("failed to run after load on link: %w", err)
		}
	}

	return ls, nil
}

type LinkWriter struct {
	db *sqlx.DB
}

func NewLinkWriter(db *sqlx.DB) *LinkWriter {
	return &LinkWriter{db: db}
}

func (s *LinkWriter) SaveLinkWithTx(ctx context.Context, tx *sqlx.Tx, l *account.Link) error {
	const query = `insert into
		links (id, account_id, title, link, favicon, index, inserted_at, updated_at)
		values (:id, :account_id, :title, :link, :favicon, :index, :inserted_at, :updated_at)
	on conflict (id) do update set
		title = :title,
		link = :link,
		favicon = :favicon,
		updated_at = :updated_at`

	if err := l.BeforeSave(); err != nil {
		return fmt.Errorf("failed to run before save on link: %w", err)
	}

	if _, err := tx.NamedExecContext(ctx, query, l); err != nil {
		return fmt.Errorf("failed to insert link: %w", err)
	}

	return nil
}
