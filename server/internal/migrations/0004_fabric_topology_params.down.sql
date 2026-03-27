-- SQLite does not support DROP COLUMN in older versions.
-- Reversing this migration requires recreating the fabrics table without the topology columns.

CREATE TABLE fabrics_backup AS
    SELECT id, design_id, name, tier, description, created_at, updated_at
    FROM fabrics;

DROP TABLE fabrics;

CREATE TABLE fabrics (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    design_id   INTEGER NOT NULL REFERENCES designs(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    tier        TEXT    NOT NULL CHECK (tier IN ('frontend', 'backend')),
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO fabrics SELECT * FROM fabrics_backup;
DROP TABLE fabrics_backup;

CREATE INDEX IF NOT EXISTS idx_fabrics_design_id ON fabrics (design_id);
