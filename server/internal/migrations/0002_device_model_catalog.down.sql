-- Reverse migration 0002: remove catalog fields from device_models.
-- SQLite does not support DROP COLUMN before 3.35.0; we recreate the table.

-- Remove seed data inserted by the up migration.
DELETE FROM device_models WHERE is_seed = 1;

-- Recreate device_models without the new columns.
CREATE TABLE device_models_old AS
    SELECT id, vendor, model, port_count, height_u, power_watts, description, created_at, updated_at
    FROM device_models;

DROP TABLE device_models;

ALTER TABLE device_models_old RENAME TO device_models;

CREATE UNIQUE INDEX IF NOT EXISTS idx_device_models_vendor_model ON device_models (vendor, model);
