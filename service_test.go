package tinydates

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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
	testCache   cache.TestCache
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

	// run tear up and down scripts at start
	if err := testStore.Down(ctx); err != nil {
		fmt.Println("dslfjdklfjkldsf")
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
	req := httptest.NewRequest("", "/user/create", nil)
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
