# Block Aggregation and Rack-to-Aggregation Connectivity

## Overview

Block aggregation is the physical placement layer that connects the logical Clos fabric
design to real racks and devices. Every group of racks (a **Block**) can have one or two
aggregation (agg) switches — one per network plane — that act as the uplink point for all
leaf/ToR switches in that block.

This document explains:
- The block model and how it relates to super-blocks and racks
- How agg switches are assigned to blocks
- How leaf uplinks are wired to agg downlinks automatically
- Capacity planning and enforcement rules

---

## The Block Model

fabrik organises physical placement in a four-level hierarchy:

```
Design
└── Site
    └── Super-Block  (a data hall, pod, or room)
        └── Block    (a row or cluster of racks)
            └── Rack
                └── Device
```

A **Block** is the fundamental unit of aggregation. All racks within a block share the
same agg switch tier. A **Super-Block** contains one or more blocks; when you add a rack
to a super-block without specifying a block, fabrik auto-creates a **default block** and
places the rack there.

---

## Network Planes

fabrik models two independent network planes:

| Plane | Purpose |
|-------|---------|
| `frontend` | Production data traffic between servers and the Clos spine layer |
| `management` | Out-of-band management: IPMI, iDRAC, iLO, console servers |

Each plane gets its own agg switch in a block. A block may have:
- No agg switch (racks can be placed, but no connectivity is modelled)
- A frontend agg only
- A management agg only
- Both a frontend and a management agg

---

## Assigning an Aggregation Switch

Use **PUT /api/blocks/{id}/aggregations/{plane}** to assign an aggregation device model to
a block.

```json
PUT /api/blocks/5/aggregations/frontend
{
  "device_model_id": 12
}
```

The response includes full capacity utilisation:

```json
{
  "id": 1,
  "block_id": 5,
  "plane": "frontend",
  "device_model_id": 12,
  "total_ports": 32,
  "allocated_ports": 0,
  "available_ports": 32,
  "utilization": "0/32 ports allocated on frontend agg"
}
```

### Changing an Agg Model

You can replace the agg model at any time. If the new model has fewer ports than the
current allocation, the request is rejected with HTTP 422:

```
constraint violation: 20 ports allocated but new model only has 16 ports
```

To downsize, first remove enough racks to free the necessary ports, then re-assign.

---

## Auto-Connection: Rack Placement

When you add a rack to a block via **POST /api/blocks/add-rack**, fabrik automatically
allocates agg ports for every **leaf** (role = `leaf`) device in the rack.

Each leaf gets one port on each agg switch assigned to that block. The allocation is
bidirectional: it represents both the leaf's uplink and the agg's downlink.

### Example

A block has a 32-port frontend agg and a 32-port management agg. A rack is added with
two leaf switches. The result:

- 2 ports allocated on the frontend agg (indices 0 and 1)
- 2 ports allocated on the management agg (indices 0 and 1)
- Total: 4 port connections created

```json
POST /api/blocks/add-rack
{
  "rack_id": 10,
  "block_id": 5
}
```

Response:

```json
{
  "rack": { "id": 10, "block_id": 5, ... },
  "connections": [
    { "block_aggregation_id": 1, "rack_id": 10, "agg_port_index": 0, "leaf_device_name": "leaf-1" },
    { "block_aggregation_id": 1, "rack_id": 10, "agg_port_index": 1, "leaf_device_name": "leaf-2" },
    { "block_aggregation_id": 2, "rack_id": 10, "agg_port_index": 0, "leaf_device_name": "leaf-1" },
    { "block_aggregation_id": 2, "rack_id": 10, "agg_port_index": 1, "leaf_device_name": "leaf-2" }
  ]
}
```

### No Leaf Devices

If the rack has no devices with role `leaf`, placement succeeds without creating any port
connections.

### No Agg Assigned

If the block has no agg switch, placement succeeds with a warning:

```json
{
  "rack": { ... },
  "connections": [],
  "warning": "no aggregation switch assigned to this block; rack placed without connectivity"
}
```

---

## Removing a Rack

**DELETE /api/blocks/racks/{rack_id}** removes a rack from its block and deallocates all
port connections for that rack across every agg plane.

---

## Default Block

When you call **POST /api/blocks/add-rack** with a `super_block_id` but no `block_id`,
fabrik:

1. Looks for an existing block named `"default"` in that super-block.
2. If none exists, auto-creates it.
3. Places the rack in the default block.

This mirrors real-world practice where early designs don't yet have a detailed row layout.

---

## Capacity Planning

### Agg Port Count

The agg switch's port count (from the device catalog) sets the hard limit on how many
leaf uplinks the block can support.

**Rule of thumb**: a single 32-port agg can serve up to 32 ToR switches.  In a 2-leaf-per-rack
design, that is 16 racks per block per plane.

### Capacity Enforcement

Adding a rack to a block where the agg is full returns HTTP 422:

```
aggregation ports full: 32/32 ports allocated on frontend agg; need 2 more for 2 leaves
```

The error message identifies the plane that is full and the shortfall, so you can either
add more agg switches (a new block) or reduce the number of leaves per rack.

### Viewing Utilisation

**GET /api/blocks/{id}/aggregations** returns a summary for every plane, including
`allocated_ports`, `available_ports`, and a `utilization` string. A `warning` field is
populated when all ports are in use.

---

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/blocks` | Create a block |
| `GET` | `/api/blocks` | List blocks (filter by `super_block_id`) |
| `GET` | `/api/blocks/{id}` | Get a block |
| `PUT` | `/api/blocks/{id}/aggregations/{plane}` | Assign agg model to plane |
| `GET` | `/api/blocks/{id}/aggregations/{plane}` | Get agg summary for plane |
| `GET` | `/api/blocks/{id}/aggregations` | List all agg summaries for block |
| `DELETE` | `/api/blocks/{id}/aggregations/{plane}` | Remove agg and all port connections |
| `GET` | `/api/blocks/{id}/aggregations/{plane}/connections` | List port connections |
| `POST` | `/api/blocks/add-rack` | Add rack to block (auto-allocates ports) |
| `DELETE` | `/api/blocks/racks/{rack_id}` | Remove rack from block (deallocates ports) |
