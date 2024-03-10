package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// tinyDatesPgStore provide access to the Store methods for a PostgreSQL
// backed database.
type tinydatesPgStore struct {
	Db *pgxpool.Pool
}

func NewTinydatesPgStore(db *pgxpool.Pool) Store {
	return &tinydatesPgStore{Db: db}
}

const (
	storeUser = `
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
		storeUser,
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

const (
	getPassword = `
        SELECT password
		FROM users
		WHERE email = $1
	`
)

func (store *tinydatesPgStore) GetPassword(
	ctx context.Context,
	email string,
) (string, error) {
	var password string

	if err := store.Db.QueryRow(
		ctx,
		getPassword,
		email,
	).Scan(
		&password,
	); err != nil {
		return "", err
	}

	return password, nil
}
