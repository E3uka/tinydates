package tinydates

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"tinydates/cache"
	"tinydates/store"
)

const (
	letters   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	maxLength = 20
	maxAge    = 120
)

var gender = []string{"male", "female", "other"}

var (
	// ErrCreateUser is returned when there is an error creating a new user
	ErrCreateUser = errors.New("error creating new user")

	// ErrInvalidPassword is returned when supplied and found passwords differs
	ErrInvalidPassword = errors.New("error invalid password")

	// ErrUnauthorised is returned when the request is not authorised
	ErrUnauthorized = errors.New("error unathorized access")

	// ErrInternalService is returned when something major has gone wrong
	ErrInternalService = errors.New("error unathorized access")
)

// Service interface encapsulates all functionalities of the tinydates service
type Service interface {
	// CreateUser creates and stores a new user in the system.
	CreateUser(ctx context.Context) (User, error)

	// Login logs a user into the system by means of an entry into the cache.
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)

	// Discover finds profiles that are a match for the user with supplied id
	Discover(
		ctx context.Context,
		id int,
		token string,
	) (DiscoverResponse, error)
}

type tinydates struct {
	store store.Store
	cache cache.Cache
}

func New(store store.Store, cache cache.Cache) Service {
	return tinydates{store: store, cache: cache}
}

func (td tinydates) CreateUser(ctx context.Context) (User, error) {
	randomName := createRandomString(maxLength)
	randomEmail := fmt.Sprintf("%v@mail.com", randomName)
	// returning password as plaintext is no bueno
	randomPassword := createRandomString(maxLength)
	randomGender := gender[rand.Intn(len(gender))]
	randomAge := rand.Intn(maxAge)

	// store user in the database
	newId, err := td.store.StoreNewUser(
		ctx,
		randomEmail,
		randomPassword,
		randomName,
		randomGender,
		randomAge,
	)
	if err != nil {
		// for simplicity not handling the specific error from the data store
		// only returning an error back that is well formatted for the caller
		return User{}, ErrCreateUser
	}

	return User{
		Id:       newId,
		Email:    randomEmail,
		Password: randomPassword,
		Name:     randomName,
		Gender:   randomGender,
		Age:      randomAge,
	}, nil
}

func (td tinydates) Login(
	ctx context.Context,
	req LoginRequest,
) (LoginResponse, error) {
	// for brevity assuming user not logged in
	storedPassword, err := td.store.GetPassword(ctx, req.Email)
	if err != nil {
		return LoginResponse{}, ErrInternalService
	}

	if req.Password == storedPassword {
		// new token created in same way as name for simplicity
		token := createRandomString(maxLength)
		if err := td.cache.StartSession(ctx, token); err != nil {
			return LoginResponse{}, err
		}
		return LoginResponse{Token: token}, err
	} else {
		return LoginResponse{}, ErrInvalidPassword
	}
}

func (td tinydates) Discover(
	ctx context.Context,
	id int,
	token string,
) (DiscoverResponse, error) {
	if !td.cache.Authorized(ctx, token) {
		return DiscoverResponse{}, ErrUnauthorized
	}

	profiles, err := td.store.Discover(ctx, id)
	if err != nil {
		return DiscoverResponse{}, ErrInternalService
	}

	discoveredUsers := make([]DiscoveredUser, 0)

	for _, profile := range profiles {
		var usr DiscoveredUser

		usr.Id = profile.Id
		usr.Name = profile.Name
		usr.Gender = profile.Gender
		usr.Age = profile.Age

		discoveredUsers = append(discoveredUsers, usr)
	}

	return DiscoverResponse{Results: discoveredUsers}, nil
}

func createRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
