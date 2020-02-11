package models

import (
	"time"

	"github.com/go-pg/pg/v9"
)

// Tile represents a parcel
type Tile struct {
	X           int `pg:",use_zero"`
	Y           int `pg:",use_zero"`
	SceneID     string
	PublishedAt time.Time
	UpdatedAt   time.Time `pg:",null"`
}

// UpsertTile update scene and publication for tile
func UpsertTile(db *pg.DB, tile *Tile) error {
	_, err := db.Model(tile).
		OnConflict("(x, y) DO UPDATE").
		Set("scene_id = excluded.scene_id, published_at = excluded.published_at").
		Insert()
	return err
}
