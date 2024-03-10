package store

import "context"

// Store are the core methods required from the database for tinydates.
type Store interface {
	// StoreNewUser inserts a new user into the database returning their
	// autoincremented user id as the value.
	StoreNewUser(
		ctx context.Context,
		email, password, name, gender string,
		age int,
	) (int, error)

	// GetPassword returns the password for the user with the supplied email 
	GetPassword(ctx context.Context, email string) (string, error)
}

// TestStore are the test methods used for testing the tinydates database.
// These embed the store methods but provide database creation and destruction
// methods for provisioning test instances.
type TestStore interface {
	Store

	// Up is a database creation method for testing only
	Up(ctx context.Context) error

	// Down is a database destruction method for testing only
	Down(ctx context.Context) error
}
