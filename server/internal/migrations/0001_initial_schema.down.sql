-- Reverse the initial schema migration.
-- Drop tables in reverse dependency order.

DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS racks;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS super_blocks;
DROP TABLE IF EXISTS sites;
DROP TABLE IF EXISTS fabrics;
DROP TABLE IF EXISTS designs;
DROP TABLE IF EXISTS device_models;
