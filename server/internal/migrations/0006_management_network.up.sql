-- Management network migration: expand device roles to include management plane roles.
--
-- block_aggregations was introduced in migration 0005 (block aggregation).
-- This migration only extends the devices table CHECK constraint.

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
