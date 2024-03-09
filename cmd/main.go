package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tinydates"
	"tinydates/store"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// run starts the tinydates service
func run() error {
	// setup; blocking error channel and parent context object
	// note: omitting creation of loggers and instrumentation for simplicity
	errChan := make(chan error)
	ctx := contextWithSignal(context.Background())

	// read environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
		os.Exit(1)
	}

	// database connection and initialisation
	dbPool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		fmt.Printf("unable to connect to the database: %v\n", dbPool)
		os.Exit(1)
	}
	defer dbPool.Close()
	postgresStore := store.NewTinydatesPgStore(dbPool)

	// run idempotent database migrations at start of application
	m, err := migrate.New(os.Getenv("MIGRATION_URL"), os.Getenv("POSTGRES_URL"))
	if err != nil {
		fmt.Printf("unable to connect to migration database: %v\n", err)
		os.Exit(1)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Printf("failed to run database migration: %v\n", err)
		os.Exit(1)
	}

	// cache connection and initialisation
	cachePool := redis.Pool{
		MaxIdle: 5,
		Dial: func() (redis.Conn, error) {
			return redis.DialContext(
				ctx,
				"tcp",
				os.Getenv("REDIS_URL"),
				redis.DialPassword(os.Getenv("REDIS_PASSWORD")),
			)
		},
	}
	// connection to be used as follows
	conn := cachePool.Get()
	defer conn.Close()

	// Tinydates service creation; dependency injection of db, and cache
	service := tinydates.New(postgresStore, &cachePool)

	// handler creation; dependency injection of context and service
	handler := tinydates.NewTinydatesHandler(ctx, service)

	// HTTP server creation with sane defaults for the type of service
	server := &http.Server{
		Addr: os.Getenv("PORT"),
		Handler: handler,
		IdleTimeout: 15 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// start the server in a goroutine; graceful shutdown upon a major error
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	// listen for user termination (CTRL+C) signal in goroutine; ungraceful
	// termination upon shutdown signal
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-ch)
	}()

	// run program forever until fatal error or user termination signal
	return <-errChan
}

// contextWithSignal cancels a context object when a SIGINT or SIGTERM signal
// is called in the host operating system.
func contextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			cancel()
		}
	}()
	return newCtx
}

// application entrypoint
func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
