-- Generalize block_aggregations into tier_aggregations.
-- Aggregation switches can now be assigned at any hierarchy level (block, super_block, site).
-- spine_count is persisted per aggregation (previously ephemeral frontend state).

CREATE TABLE IF NOT EXISTS tier_aggregations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type       TEXT    NOT NULL CHECK (scope_type IN ('block', 'super_block', 'site')),
    scope_id         INTEGER NOT NULL,
    plane            TEXT    NOT NULL CHECK (plane IN ('front_end', 'management')),
    device_model_id  INTEGER NOT NULL REFERENCES device_models(id),
    spine_count      INTEGER NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (scope_type, scope_id, plane)
);

CREATE INDEX IF NOT EXISTS idx_tier_aggregations_scope ON tier_aggregations (scope_type, scope_id);

CREATE TABLE IF NOT EXISTS tier_port_connections (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    tier_aggregation_id   INTEGER NOT NULL REFERENCES tier_aggregations(id) ON DELETE CASCADE,
    child_id              INTEGER NOT NULL,
    agg_port_index        INTEGER NOT NULL,
    child_device_name     TEXT    NOT NULL,
    created_at            DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (tier_aggregation_id, agg_port_index)
);

CREATE INDEX IF NOT EXISTS idx_tier_port_connections_agg_id   ON tier_port_connections (tier_aggregation_id);
CREATE INDEX IF NOT EXISTS idx_tier_port_connections_child_id ON tier_port_connections (child_id);

-- Migrate existing data from old tables.
INSERT INTO tier_aggregations (id, scope_type, scope_id, plane, device_model_id, spine_count, created_at, updated_at)
SELECT id, 'block', block_id, plane, device_model_id, 0, created_at, updated_at
FROM block_aggregations;

INSERT INTO tier_port_connections (id, tier_aggregation_id, child_id, agg_port_index, child_device_name, created_at)
SELECT id, block_aggregation_id, rack_id, agg_port_index, leaf_device_name, created_at
FROM port_connections;

-- Drop old tables (data preserved in new tables).
DROP TABLE IF EXISTS port_connections;
DROP TABLE IF EXISTS block_aggregations;
