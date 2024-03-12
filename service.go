package tinydates

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
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

	// ErrMinOrMaxAgeMissing is returned when a request is invalid
	ErrMinOrMaxAgeMissing = errors.New("error min age or max age not supplied")

	// ErrMinOrMaxAgeInvalid is returned when the supplied min or max age is not
	// an integer
	ErrMinOrMaxAgeInvalid = errors.New(
		"error min age or max age can only be an integer",
	)

	// ErrorMinOrMaxFormat is returned when min age is not less than max age
	ErrorMinOrMaxFormat = errors.New("error min age must be less than max age")
)

// Service interface encapsulates all functionalities of the tinydates service
type Service interface {
	// CreateUser creates and stores a new user in the system.
	CreateUser(ctx context.Context) (User, error)

	// Login logs a user into the system by means of an entry into the cache.
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)

	// Discover finds profiles that are a match for the user with supplied id.
	Discover(
		ctx context.Context,
		id int,
		token string,
		minAge string,
		minAgeSupplied bool,
		maxAge string,
		maxAgeSupplied bool,
	) (DiscoverResponse, error)

	// Swipe handles the action when a user swipes on a discovered profile.
	Swipe(
		ctx context.Context,
		token string,
		req SwipeRequest,
	) (SwipeResponse, error)
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
	minAge string,
	minAgeSupplied bool,
	maxAge string,
	maxAgeSupplied bool,
) (DiscoverResponse, error) {
	if !td.cache.Authorized(ctx, token) {
		return DiscoverResponse{}, ErrUnauthorized
	}

	// for simplicity enforcing min and max age supplied
	if minAgeSupplied == true && maxAgeSupplied == false ||
		minAgeSupplied == false && maxAgeSupplied == true {
		return DiscoverResponse{}, ErrMinOrMaxAgeMissing
	}

	var profiles []store.PotentialMatch

	if minAgeSupplied {
		minAgeInt, err := strconv.Atoi(minAge)
		if err != nil {
			return DiscoverResponse{}, ErrMinOrMaxAgeInvalid
		}

		maxAgeInt, err := strconv.Atoi(maxAge)
		if err != nil {
			return DiscoverResponse{}, ErrMinOrMaxAgeInvalid
		}

		if minAgeInt > maxAgeInt {
			return DiscoverResponse{}, ErrorMinOrMaxFormat
		}

		foundProfiles, err := td.store.DiscoverWithAge(ctx, id, minAgeInt, maxAgeInt)
		if err != nil {
			return DiscoverResponse{}, ErrInternalService
		}

		profiles = foundProfiles
	} else {
		foundProfiles, err := td.store.Discover(ctx, id)
		if err != nil {
			return DiscoverResponse{}, ErrInternalService
		}

		profiles = foundProfiles
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

func (td tinydates) Swipe(
	ctx context.Context,
	token string,
	req SwipeRequest,
) (SwipeResponse, error) {
	if !td.cache.Authorized(ctx, token) {
		return SwipeResponse{}, ErrUnauthorized
	}

	matchId, match, err := td.store.Swipe(
		ctx,
		req.SwiperId,
		req.SwipeeId,
		req.Decision,
	)

	if err != nil {
		return SwipeResponse{}, ErrInternalService
	}

	// only a match if the swiper also swiped favourably
	if match && req.Decision == true {
		return SwipeResponse{Matched: match, MatchId: matchId}, nil
	} else {
		return SwipeResponse{Matched: match}, nil
	}
}

func createRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
