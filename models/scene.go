package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

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

// PointerToCoords converts a string pointer into numbers
func PointerToCoords(pointer string) (x int, y int, err error) {
	coords := strings.Split(pointer, ",")
	if len(coords) < 2 {
		return 0, 0, errors.New("Invalid pointer format")
	}
	x, err = strconv.Atoi(coords[0])
	if err != nil {
		return 0, 0, err
	}
	y, err = strconv.Atoi(coords[1])
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}
