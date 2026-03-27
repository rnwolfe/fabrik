-- Rack modeling migration: rack types, enhanced racks, and device placement.

-- Rack type templates (new table)
CREATE TABLE IF NOT EXISTS rack_types (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    name              TEXT    NOT NULL,
    height_u          INTEGER NOT NULL DEFAULT 42,
    power_capacity_w  INTEGER NOT NULL DEFAULT 0,
    description       TEXT    NOT NULL DEFAULT '',
    created_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rack_types_name ON rack_types (name);

-- Rebuild racks table: make block_id nullable, add rack_type_id and power_capacity_w.
-- SQLite does not support ALTER COLUMN, so we recreate the table.
CREATE TABLE IF NOT EXISTS racks_new (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id         INTEGER REFERENCES blocks(id) ON DELETE SET NULL,
    rack_type_id     INTEGER REFERENCES rack_types(id) ON DELETE SET NULL,
    name             TEXT    NOT NULL,
    height_u         INTEGER NOT NULL DEFAULT 42,
    power_capacity_w INTEGER NOT NULL DEFAULT 0,
    description      TEXT    NOT NULL DEFAULT '',
    created_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO racks_new (id, block_id, name, height_u, description, created_at, updated_at)
SELECT id, block_id, name, height_u, description, created_at, updated_at
FROM racks;

DROP TABLE racks;
ALTER TABLE racks_new RENAME TO racks;

CREATE INDEX IF NOT EXISTS idx_racks_block_id    ON racks (block_id);
CREATE INDEX IF NOT EXISTS idx_racks_rack_type_id ON racks (rack_type_id);
