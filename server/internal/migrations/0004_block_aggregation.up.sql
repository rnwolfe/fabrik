-- Block aggregation migration: block-level agg switch assignment and port connections.

-- Block aggregation assignments: one agg switch model per network plane per block.
CREATE TABLE IF NOT EXISTS block_aggregations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id         INTEGER NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    plane            TEXT    NOT NULL CHECK (plane IN ('frontend', 'management')),
    device_model_id  INTEGER NOT NULL REFERENCES device_models(id),
    created_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (block_id, plane)
);

CREATE INDEX IF NOT EXISTS idx_block_aggregations_block_id ON block_aggregations (block_id);

-- Port connections: bidirectional wiring between leaf uplinks and agg downlinks.
-- Each row represents a single port-pair allocated when a rack is added to a block.
CREATE TABLE IF NOT EXISTS port_connections (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    block_aggregation_id   INTEGER NOT NULL REFERENCES block_aggregations(id) ON DELETE CASCADE,
    rack_id                INTEGER NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    agg_port_index         INTEGER NOT NULL, -- 0-based port index on the agg switch
    leaf_device_name       TEXT    NOT NULL, -- logical name of the leaf/ToR in the rack
    created_at             DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (block_aggregation_id, agg_port_index)
);

CREATE INDEX IF NOT EXISTS idx_port_connections_block_aggregation_id ON port_connections (block_aggregation_id);
CREATE INDEX IF NOT EXISTS idx_port_connections_rack_id              ON port_connections (rack_id);
