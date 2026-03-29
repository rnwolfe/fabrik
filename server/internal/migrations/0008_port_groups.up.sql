-- This migration originally added downlink/uplink columns to device_models.
-- Those columns are cleaned up by migration 0009 which replaces them with
-- a proper port_groups table. This file is kept for migration versioning.

ALTER TABLE device_models ADD COLUMN downlink_count     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN downlink_speed_gbps INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN uplink_count       INTEGER NOT NULL DEFAULT 0;
ALTER TABLE device_models ADD COLUMN uplink_speed_gbps  INTEGER NOT NULL DEFAULT 0;
