-- Revert seed device spec updates by zeroing out server resource fields
-- and resetting power fields to migration 0007 values.

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Generic' AND model = '48-port switch';

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    cpu_sockets = 0, cores_per_socket = 0, ram_gb = 0, storage_tb = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Generic' AND model = '1RU server';

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Cisco' AND model = 'Nexus 9364C-GX2A';

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Cisco' AND model = 'Nexus 93180YC-FX3';

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    cpu_sockets = 0, cores_per_socket = 0, ram_gb = 0, storage_tb = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Dell' AND model = 'PowerEdge R750';

UPDATE device_models
SET power_watts_idle = 0, power_watts_max = 0,
    cpu_sockets = 0, cores_per_socket = 0, ram_gb = 0, storage_tb = 0,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE is_seed = 1 AND vendor = 'Dell' AND model = 'PowerEdge R6625';

DELETE FROM device_models
WHERE is_seed = 1 AND vendor = 'Arista' AND model IN ('7050CX3-32S', '7060CX2-32S');

DELETE FROM device_models
WHERE is_seed = 1 AND vendor = 'Dell' AND model IN ('PowerEdge R750xa', 'PowerEdge XE9680');
