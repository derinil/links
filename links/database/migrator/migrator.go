package migrator

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type (
	Migrator struct {
		root  string
		sqlFs embed.FS
		db    *sqlx.DB
	}

	Migration struct {
		Path         string    `db:"-"`
		Name         string    `db:"name"`
		RanAt        time.Time `db:"ran_at"`
		PrefixNumber int       `db:"prefix_number"`
	}
)

func New(sqlFs embed.FS, root string, db *sqlx.DB) *Migrator {
	return &Migrator{
		db:    db,
		root:  root,
		sqlFs: sqlFs,
	}
}

func (m *Migrator) Up() error {
	return m.walkExec(func(s string) bool {
		return strings.HasSuffix(s, ".up.sql")
	})
}

func (m *Migrator) Down() error {
	return m.walkExec(func(s string) bool {
		return strings.HasSuffix(s, ".down.sql")
	})
}

func (m *Migrator) walkExec(filter func(string) bool) error {
	if err := m.migratorTable(); err != nil {
		return fmt.Errorf("failed to create migrator tracker table: %w", err)
	}

	migrations := make([]Migration, 0, 100)

	err := fs.WalkDir(m.sqlFs, m.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read dir: %w", err)
		}

		if !filter(p) {
			return nil
		}

		var (
			prefix int
			name   = d.Name()
		)

		if v := strings.Split(name, "_"); len(v) > 1 {
			prefix, err = strconv.Atoi(v[0])
			if err != nil {
				return fmt.Errorf("failed to get prefix number from name: %w", err)
			}
		}

		migrations = append(migrations, Migration{
			Path:         p,
			Name:         name,
			PrefixNumber: prefix,
			RanAt:        time.Now().UTC(),
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk dir: %w", err)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	tx := m.db.MustBeginTx(context.TODO(), nil)
	defer tx.Rollback()

	for i := range migrations {
		mig := &migrations[i]

		b, err := m.sqlFs.ReadFile(mig.Path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		exists, err := m.incrementMigration(tx, mig)
		if err != nil {
			return fmt.Errorf("failed to increment migration: %w", err)
		}

		if exists {
			continue
		}

		_, err = tx.Exec(string(b))
		if err != nil {
			return fmt.Errorf("failed to exec statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migrations: %w", err)
	}

	return nil
}

func (m *Migrator) migratorTable() error {
	const query = `create table if not exists migrator_tracker (
		prefix_number integer primary key,
		name text not null,
		ran_at timestamp not null
	)`

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create migrator tracker table: %w", err)
	}

	return nil
}

func (m *Migrator) incrementMigration(tx *sqlx.Tx, migration *Migration) (bool, error) {
	const query = `insert into
		migrator_tracker (prefix_number, name, ran_at)
		values (:prefix_number, :name, :ran_at)
	on conflict do nothing`

	r, err := tx.NamedExec(query, migration)
	if err != nil {
		return false, fmt.Errorf("failed to create migrator tracker table: %w", err)
	}

	ra, err := r.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return ra == 0, nil
}
