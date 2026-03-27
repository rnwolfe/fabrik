-- Reverse of 0006_power_capacity.up.sql

-- Restore device_models to its pre-migration schema.
CREATE TABLE IF NOT EXISTS device_models_old (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor      TEXT    NOT NULL,
    model       TEXT    NOT NULL,
    port_count  INTEGER NOT NULL DEFAULT 0,
    height_u    INTEGER NOT NULL DEFAULT 1,
    power_watts INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL DEFAULT '',
    is_seed     INTEGER NOT NULL DEFAULT 0,
    archived_at DATETIME,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO device_models_old (
    id, vendor, model, port_count, height_u,
    power_watts, description, is_seed, archived_at, created_at, updated_at
)
SELECT
    id, vendor, model, port_count, height_u,
    power_watts_typical AS power_watts,
    description, is_seed, archived_at, created_at, updated_at
FROM device_models;

DROP TABLE device_models;
ALTER TABLE device_models_old RENAME TO device_models;

CREATE UNIQUE INDEX IF NOT EXISTS idx_device_models_vendor_model ON device_models (vendor, model);
CREATE INDEX IF NOT EXISTS idx_device_models_archived_at ON device_models (archived_at);

-- Note: SQLite does not support DROP COLUMN in older versions.
-- The rack/rack_type columns are left in place; a full table rebuild would be
-- required to remove them, which is unnecessary for a down migration.
