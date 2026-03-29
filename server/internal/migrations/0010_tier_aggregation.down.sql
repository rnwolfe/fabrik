-- Reverse tier_aggregation migration: restore block_aggregations and port_connections.

CREATE TABLE IF NOT EXISTS block_aggregations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id         INTEGER NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    plane            TEXT    NOT NULL CHECK (plane IN ('front_end', 'management')),
    device_model_id  INTEGER NOT NULL REFERENCES device_models(id),
    created_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (block_id, plane)
);

CREATE INDEX IF NOT EXISTS idx_block_aggregations_block_id ON block_aggregations (block_id);

CREATE TABLE IF NOT EXISTS port_connections (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    block_aggregation_id   INTEGER NOT NULL REFERENCES block_aggregations(id) ON DELETE CASCADE,
    rack_id                INTEGER NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    agg_port_index         INTEGER NOT NULL,
    leaf_device_name       TEXT    NOT NULL,
    created_at             DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (block_aggregation_id, agg_port_index)
);

CREATE INDEX IF NOT EXISTS idx_port_connections_block_aggregation_id ON port_connections (block_aggregation_id);
CREATE INDEX IF NOT EXISTS idx_port_connections_rack_id              ON port_connections (rack_id);

-- Migrate data back (only block-scoped aggregations).
INSERT INTO block_aggregations (id, block_id, plane, device_model_id, created_at, updated_at)
SELECT id, scope_id, plane, device_model_id, created_at, updated_at
FROM tier_aggregations WHERE scope_type = 'block';

INSERT INTO port_connections (id, block_aggregation_id, rack_id, agg_port_index, leaf_device_name, created_at)
SELECT id, tier_aggregation_id, child_id, agg_port_index, child_device_name, created_at
FROM tier_port_connections;

DROP TABLE IF EXISTS tier_port_connections;
DROP TABLE IF EXISTS tier_aggregations;
