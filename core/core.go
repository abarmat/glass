package core

import (
	"log"
	"sync"
	"time"

	"github.com/go-pg/pg/v9"
	"melonwave.com/glass/api"
)

// Job is request for indexing that can be run in the indexer
type Job struct {
	EntityType string
	EntityID   string
}

// Index is a runnable computation run by the indexer
type Index interface {
	Run(indexer *ContentIndexer, entityType string, entityID string) error
}

// ContentIndexer is the manager of the indexing process
type ContentIndexer struct {
	Client       *api.APIClient
	DB           *pg.DB
	Indexes      *[]Index
	indexWorkers int
}

// NewContentIndexer creates a new indexer
func NewContentIndexer(client *api.APIClient, db *pg.DB, indexes *[]Index, indexWorkers int) *ContentIndexer {
	return &ContentIndexer{client, db, indexes, indexWorkers}
}

func (indexer *ContentIndexer) indexAll(entityType string, entityID string) {
	for _, index := range *indexer.Indexes {
		index.Run(indexer, entityType, entityID)
	}
}

func (indexer *ContentIndexer) runWorker(id int, jobs chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case job, more := <-jobs:
			{
				if !more {
					return
				}
				indexer.indexAll(job.EntityType, job.EntityID)
			}
		}
	}
}

func (indexer *ContentIndexer) indexTiles(from int64) error {
	var scenes []Scene
	err := indexer.DB.Model(&scenes).
		Column("id", "pointers").
		Order("published_at ASC").
		Where("published_at >= ?", time.Unix(from, 0)).
		Select()
	if err != nil {
		return err
	}

	for _, scene := range scenes {
		for _, pointer := range scene.Pointers {
			x, y := pointerToCoords(pointer)
			tile := Tile{x, y, scene.ID, time.Now()}
			_, err := indexer.DB.Model(&tile).
				OnConflict("(x, y) DO UPDATE").
				Set("scene_id = ?scene_id").
				Insert()
			// TODO: what should I do on error?
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Run process new content entries and index them
func (indexer *ContentIndexer) Run() {
	// TODO: put the process of indexing within a Timer
	log.Printf("(Indexer) Starting %d workers", indexer.indexWorkers)
	wg := sync.WaitGroup{}
	wg.Add(indexer.indexWorkers)
	jobs := make(chan Job, indexer.indexWorkers)
	for i := 0; i < indexer.indexWorkers; i++ {
		go indexer.runWorker(i, jobs, &wg)
	}

	log.Printf("(Indexer) Fetching full history")
	history, err := indexer.Client.GetHistory()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("(Indexer) Processing %d entries", len(*history))
	for idx, entry := range *history {
		job := Job{entry.EntityType, entry.EntityID}
		jobs <- job
		log.Printf("(Indexer) New job %+v", job)
		// HACK: remove
		if idx > 40 {
			break
		}
		// HACK: remove
	}
	log.Println("(Indexer) Waiting for workers...")
	close(jobs)
	wg.Wait()

	log.Println("(Indexer) Processing tiles...")
	// TODO: complete with proper from ts
	err = indexer.indexTiles(0)
	if err != nil {
		log.Panicln(err)
	}

	// TODO: when it finishes keep pooling and process new history (from, to)
	// TODO: resume capabilities
}
