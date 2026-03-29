-- Remove seed models added in 0009
DELETE FROM device_models WHERE vendor = 'Cisco' AND model IN (
    'Nexus 9336C-FX2', 'Nexus 9332D-GX2B', 'Nexus 9348D-GX2A',
    'Nexus 93600CD-GX', 'Nexus 9316D-GX', 'Nexus 93108TC-FX3P',
    'Nexus 9348GC-FX3'
);

-- Restore original power draw for 9364C-GX2A
UPDATE device_models SET power_watts_typical = 2000
WHERE vendor = 'Cisco' AND model = 'Nexus 9364C-GX2A';

DROP TABLE IF EXISTS device_model_port_groups;

-- Re-add the columns that 0008 originally created
ALTER TABLE device_models ADD COLUMN downlink_count     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN downlink_speed_gbps INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN uplink_count       INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN uplink_speed_gbps  INTEGER NOT NULL DEFAULT 0;
