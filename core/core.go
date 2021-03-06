package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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
	Client        *api.CatalystClient
	DB            *pg.DB
	Indexes       *[]Index
	indexWorkers  int
	indexInterval int
}

// NewContentIndexer creates a new indexer
func NewContentIndexer(
	client *api.CatalystClient,
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

func (indexer *ContentIndexer) indexTiles(from int64) error {
	logger := log.WithField("index", "tiles")

	scenes, err := models.GetScenesPointersFromDate(indexer.DB, from)
	if err != nil {
		return err
	}

	logger.Printf("(Tiles) Updating %d scenes tiles", len(*scenes))
	for _, scene := range *scenes {
		logger.WithFields(
			log.Fields{
				"scene":   scene.ID,
				"n_tiles": len(scene.Pointers),
			},
		).Info("Updating tile")

		for _, pointer := range scene.Pointers {
			// parse scene pointer
			x, y, err := models.PointerToCoords(pointer)
			if err != nil {
				logger.Warn(err)
				continue
			}

			// insert or update tile
			tile := models.Tile{x, y, scene.ID, scene.PublishedAt, time.Now()}
			err = models.UpsertTile(indexer.DB, &tile)
			if err != nil {
				logger.Warn(err)
				continue
			}
		}
	}

	return nil
}

func (indexer *ContentIndexer) processHistory(ctx context.Context, jobs chan Job) error {
	queryOffset := 0

	for {
		select {
		case <-ctx.Done():
			return errors.New("Execution interrupted")
		default:
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
				return nil
			}

			queryOffset = history.Pagination.Offset + history.Pagination.Limit
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

func (indexer *ContentIndexer) runEpoch(ctx context.Context) error {
	// start workers to handle indexing
	log.Printf("(Indexer) Starting %d workers", indexer.indexWorkers)
	wg := sync.WaitGroup{}
	wg.Add(indexer.indexWorkers)
	jobs := make(chan Job, indexer.indexWorkers)
	for i := 0; i < indexer.indexWorkers; i++ {
		go indexer.runWorker(i, jobs, &wg)
	}

	// replay history
	err := indexer.processHistory(ctx, jobs)
	log.Info("(Indexer) Waiting for workers...")
	close(jobs)
	wg.Wait()
	if err != nil {
		return err
	}

	// index tiles based on all the data
	log.Info("(Indexer) Processing tiles...")
	err = indexer.indexTiles(0)
	if err != nil {
		return err
	}

	return nil
}

func (indexer *ContentIndexer) downloadScene(dataDir string, sceneID string) error {
	sceneContents := []models.SceneContent{}
	err := indexer.DB.Model(&sceneContents).Where("scene_id = ?", sceneID).Select()
	if err != nil {
		return err
	}

	for _, sceneContent := range sceneContents {
		data, err := indexer.Client.GetContent(sceneContent.ID)
		if err != nil {
			return err
		}
		dir, _ := filepath.Split(sceneContent.File)
		os.MkdirAll(filepath.Join(dataDir, sceneID, dir), 0700)
		file, err := os.Create(filepath.Join(dataDir, sceneID, sceneContent.File))
		if err != nil {
			return err
		}
		file.Write(data)
		file.Close()
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
				err := indexer.runEpoch(ctx)
				if err != nil {
					log.Error(err)
				}
				lastRunTimestamp = time.Now()
				continue
			}
			time.Sleep(time.Second)
		}
	}
}
