-- Reverse management network migration.

DROP TABLE IF EXISTS block_aggregations;

-- Recreate devices table without management roles.
CREATE TABLE IF NOT EXISTS devices_old (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    rack_id         INTEGER NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    device_model_id INTEGER NOT NULL REFERENCES device_models(id),
    name            TEXT    NOT NULL,
    role            TEXT    NOT NULL CHECK (role IN ('spine', 'leaf', 'super_spine', 'server', 'other')),
    position        INTEGER NOT NULL DEFAULT 1,
    description     TEXT    NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Only migrate rows that had pre-existing roles (exclude management roles).
INSERT INTO devices_old SELECT * FROM devices
    WHERE role IN ('spine', 'leaf', 'super_spine', 'server', 'other');

DROP TABLE devices;
ALTER TABLE devices_old RENAME TO devices;

CREATE INDEX IF NOT EXISTS idx_devices_rack_id ON devices (rack_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_model_id ON devices (device_model_id);
