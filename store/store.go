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

// tinyDatesPgStore wraps an initialised context and database connection pool
// object to provide access to the Store methods
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

