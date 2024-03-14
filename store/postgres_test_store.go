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
        -- migration 000001
		CREATE TABLE IF NOT EXISTS "users" (
			"id" bigserial PRIMARY KEY,
			"email" varchar NOT NULL,
			"password" varchar NOT NULL,
			"name" varchar NOT NULL,
			"gender" varchar NOT NULL,
			"age" integer NOT NULL,
			UNIQUE(id, email)
		);

        -- migration 000002
		CREATE TABLE IF NOT EXISTS "swipes" (
			"id" bigserial PRIMARY KEY,
			"swiper" integer NOT NULL,
			"swipee" integer NOT NULL,
			"decision" boolean NOT NULL
		);

        -- migration 000003
		ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "location" integer NOT NULL DEFAULT 0;
	`
)

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
        -- migration 000003
		ALTER TABLE IF EXISTS "users" DROP COLUMN "location";

        -- migration 000002
		DROP TABLE IF EXISTS "swipes";

        -- migration 000001
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
