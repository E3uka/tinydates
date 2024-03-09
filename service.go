package tinydates

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"tinydates/store"

	"github.com/gomodule/redigo/redis"
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
)

// Service interface encapsulates all functionalities of the tinydates service
type Service interface {
	// CreateUser creates and stores a new user in the system.
	CreateUser(ctx context.Context) (User, error)
}

type tinydates struct {
	store store.Store
	cache *redis.Pool
}

func New(store store.Store, cache *redis.Pool) Service {
	return tinydates{store: store, cache: cache}
}

func (td tinydates) CreateUser(ctx context.Context) (User, error) {
	randomName := createRandomName(maxNameLength)
	randomEmail := fmt.Sprintf("%v@mail.com", randomName)
	// returning password as plaintext is no bueno
	randomPassword := createRandomName(maxNameLength)
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

func createRandomName(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
