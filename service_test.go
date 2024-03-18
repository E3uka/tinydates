package tinydates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
	"tinydates/cache"
	"tinydates/store"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var (
	dbPool      *pgxpool.Pool
	testStore   store.TestStore
	testCache   cache.Cache
	testHandler http.Handler
	service     Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	resource, err := pool.Run(
		"postgres",
		"13-alpine",
		[]string{
			"POSTGRES_DB=tinydb",
			"POSTGRES_USER=root",
			"POSTGRES_PASSWORD=password",
		},
	)
	if err != nil {
		log.Fatalf("could not start resource: %s", err)
	}

	// set max wait time to connect to the container
	pool.MaxWait = 15 * time.Second

	// set max life of the test container
	resource.Expire(60)

	if err = pool.Retry(func() error {
		dbPool, err = pgxpool.New(
			ctx,
			fmt.Sprintf(
				"postgresql://root:password@localhost:%s/tinydb?sslmode=disable",
				resource.GetPort("5432/tcp"),
			),
		)
		return err
	}); err != nil {
		log.Fatalf("could not connect to docker: %s", err.Error())
	}
	defer dbPool.Close()

	// need to introduce a sleep here because it is too fast to connec to the
	// postgres container and needs to send a fast restart message
	time.Sleep(time.Second * 2)

	// initialise test version of the store
	testStore = store.NewTestTinydatesPgStore(dbPool)

	// create and initialise test version of the cache
	testCache = cache.NewTinydatesInMemoryCache()

	// create and wire up service
	service = New(testStore, testCache)

	// router initialisation
	testHandler = NewTinydatesHandler(ctx, service)

	// run tear up and down scripts at start, tear up to connect, reuse and
	// drop any lingering container resources
	if err := testStore.Down(ctx); err != nil {
		log.Fatalf("could not teardown the resource: %s", err)
	}
	if err := testStore.Up(ctx); err != nil {
		log.Fatalf("could not create the resource: %s", err)
	}

	// run tests below
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge the resource: %s", err)
	}

	os.Exit(code)
}

func TestNewUserCreation(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest("GET", "/user/create", nil)
	req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	var user User
	err := json.NewDecoder(resp.Body).Decode(&user)

	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestUserLogin(t *testing.T) {
	ctx := context.Background()
	// create a new user
	user, err := service.CreateUser(ctx)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		testCase LoginRequest
	}{
		{"success", LoginRequest{user.Email, user.Password}},
		{"failure email not found", LoginRequest{"invalid email", user.Password}},
		{"failure password incorrect", LoginRequest{user.Email, "invalid password"}},
	}

	for _, tc := range testCases {
		if tc.name == "success" {
			jsonRequest, err := json.Marshal(tc.testCase)
			require.NoError(t, err)
			reader := bytes.NewReader(jsonRequest)
			req := httptest.NewRequest("POST", "/login", reader)
			req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			testHandler.ServeHTTP(rec, req)
			resp := rec.Result()

			var success LoginResponse
			err = json.NewDecoder(resp.Body).Decode(&success)

			require.Equal(t, 201, resp.StatusCode)
			require.NoError(t, err)
			require.NotNil(t, success)
		} else {
			jsonRequest, err := json.Marshal(tc.testCase)
			require.NoError(t, err)
			reader := bytes.NewReader(jsonRequest)
			req := httptest.NewRequest("POST", "/login", reader)
			req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			testHandler.ServeHTTP(rec, req)
			resp := rec.Result()

			require.Equal(t, 500, resp.StatusCode)
		}
	}
}

func TestUserDiscoveryAndResultOrder(t *testing.T) {
	ctx := context.Background()
	user1, err := service.CreateUser(ctx)
	require.NoError(t, err)
	_, err = service.CreateUser(ctx)
	require.NoError(t, err)

	// discovery based on user 1, login to obtain loginResponse
	loginResponse, err := service.Login(ctx, LoginRequest{user1.Email, user1.Password})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/discover", nil)
	req.WithContext(ctx)
	// for simplicity not following standard Authorization: <scheme> <token>
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", loginResponse.Token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(resp.Body).Decode(&discoverResponse)

	// for simplicity not checking exact users match just status code and
	// correct number of results
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, err)
	require.Equal(t, 3, len(discoverResponse.Results))
	// check results are ordered by closest to furthest; for simplicity and to
	// account for randomness three users should be enough for confirmation
	require.LessOrEqual(
		t,
		discoverResponse.Results[0].DistanceFromMe,
		discoverResponse.Results[1].DistanceFromMe,
	)
	require.LessOrEqual(
		t,
		discoverResponse.Results[1].DistanceFromMe,
		discoverResponse.Results[2].DistanceFromMe,
	)
}

func TestUserDiscoveryByAge(t *testing.T) {
	ctx := context.Background()
	// create new users
	user1, err := service.CreateUser(ctx)
	require.NoError(t, err)
	_, err = service.CreateUser(ctx)
	require.NoError(t, err)

	// discovery based on user 1, login to obtain loginResponse
	loginResponse, err := service.Login(ctx, LoginRequest{user1.Email, user1.Password})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/discover?minAge=20&maxAge=40", nil)
	req.WithContext(ctx)
	// for simplicity not following standard Authorization: <scheme> <token>
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", loginResponse.Token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(resp.Body).Decode(&discoverResponse)

	// for simplicity not checking exact users match just status code and
	// correct number of results
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, err)
	for _, found := range discoverResponse.Results {
		require.GreaterOrEqual(t, found.Age, 20)
		require.LessOrEqual(t, found.Age, 40)
	}
}

func TestUserDiscoveryByPopularity(t *testing.T) {
	ctx := context.Background()
	// create new users
	user1, err := service.CreateUser(ctx)
	require.NoError(t, err)
	_, err = service.CreateUser(ctx)
	require.NoError(t, err)

	// discovery based on user 1, login to obtain loginResponse
	loginResponse, err := service.Login(ctx, LoginRequest{user1.Email, user1.Password})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/discover?orderByPopularity=true", nil)
	req.WithContext(ctx)
	// for simplicity not following standard Authorization: <scheme> <token>
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", loginResponse.Token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(resp.Body).Decode(&discoverResponse)

	// for simplicity not checking exact users match just status code and
	// correct number of results
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, err)
	require.Equal(t, 7, len(discoverResponse.Results))
	// check results are ordered by Popularity;
	require.GreaterOrEqual(
		t,
		discoverResponse.Results[0].Popularity,
		discoverResponse.Results[1].Popularity,
	)
	require.GreaterOrEqual(
		t,
		discoverResponse.Results[1].Popularity,
		discoverResponse.Results[2].Popularity,
	)
	require.GreaterOrEqual(
		t,
		discoverResponse.Results[2].Popularity,
		discoverResponse.Results[3].Popularity,
	)
	require.GreaterOrEqual(
		t,
		discoverResponse.Results[3].Popularity,
		discoverResponse.Results[4].Popularity,
	)
}

func TestInvalidQueryUserDiscovery(t *testing.T) {
	ctx := context.Background()
	// create new users
	user1, err := service.CreateUser(ctx)
	require.NoError(t, err)

	// discovery based on user 1, login to obtain loginResponse
	loginResponse, err := service.Login(ctx, LoginRequest{user1.Email, user1.Password})
	require.NoError(t, err)

	// invalid query parameter - minAge is greater than maxAge
	req := httptest.NewRequest("GET", "/discover?minAge=30&maxAge=26", nil)
	req.WithContext(ctx)
	// for simplicity not following standard Authorization: <scheme> <token>
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", loginResponse.Token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(resp.Body).Decode(&discoverResponse)

	// for simplicity not checking exact users match just status code and
	// correct number of results
	require.Equal(t, 400, resp.StatusCode)
	require.NoError(t, err)
}

func TestUserSwipes(t *testing.T) {
	ctx := context.Background()
	// create new users
	user1, err := service.CreateUser(ctx)
	require.NoError(t, err)
	user2, err := service.CreateUser(ctx)
	require.NoError(t, err)

	// login both users to obtain tokens
	user1Login, err := service.Login(ctx, LoginRequest{user1.Email, user1.Password})
	require.NoError(t, err)
	user2Login, err := service.Login(ctx, LoginRequest{user2.Email, user2.Password})
	require.NoError(t, err)

	// create swipe requests for user1 and user2 to send
	user1SwipeRequest := SwipeRequest{SwiperId: user1.Id, SwipeeId: user2.Id, Decision: true}
	user2SwipeRequest := SwipeRequest{SwiperId: user2.Id, SwipeeId: user1.Id, Decision: true}

	// marshall each request and create readers ready to send
	user1Json, err := json.Marshal(user1SwipeRequest)
	require.NoError(t, err)
	user2Json, err := json.Marshal(user2SwipeRequest)
	require.NoError(t, err)
	user1Reader := bytes.NewReader(user1Json)
	user2Reader := bytes.NewReader(user2Json)

	// send the first swipe request
	req := httptest.NewRequest("POST", "/swipe", user1Reader)
	req.WithContext(ctx)
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", user1Login.Token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp := rec.Result()

	// decode the response and check that matchId is not present primarily
	var user1SwipeResponse SwipeResponse
	err = json.NewDecoder(resp.Body).Decode(&user1SwipeResponse)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, false, user1SwipeResponse.Matched)
	require.Empty(t, user1SwipeResponse.MatchId)

	// send the second swipe request
	req = httptest.NewRequest("POST", "/swipe", user2Reader)
	req.WithContext(ctx)
	req.Header.Set("Id", strconv.Itoa(user1.Id))
	req.Header.Set("Authorization", user2Login.Token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	testHandler.ServeHTTP(rec, req)
	resp = rec.Result()

	// decode the response and check that matchId is not present primarily
	var user2SwipeResponse SwipeResponse
	err = json.NewDecoder(resp.Body).Decode(&user2SwipeResponse)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, true, user2SwipeResponse.Matched)
	require.NotEmpty(t, user2SwipeResponse.MatchId)
}
