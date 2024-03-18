package store

import "context"

// Store are the core methods required from the database for tinydates.
type Store interface {
	// StoreNewUser inserts a new user into the database returning their
	// autoincremented user id as the value.
	StoreNewUser(
		ctx context.Context,
		email, password, name, gender string,
		age, location int,
	) (int, error)

	// GetPassword returns the password for the user with the supplied email 
	GetPassword(ctx context.Context, email string) (string, error)

	// Discover finds profiles that are a match for the user with supplied id
	Discover(ctx context.Context, id int) ([]PotentialMatch, error)

	// DiscoverWithPopulariy finds potential profiles that are a match for the 
	// user supplied id and ordered by popularity
	DiscoverByPopularity(
		ctx context.Context,
		id int,
	) ([]PotentialMatch, error)

	// Swipe adds a swipe decision for the swiper and returns the match id and
	// whether the swiper has also been favourably swiped by the swipee
	Swipe(
		ctx context.Context,
		swiperId, swipeeId int,
		decision bool,
	) (int, error)

	// IsMatch returns whether the swipee has been also been favourably swiped
	// by the swiper
	IsMatch(ctx context.Context, swipeeId, swiperId int) (bool, error)

	// GetLocation returns the location for the user with the supplied id
	GetLocation(ctx context.Context, id int) (int, error)
}

// TestStore are the test methods used for testing the tinydates database.
// These embed the store methods to get access to all its methods but provide
// database creation and destruction methods for provisioning test instances.
//
// Safety: This is a whole separate interface so that the 'production' database
// does not accidentally get dropped.
type TestStore interface {
	Store

	// Up is a database creation method for testing only
	Up(ctx context.Context) error

	// Down is a database destruction method for testing only
	Down(ctx context.Context) error
}
