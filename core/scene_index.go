package core

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"melonwave.com/glass/api"
)

// TODO: evaluate if indexer *ContentIndexer should be -> db and client

// Scene information
type Scene struct {
	tableName   struct{} `pg:"scenes"`
	ID          string
	Title       string          `pg:",notnull"`
	Pointers    []string        `pg:",array,notnull"`
	Raw         json.RawMessage `pg:",notnull"`
	PublishedAt time.Time       `pg:",notnull"`
	CreatedAt   time.Time       `pg:",null"`
}

// SceneContent file part of the content of a scene
type SceneContent struct {
	tableName struct{} `pg:"scenes_contents"`
	ID        string
	File      string    `pg:",notnull"`
	SceneID   string    `pg:",notnull"`
	CreatedAt time.Time `pg:",null"`
}

// Tile represents a parcel
type Tile struct {
	X         int
	Y         int
	sceneID   string
	UpdatedAt time.Time `pg:",null"`
}

func pointerToCoords(pointer string) (x int, y int) {
	coords := strings.Split(pointer, ",")
	x, _ = strconv.Atoi(coords[0])
	y, _ = strconv.Atoi(coords[1])
	return x, y
}

// SceneIndex is a runnable scene index
type SceneIndex struct{}

// Run scene index
func (index *SceneIndex) Run(indexer *ContentIndexer, entityType string, entityID string) error {
	// only scene types
	if entityType != api.EntityTypeScene {
		return nil
	}

	log.Printf("(SceneIndex) ID{%s} Index scene", entityID)

	// get scene
	// TODO: try to get from DB cache first
	sceneEntity, err := indexer.Client.GetSceneEntityByID(entityID)
	if err != nil {
		log.Printf("(SceneIndex) ID{%s} %s", entityID, err)
		return err
	}

	rawScene, err := json.Marshal(sceneEntity)
	if err != nil {
		log.Printf("(SceneIndex) ID{%s} %s", entityID, err)
		return err
	}

	// save scene
	scene := &Scene{
		ID:          entityID,
		Title:       sceneEntity.Metadata.Display.Title,
		Pointers:    sceneEntity.Pointers,
		Raw:         rawScene,
		PublishedAt: time.Unix(sceneEntity.Timestamp/1000, 0),
	}
	_, err = indexer.DB.Model(scene).
		Where("id = ?", scene.ID).
		OnConflict("DO NOTHING").
		SelectOrInsert()
	if err != nil {
		log.Printf("(SceneIndex) ID{%s} %s", entityID, err)
		return err
	}

	// save scene content
	var sceneContentList = make([]*SceneContent, len(sceneEntity.Content))
	for i, content := range sceneEntity.Content {
		sceneContent := &SceneContent{
			ID:      content.Hash,
			File:    content.File,
			SceneID: entityID,
		}
		sceneContentList[i] = sceneContent
	}
	err = indexer.DB.Insert(&sceneContentList)
	if err != nil {
		log.Printf("(SceneIndex) ID{%s} %s", entityID, err)
		return err
	}

	log.Printf("(SceneIndex) ID{%s} Save entity", entityID)
	return nil
}
