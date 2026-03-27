-- Reverse rack modeling migration.

-- Rebuild racks table without rack_type_id and power_capacity_w, block_id NOT NULL.
CREATE TABLE IF NOT EXISTS racks_old (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id    INTEGER NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL CHECK (type IN ('physical', 'logical')),
    height_u    INTEGER NOT NULL DEFAULT 42,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO racks_old (id, block_id, name, type, height_u, description, created_at, updated_at)
SELECT id, COALESCE(block_id, 0), name, 'physical', height_u, description, created_at, updated_at
FROM racks;

DROP TABLE racks;
ALTER TABLE racks_old RENAME TO racks;

CREATE INDEX IF NOT EXISTS idx_racks_block_id ON racks (block_id);

DROP TABLE IF EXISTS rack_types;
