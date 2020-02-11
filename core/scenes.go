package core

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-pg/pg/v9"
	"melonwave.com/glass/api"
	"melonwave.com/glass/models"
)

// SceneIndex is a runnable scene index
type SceneIndex struct{}

func (index *SceneIndex) saveSceneContents(db *pg.DB, sceneEntity *api.SceneEntity) error {
	var sceneContentList = make([]*models.SceneContent, len(sceneEntity.Content))
	for i, content := range sceneEntity.Content {
		sceneContent := &models.SceneContent{
			ID:      content.Hash,
			File:    content.File,
			SceneID: sceneEntity.ID,
		}
		sceneContentList[i] = sceneContent
	}
	return db.Insert(&sceneContentList)
}

// GetName returns the index name
func (index *SceneIndex) GetName() string {
	return "SceneIndex"
}

// Run scene index
func (index *SceneIndex) Run(logger *log.Entry, indexer *ContentIndexer, entityType string, entityID string) error {
	// only scene types
	if entityType != api.EntityTypeScene {
		return nil
	}

	logger.Info("Indexing")

	// check if already indexed
	err := indexer.DB.Select(&models.Scene{ID: entityID})
	if err == nil {
		logger.Warn("Skip as already present")
		return nil
	}

	// get scene from api
	sceneEntity, err := indexer.Client.GetSceneEntityByID(entityID)
	if err != nil {
		return err
	}

	// save scene
	rawScene, err := json.Marshal(sceneEntity)
	if err != nil {
		return err
	}
	newScene := &models.Scene{
		ID:          entityID,
		Title:       sceneEntity.Metadata.Display.Title,
		Pointers:    sceneEntity.Pointers,
		Raw:         rawScene,
		PublishedAt: time.Unix(sceneEntity.Timestamp/1000, 0),
	}
	err = indexer.DB.Insert(newScene)
	if err != nil {
		return err
	}

	// save scene files
	err = index.saveSceneContents(indexer.DB, sceneEntity)
	if err != nil {
		return err
	}

	return nil
}
