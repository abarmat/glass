CREATE TABLE tiles (
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    scene_id TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (x, y)
);

CREATE INDEX ix_tiles_scene_id ON tiles USING BTREE(scene_id);