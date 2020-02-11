package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/go-pg/pg/v9"

	"melonwave.com/glass/api"
	"melonwave.com/glass/core"
)

func initLog() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
}

func connectDB(databaseURL string) (*pg.DB, error) {
	opts, err := pg.ParseURL(databaseURL)
	if err != nil {
		return nil, err
	}
	return pg.Connect(opts), nil
}

func main() {
	initLog()

	// parse opts
	opts, err := core.NewOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// connect db
	db, err := connectDB(opts.DatabaseURL)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// listen for term signal
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		select {
		case <-c:
			log.Info("Shutting down gracefully...")
			cancel()
		}
	}()

	// start indexing
	client := api.NewClient(opts.ContentServerURL)
	indexes := []core.Index{&core.SceneIndex{}}
	indexer := core.NewContentIndexer(client, db, &indexes, opts.IndexWorkers, opts.IndexInterval)
	indexer.Run(ctx)
}
