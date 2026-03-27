-- Management network migration: expand device roles and add block aggregation support.

-- SQLite does not support ALTER TABLE ... MODIFY COLUMN with CHECK constraint changes.
-- We recreate the devices table to add management_tor and management_agg roles.
CREATE TABLE IF NOT EXISTS devices_new (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    rack_id         INTEGER NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    device_model_id INTEGER NOT NULL REFERENCES device_models(id),
    name            TEXT    NOT NULL,
    role            TEXT    NOT NULL CHECK (role IN (
                        'spine', 'leaf', 'super_spine', 'server', 'other',
                        'management_tor', 'management_agg'
                    )),
    position        INTEGER NOT NULL DEFAULT 1,
    description     TEXT    NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO devices_new SELECT * FROM devices;
DROP TABLE devices;
ALTER TABLE devices_new RENAME TO devices;

CREATE INDEX IF NOT EXISTS idx_devices_rack_id ON devices (rack_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_model_id ON devices (device_model_id);

-- Block aggregation assignments: one row per (block, plane) pair.
-- plane: 'front_end' or 'management'
-- device_id: optional reference to the aggregation switch device
-- max_ports: total uplink ports available on the agg switch
-- used_ports: ports already allocated to ToRs in this block
CREATE TABLE IF NOT EXISTS block_aggregations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id    INTEGER NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    plane       TEXT    NOT NULL CHECK (plane IN ('front_end', 'management')),
    device_id   INTEGER REFERENCES devices(id) ON DELETE SET NULL,
    max_ports   INTEGER NOT NULL DEFAULT 0,
    used_ports  INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (block_id, plane)
);

CREATE INDEX IF NOT EXISTS idx_block_aggregations_block_id ON block_aggregations (block_id);
