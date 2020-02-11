package core

import (
	"time"
)

// Tile represents a parcel
type Tile struct {
	X           int `pg:",use_zero"`
	Y           int `pg:",use_zero"`
	SceneID     string
	PublishedAt time.Time
	UpdatedAt   time.Time `pg:",null"`
}
