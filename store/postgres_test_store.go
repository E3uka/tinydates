package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTestTinydatesPgStore(db *pgxpool.Pool) TestStore {
	return &tinydatesPgStore{Db: db}
}

const (
	Up = `
		BEGIN;

		CREATE TABLE IF NOT EXISTS "users" (
			"id" bigserial PRIMARY KEY,
			"email" varchar NOT NULL,
			"password" varchar NOT NULL,
			"name" varchar NOT NULL,
			"gender" varchar NOT NULL,
			"age" integer NOT NULL,
			UNIQUE(id, email)
		);

		COMMIT;
	`
)

// manual calling of database migration to create test instance
func (store *tinydatesPgStore) Up(ctx context.Context) error {
	tx, err := store.Db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if _, err = store.Db.Exec(ctx, Up); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

const (
	Down = `
		DROP TABLE IF EXISTS "users";
	`
)

func (store *tinydatesPgStore) Down(ctx context.Context) error {
	tx, err := store.Db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if _, err = store.Db.Exec(ctx, Down); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
