package core

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-pg/pg/v9"
	"melonwave.com/glass/api"
	"melonwave.com/glass/models"
)

// Job is request for indexing that can be run in the indexer
type Job struct {
	EntityType string
	EntityID   string
}

// Index is a runnable computation run by the indexer
type Index interface {
	GetName() string
	Run(logger *log.Entry, indexer *ContentIndexer, entityType string, entityID string) error
}

// ContentIndexer is the manager of the indexing process
type ContentIndexer struct {
	Client        *api.Client
	DB            *pg.DB
	Indexes       *[]Index
	indexWorkers  int
	indexInterval int
}

// NewContentIndexer creates a new indexer
func NewContentIndexer(
	client *api.Client,
	db *pg.DB, indexes *[]Index,
	indexWorkers int,
	indexInterval int) *ContentIndexer {
	return &ContentIndexer{client, db, indexes, indexWorkers, indexInterval}
}

func (indexer *ContentIndexer) indexEntity(logger *log.Entry, entityType string, entityID string) {
	for _, index := range *indexer.Indexes {
		logger = logger.WithFields(log.Fields{"index": index.GetName(), "id": entityID})
		err := index.Run(logger, indexer, entityType, entityID)
		if err != nil {
			logger.Error(err)
		}
	}
}

func (indexer *ContentIndexer) runWorker(id int, jobs chan Job, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := log.WithFields(log.Fields{"wk": id})
	for {
		select {
		case job, more := <-jobs:
			{
				if !more {
					return
				}
				indexer.indexEntity(logger, job.EntityType, job.EntityID)
			}
		}
	}
}

func (indexer *ContentIndexer) indexTiles(from int64) error {
	logger := log.WithField("index", "tiles")

	var scenes []models.Scene
	err := indexer.DB.Model(&scenes).
		Column("id", "pointers", "published_at").
		Order("published_at ASC").
		Where("published_at >= ?", time.Unix(from, 0)).
		Select()
	if err != nil {
		return err
	}

	log.Printf("(Tiles) Updating %d scenes tiles", len(scenes))
	for _, scene := range scenes {
		logger.WithFields(
			log.Fields{"scene": scene.ID, "n_tiles": len(scene.Pointers)},
		).Info("Updating tile")

		for _, pointer := range scene.Pointers {
			// parse pointer
			x, y, err := models.PointerToCoords(pointer)
			if err != nil {
				logger.Warn(err)
				continue
			}

			// insert or update tile
			tile := Tile{x, y, scene.ID, scene.PublishedAt, time.Now()}
			_, err = indexer.DB.Model(&tile).
				OnConflict("(x, y) DO UPDATE").
				Set("scene_id = excluded.scene_id, published_at = excluded.published_at").
				Insert()
			if err != nil {
				logger.Warn(err)
				continue
			}
		}
	}

	return nil
}

func (indexer *ContentIndexer) getLastTimestamp() (int64, error) {
	var scene models.Scene

	err := indexer.DB.Model(&scene).
		Column("published_at").
		Order("published_at ASC").
		Select()
	if err != nil {
		return 0, err
	}

	return scene.PublishedAt.Unix(), nil
}

func (indexer *ContentIndexer) indexHistory(jobs chan Job) error {
	queryOffset := 0

	for {
		log.Printf("(Indexer) Fetching history {offset: %d}", queryOffset)
		history, err := indexer.Client.GetHistoryWithOpts(api.GetHistoryParams{Offset: queryOffset})
		if err != nil {
			return err
		}

		log.Printf("(Indexer) Processing %d entries", len((*history).Events))
		for _, entry := range (*history).Events {
			job := Job{entry.EntityType, entry.EntityID}
			jobs <- job
		}

		if !history.Pagination.MoreData {
			break
		}

		queryOffset += history.Pagination.Offset + history.Pagination.Limit
	}
	return nil
}

func (indexer *ContentIndexer) runEpoch() error {
	// start workers to handle indexing
	log.Printf("(Indexer) Starting %d workers", indexer.indexWorkers)
	wg := sync.WaitGroup{}
	wg.Add(indexer.indexWorkers)
	jobs := make(chan Job, indexer.indexWorkers)
	for i := 0; i < indexer.indexWorkers; i++ {
		go indexer.runWorker(i, jobs, &wg)
	}

	// replay history
	err := indexer.indexHistory(jobs)
	if err != nil {
		return err
	}

	log.Info("(Indexer) Waiting for workers...")
	close(jobs)
	wg.Wait()

	// index tiles based on all the data
	log.Info("(Indexer) Processing tiles...")
	err = indexer.indexTiles(0)
	if err != nil {
		return err
	}

	return nil
}

// Run process new content entries and index them
func (indexer *ContentIndexer) Run(ctx context.Context) {
	lastRunTimestamp := time.Time{}
	indexIntervalSeconds := time.Duration(indexer.indexInterval) * time.Second

	// main loop
	for {
		select {
		case <-ctx.Done():
			log.Info("Bye")
			return
		default:
			awakeIndexer := time.Now().Sub(lastRunTimestamp) > indexIntervalSeconds
			if awakeIndexer {
				err := indexer.runEpoch()
				if err != nil {
					log.Panic(err)
				}
				lastRunTimestamp = time.Now()
				continue
			}
			time.Sleep(time.Second)
		}
	}
}
