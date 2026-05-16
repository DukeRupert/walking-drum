package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/dukerupert/walking-drum/internal/db"
	"github.com/dukerupert/walking-drum/internal/envfile"
)

func main() {
	// environment and db setup
	if err := envfile.Load(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// ...rest of app

}
