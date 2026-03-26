-- Initial schema for fabrik
-- Creates all core tables with foreign keys and indexes.

-- Device model catalog (no FK dependencies)
CREATE TABLE IF NOT EXISTS device_models (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor      TEXT    NOT NULL,
    model       TEXT    NOT NULL,
    port_count  INTEGER NOT NULL DEFAULT 0,
    height_u    INTEGER NOT NULL DEFAULT 1,
    power_watts INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_device_models_vendor_model ON device_models (vendor, model);

-- Designs (top-level project)
CREATE TABLE IF NOT EXISTS designs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_designs_name ON designs (name);

-- Fabrics (Clos fabric tiers within a design)
CREATE TABLE IF NOT EXISTS fabrics (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    design_id   INTEGER NOT NULL REFERENCES designs(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    tier        TEXT    NOT NULL CHECK (tier IN ('frontend', 'backend')),
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_fabrics_design_id ON fabrics (design_id);

-- Sites (physical datacenter locations)
CREATE TABLE IF NOT EXISTS sites (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    design_id   INTEGER NOT NULL REFERENCES designs(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_sites_design_id ON sites (design_id);

-- Super-blocks (data halls or pods within a site)
CREATE TABLE IF NOT EXISTS super_blocks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id     INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_super_blocks_site_id ON super_blocks (site_id);

-- Blocks (rows or clusters within a super-block)
CREATE TABLE IF NOT EXISTS blocks (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    super_block_id INTEGER NOT NULL REFERENCES super_blocks(id) ON DELETE CASCADE,
    name           TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    created_at     DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at     DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_blocks_super_block_id ON blocks (super_block_id);

-- Racks (physical or logical racks within a block)
CREATE TABLE IF NOT EXISTS racks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id    INTEGER NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL CHECK (type IN ('physical', 'logical')),
    height_u    INTEGER NOT NULL DEFAULT 42,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_racks_block_id ON racks (block_id);

-- Devices (network devices installed in racks)
CREATE TABLE IF NOT EXISTS devices (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    rack_id         INTEGER NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    device_model_id INTEGER NOT NULL REFERENCES device_models(id),
    name            TEXT    NOT NULL,
    role            TEXT    NOT NULL CHECK (role IN ('spine', 'leaf', 'super_spine', 'server', 'other')),
    position        INTEGER NOT NULL DEFAULT 1,
    description     TEXT    NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_devices_rack_id ON devices (rack_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_model_id ON devices (device_model_id);

-- Ports (physical network ports on devices)
CREATE TABLE IF NOT EXISTS ports (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id   INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL CHECK (type IN ('ethernet', 'fiber', 'dac', 'other')),
    speed_gbps  INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_ports_device_id ON ports (device_id);
