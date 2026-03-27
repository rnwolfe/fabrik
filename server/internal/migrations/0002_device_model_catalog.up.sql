-- Add catalog fields to device_models: is_seed flag and soft-delete timestamp.

ALTER TABLE device_models ADD COLUMN is_seed     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN archived_at DATETIME;

CREATE INDEX IF NOT EXISTS idx_device_models_archived_at ON device_models (archived_at);

-- Seed data: Generic templates
INSERT INTO device_models (vendor, model, port_count, height_u, power_watts, description, is_seed)
VALUES
    ('Generic', '48-port switch',  48,  1, 300,  'Generic 48-port switch for quick-start designs', 1),
    ('Generic', '1RU server',       0,  1, 500,  'Generic 1RU 2-socket server for quick-start designs', 1);

-- Seed data: Cisco Nexus 9300 series
INSERT INTO device_models (vendor, model, port_count, height_u, power_watts, description, is_seed)
VALUES
    ('Cisco', 'Nexus 9364C-GX2A',   64, 2, 2000, '64x 400GbE QSFP-DD spine switch', 1),
    ('Cisco', 'Nexus 93180YC-FX3',  54, 1,  400, '48x 25GbE SFP28 + 6x 100GbE QSFP28 leaf switch', 1);

-- Seed data: Dell PowerEdge servers
INSERT INTO device_models (vendor, model, port_count, height_u, power_watts, description, is_seed)
VALUES
    ('Dell', 'PowerEdge R750',   0, 1, 800, '2-socket 1RU server, Intel Xeon Scalable', 1),
    ('Dell', 'PowerEdge R6625',  0, 1, 600, '2-socket 1RU server, AMD EPYC', 1);
