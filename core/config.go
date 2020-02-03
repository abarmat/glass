package core

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Options contains all the app config vars
type Options struct {
	DatabaseURL      string
	ContentServerURL string
	IndexWorkers     int
}

// NewOptionsFromEnv returns app options based on env
func NewOptionsFromEnv() (Options, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	opts := Options{}
	opts.DatabaseURL = os.Getenv("DATABASE_URL")
	opts.ContentServerURL = os.Getenv("CONTENT_SERVER_URL")
	indexWorkers, err := strconv.Atoi(os.Getenv("NUM_WORKERS"))
	if err != nil {
		return opts, err
	}
	opts.IndexWorkers = indexWorkers

	return opts, nil
}
