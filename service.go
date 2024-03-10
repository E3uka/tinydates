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
	letters           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	maxNameLength     = 20
	maxPasswordLength = 30
	maxAge            = 120
)

var gender = []string{"male", "female", "other"}

var (
	// ErrCreateUser is returned when there is an error creating a new user
	ErrCreateUser = errors.New("error creating new user")

	// ErrInvalidPassword is returned when supplied and found passwords differs
	ErrInvalidPassword = errors.New("error invalid password")
)

// Service interface encapsulates all functionalities of the tinydates service
type Service interface {
	// CreateUser creates and stores a new user in the system.
	CreateUser(ctx context.Context) (User, error)

	// Login logs a user into the system by means of an entry into the cache
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
}

type tinydates struct {
	store store.Store
	cache cache.Cache
}

func New(store store.Store, cache cache.Cache) Service {
	return tinydates{store: store, cache: cache}
}

func (td tinydates) CreateUser(ctx context.Context) (User, error) {
	randomName := createRandomString(maxNameLength)
	randomEmail := fmt.Sprintf("%v@mail.com", randomName)
	// returning password as plaintext is no bueno
	randomPassword := createRandomString(maxNameLength)
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
		return LoginResponse{}, err
	}

	if req.Password == storedPassword {
		// new token created in same way as name for simplicity
		token := createRandomString(maxNameLength)
		td.cache.StartSession(ctx, token)
		return LoginResponse{Token: token}, err
	} else {
		return LoginResponse{}, ErrInvalidPassword
	}
}

func createRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
