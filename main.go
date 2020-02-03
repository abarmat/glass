package main

import (
	"log"

	"github.com/go-pg/pg/v9"

	"melonwave.com/glass/api"
	"melonwave.com/glass/core"
)

//TODO: use logrus

func connectDB(databaseURL string) (*pg.DB, error) {
	opts, err := pg.ParseURL(databaseURL)
	if err != nil {
		return nil, err
	}
	return pg.Connect(opts), nil
}

func main() {
	opts, err := core.NewOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := connectDB(opts.DatabaseURL)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	client := api.NewClient(opts.ContentServerURL)

	indexes := []core.Index{&core.SceneIndex{}}
	indexer := core.NewContentIndexer(client, db, &indexes, opts.IndexWorkers)
	indexer.Run()
}
