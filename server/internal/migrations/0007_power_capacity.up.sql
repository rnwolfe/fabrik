-- Power and resource capacity migration.
-- Adds device model type, granular power fields, server resource fields,
-- and rack-level power oversubscription thresholds.

-- Add device_model_type column to device_models.
ALTER TABLE device_models ADD COLUMN device_model_type TEXT NOT NULL DEFAULT 'network'
    CHECK (device_model_type IN ('network', 'server', 'storage', 'other'));

-- Rename power_watts to power_watts_typical and add idle/max columns.
-- SQLite does not support ALTER COLUMN RENAME, so we recreate the table.
CREATE TABLE IF NOT EXISTS device_models_new (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor              TEXT    NOT NULL,
    model               TEXT    NOT NULL,
    device_model_type   TEXT    NOT NULL DEFAULT 'network'
                            CHECK (device_model_type IN ('network', 'server', 'storage', 'other')),
    port_count          INTEGER NOT NULL DEFAULT 0,
    height_u            INTEGER NOT NULL DEFAULT 1,
    power_watts_idle    INTEGER NOT NULL DEFAULT 0,
    power_watts_typical INTEGER NOT NULL DEFAULT 0,
    power_watts_max     INTEGER NOT NULL DEFAULT 0,
    -- Server resource fields (null/0 for non-server models)
    cpu_sockets         INTEGER NOT NULL DEFAULT 0,
    cores_per_socket    INTEGER NOT NULL DEFAULT 0,
    ram_gb              INTEGER NOT NULL DEFAULT 0,
    storage_tb          REAL    NOT NULL DEFAULT 0,
    gpu_count           INTEGER NOT NULL DEFAULT 0,
    description         TEXT    NOT NULL DEFAULT '',
    is_seed             INTEGER NOT NULL DEFAULT 0,
    archived_at         DATETIME,
    created_at          DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at          DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO device_models_new (
    id, vendor, model, device_model_type,
    port_count, height_u,
    power_watts_idle, power_watts_typical, power_watts_max,
    cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
    description, is_seed, archived_at, created_at, updated_at
)
SELECT
    id, vendor, model,
    CASE
        WHEN port_count > 0 THEN 'network'
        ELSE 'server'
    END AS device_model_type,
    port_count, height_u,
    0 AS power_watts_idle,
    power_watts AS power_watts_typical,
    0 AS power_watts_max,
    0, 0, 0, 0, 0,
    description, is_seed, archived_at, created_at, updated_at
FROM device_models;

DROP TABLE device_models;
ALTER TABLE device_models_new RENAME TO device_models;

CREATE UNIQUE INDEX IF NOT EXISTS idx_device_models_vendor_model ON device_models (vendor, model);
CREATE INDEX IF NOT EXISTS idx_device_models_archived_at ON device_models (archived_at);
CREATE INDEX IF NOT EXISTS idx_device_models_type ON device_models (device_model_type);

-- Add power oversubscription threshold columns to racks.
ALTER TABLE racks ADD COLUMN power_oversub_pct_warn INTEGER NOT NULL DEFAULT 100;
ALTER TABLE racks ADD COLUMN power_oversub_pct_max  INTEGER NOT NULL DEFAULT 110;

-- Add power oversubscription threshold columns to rack_types.
ALTER TABLE rack_types ADD COLUMN power_oversub_pct_warn INTEGER NOT NULL DEFAULT 100;
ALTER TABLE rack_types ADD COLUMN power_oversub_pct_max  INTEGER NOT NULL DEFAULT 110;
