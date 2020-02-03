CREATE TABLE IF NOT EXISTS scenes (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    pointers TEXT[] NOT NULL,
    raw JSONB NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scenes_contents (
    id TEXT,
    scene_id TEXT NOT NULL,
    file TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, scene_id, file)
);

CREATE INDEX ix_scenes_contents_scene_id ON scenes_contents USING BTREE(scene_id);