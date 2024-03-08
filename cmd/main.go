package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	httpAddr = flag.String("http", ":8080", "http listen address")
)

// run starts the tinydates service
func run() error {
	// setup; parse flags, blocking error channel and parent context object
	flag.Parse()
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
		fmt.Printf("Unable to connect to the database: %v\n", dbPool)
		os.Exit(1)
	}
	defer dbPool.Close()

	// cache connection and initialisation
	cachePool := redis.Pool{
		MaxIdle: 5,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				os.Getenv("REDIS_URL"),
				redis.DialPassword(os.Getenv("REDIS_PASSWORD")),
			)
		},
	}
	// connection to be used as follows
	conn := cachePool.Get()
	defer conn.Close()

	// listen for user termination (CTRL+C) signal in goroutine. Ungraceful
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

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
