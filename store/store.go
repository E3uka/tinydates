package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store are the core methods required from the database for tinydates.
type Store interface {
	// StoreNewUser inserts a new user into the database returning their
	// autoincremented user id as the value.
	StoreNewUser(
		ctx context.Context, 
		email, password, name, gender string,
		age int,
	) (int, error)
}

// tinyDatesPgStore provide access to the Store methods for a PostgreSQL
// backed database.
type tinydatesPgStore struct {
	Db *pgxpool.Pool
}

func NewTinydatesPgStore(db *pgxpool.Pool) Store {
	return &tinydatesPgStore{Db: db}
}

const (
	StoreUser = `
        INSERT INTO users (email, password, name, gender, age)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
)

func (store *tinydatesPgStore) StoreNewUser(
	ctx context.Context,
	email,
	password,
	name,
	gender string,
	age int,
) (int, error) {
	var id int

	if err := store.Db.QueryRow(
		ctx,
		StoreUser,
		email,
		password,
		name,
		gender,
		age,
	).Scan(
		&id,
	); err != nil {
		return 0, err
	}
	return id, nil
}

/*****************************

TEST METHODS ONLY
should not be called in service

*****************************/

type TestStore interface {
	Store

	// Up is a database creation method for testing only
	Up(ctx context.Context) error

	// Down is a database destruction method for testing only
	Down(ctx context.Context) error
}

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
