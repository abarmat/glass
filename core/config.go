package core

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

// Options contains all the app config vars
type Options struct {
	DatabaseURL      string
	ContentServerURL string
	DataDir          string
	IndexWorkers     int
	IndexInterval    int
}

// NewOptionsFromEnv returns app options based on env
func NewOptionsFromEnv() (Options, error) {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	opts := Options{}
	opts.DatabaseURL = os.Getenv("DATABASE_URL")
	opts.ContentServerURL = os.Getenv("CONTENT_SERVER_URL")
	opts.DataDir = os.Getenv("DATA_DIR")
	opts.IndexWorkers, err = strconv.Atoi(os.Getenv("INDEX_WORKERS"))
	opts.IndexInterval, err = strconv.Atoi(os.Getenv("INDEX_INTERVAL"))

	return opts, err
}
