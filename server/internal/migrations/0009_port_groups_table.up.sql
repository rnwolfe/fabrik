-- Replace the per-model downlink/uplink columns (added in 0008) with a proper
-- port_groups table. A device model can have multiple port groups (e.g., 48×25G + 6×100G).
-- Uplink/downlink role is determined by the fabric context, not the device itself.

-- Drop the columns from 0008
ALTER TABLE device_models DROP COLUMN downlink_count;
ALTER TABLE device_models DROP COLUMN downlink_speed_gbps;
ALTER TABLE device_models DROP COLUMN uplink_count;
ALTER TABLE device_models DROP COLUMN uplink_speed_gbps;

-- Create the port groups table
CREATE TABLE IF NOT EXISTS device_model_port_groups (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    device_model_id  INTEGER NOT NULL REFERENCES device_models(id) ON DELETE CASCADE,
    count            INTEGER NOT NULL CHECK (count > 0),
    speed_gbps       INTEGER NOT NULL CHECK (speed_gbps > 0),
    label            TEXT    NOT NULL DEFAULT '',
    created_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_port_groups_device_model_id
    ON device_model_port_groups (device_model_id);

-- Seed port groups for existing network device models.

-- Cisco Nexus 93180YC-FX3: 48×25GbE + 6×100GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 48, 25, '25GbE SFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93180YC-FX3';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 6, 100, '100GbE QSFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93180YC-FX3';

-- Cisco Nexus 9364C-GX2A: 64×400GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 64, 400, '400GbE QSFP-DD' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9364C-GX2A';

-- Generic 48-port switch: 48×10GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 48, 10, '10GbE' FROM device_models
WHERE vendor = 'Generic' AND model = '48-port switch';

-- Fix power draw for Nexus 9364C-GX2A (seeded at 2000W, actual typical is 1324W)
UPDATE device_models SET power_watts_typical = 1324
WHERE vendor = 'Cisco' AND model = 'Nexus 9364C-GX2A';

-- Add remaining Cisco Nexus 9300 series models
INSERT INTO device_models (vendor, model, port_count, height_u, power_watts_typical, description, is_seed)
VALUES
    ('Cisco', 'Nexus 9336C-FX2',     36, 1,  367, '36x 100GbE QSFP28 spine/leaf switch', 1),
    ('Cisco', 'Nexus 9332D-GX2B',    34, 1,  638, '32x 400GbE QSFP-DD + 2x 10GbE SFP+ spine switch', 1),
    ('Cisco', 'Nexus 9348D-GX2A',    50, 2, 1000, '48x 400GbE QSFP-DD + 2x 10GbE SFP+ spine switch', 1),
    ('Cisco', 'Nexus 93600CD-GX',    36, 1,  586, '28x 100GbE QSFP28 + 8x 400GbE QSFP-DD spine switch', 1),
    ('Cisco', 'Nexus 9316D-GX',      16, 1, 1010, '16x 400GbE QSFP-DD spine switch', 1),
    ('Cisco', 'Nexus 93108TC-FX3P',  54, 1,  360, '48x 10GBASE-T + 6x 100GbE QSFP28 leaf switch', 1),
    ('Cisco', 'Nexus 9348GC-FX3',    54, 1,  150, '48x 1GBASE-T + 4x 25GbE SFP28 + 2x 100GbE QSFP28 access switch', 1);

-- Port groups for new models

-- Nexus 9336C-FX2: 36×100GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 36, 100, '100GbE QSFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9336C-FX2';

-- Nexus 9332D-GX2B: 32×400GbE + 2×10GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 32, 400, '400GbE QSFP-DD' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9332D-GX2B';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 2, 10, '10GbE SFP+' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9332D-GX2B';

-- Nexus 9348D-GX2A: 48×400GbE + 2×10GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 48, 400, '400GbE QSFP-DD' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9348D-GX2A';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 2, 10, '10GbE SFP+' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9348D-GX2A';

-- Nexus 93600CD-GX: 28×100GbE + 8×400GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 28, 100, '100GbE QSFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93600CD-GX';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 8, 400, '400GbE QSFP-DD' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93600CD-GX';

-- Nexus 9316D-GX: 16×400GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 16, 400, '400GbE QSFP-DD' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9316D-GX';

-- Nexus 93108TC-FX3P: 48×10GbE + 6×100GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 48, 10, '10GBASE-T' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93108TC-FX3P';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 6, 100, '100GbE QSFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 93108TC-FX3P';

-- Nexus 9348GC-FX3: 48×1GbE + 4×25GbE + 2×100GbE
INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 48, 1, '1GBASE-T' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9348GC-FX3';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 4, 25, '25GbE SFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9348GC-FX3';

INSERT INTO device_model_port_groups (device_model_id, count, speed_gbps, label)
SELECT id, 2, 100, '100GbE QSFP28' FROM device_models
WHERE vendor = 'Cisco' AND model = 'Nexus 9348GC-FX3';
