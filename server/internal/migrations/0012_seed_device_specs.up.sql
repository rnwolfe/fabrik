-- Update seed device models with real-world specifications.
-- Server models receive CPU, memory, and storage specs.
-- All models receive accurate idle/max power figures alongside the existing typical.

-- Generic 48-port switch
UPDATE device_models
SET power_watts_idle = 100,
    power_watts_max  = 400,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Generic' AND model = '48-port switch';

-- Generic 1RU server — mid-range dual-socket config
UPDATE device_models
SET device_model_type  = 'server',
    power_watts_idle   = 150,
    power_watts_max    = 550,
    cpu_sockets        = 2,
    cores_per_socket   = 16,
    ram_gb             = 256,
    storage_tb         = 2,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Generic' AND model = '1RU server';

-- Cisco Nexus 9364C-GX2A — 64-port 400GbE spine (2RU)
UPDATE device_models
SET power_watts_idle = 800,
    power_watts_max  = 3000,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Cisco' AND model = 'Nexus 9364C-GX2A';

-- Cisco Nexus 93180YC-FX3 — 48x25GbE + 6x100GbE leaf (1RU)
UPDATE device_models
SET power_watts_idle = 150,
    power_watts_max  = 520,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Cisco' AND model = 'Nexus 93180YC-FX3';

-- Dell PowerEdge R750 — 1RU, 2x Intel Xeon Gold 6338 (32c/socket), 512 GB DDR4, 4 TB NVMe
UPDATE device_models
SET device_model_type  = 'server',
    power_watts_idle   = 250,
    power_watts_max    = 1000,
    cpu_sockets        = 2,
    cores_per_socket   = 32,
    ram_gb             = 512,
    storage_tb         = 4,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Dell' AND model = 'PowerEdge R750';

-- Dell PowerEdge R6625 — 1RU, 2x AMD EPYC 9334 (32c/socket), 384 GB DDR5, 3.84 TB NVMe
UPDATE device_models
SET device_model_type  = 'server',
    power_watts_idle   = 200,
    power_watts_max    = 700,
    cpu_sockets        = 2,
    cores_per_socket   = 32,
    ram_gb             = 384,
    storage_tb         = 3.84,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Dell' AND model = 'PowerEdge R6625';

-- Add additional seed models not present in earlier migrations.

-- Arista 7050CX3-32S — 32x100GbE QSFP28 spine/leaf (1RU)
INSERT OR IGNORE INTO device_models
    (vendor, model, device_model_type, port_count, height_u,
     power_watts_idle, power_watts_typical, power_watts_max,
     description, is_seed)
VALUES
    ('Arista', '7050CX3-32S', 'network', 32, 1,
     120, 280, 400,
     '32x 100GbE QSFP28 leaf/spine switch, 1RU', 1);

-- Arista 7060CX2-32S — 32x100GbE QSFP28 high-density spine (1RU)
INSERT OR IGNORE INTO device_models
    (vendor, model, device_model_type, port_count, height_u,
     power_watts_idle, power_watts_typical, power_watts_max,
     description, is_seed)
VALUES
    ('Arista', '7060CX2-32S', 'network', 32, 1,
     130, 310, 460,
     '32x 100GbE QSFP28 spine switch, 1RU', 1);

-- Dell PowerEdge R750xa — 1RU, GPU-capable, 2x Xeon Gold + 4x NVIDIA A30
INSERT OR IGNORE INTO device_models
    (vendor, model, device_model_type, port_count, height_u,
     power_watts_idle, power_watts_typical, power_watts_max,
     cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
     description, is_seed)
VALUES
    ('Dell', 'PowerEdge R750xa', 'server', 0, 2,
     400, 1500, 2000,
     2, 32, 512, 2, 4,
     '2RU GPU server: 2x Xeon Gold 6338 + 4x NVIDIA A30, 512 GB DDR4', 1);

-- Dell PowerEdge XE9680 — 8x NVIDIA H100 SXM5 GPU server (8RU)
INSERT OR IGNORE INTO device_models
    (vendor, model, device_model_type, port_count, height_u,
     power_watts_idle, power_watts_typical, power_watts_max,
     cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
     description, is_seed)
VALUES
    ('Dell', 'PowerEdge XE9680', 'server', 0, 8,
     2000, 10000, 12000,
     2, 52, 2048, 12, 8,
     '8RU AI/ML server: 2x Intel Xeon w9-3595X + 8x NVIDIA H100 SXM5 80 GB', 1);
