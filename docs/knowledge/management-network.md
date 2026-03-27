# Management Network

## Overview

The management network is an out-of-band (OOB) network that provides access to the
management interfaces of all network devices — independent of the front-end production
fabric. In fabrik, the management plane is modeled as a first-class design object
alongside the Clos fabric.

Every management device consumes rack space (RU) and power just like front-end fabric
devices. The management plane uses the same block aggregation model as the front-end
fabric: each block gets a management aggregation switch alongside the front-end
aggregation switch.

---

## Device Roles

Two new device roles support management network modeling:

| Role | Description |
|------|-------------|
| `management_tor` | Management top-of-rack switch — one per rack, connected to the block management agg |
| `management_agg` | Management block aggregation switch — one per block, aggregates all management ToRs |

These roles are accepted anywhere a device role is required (e.g., when placing a device
in a rack via the API).

### Existing Roles

The front-end fabric roles remain unchanged:

| Role | Description |
|------|-------------|
| `leaf` | Clos leaf switch |
| `spine` | Clos spine switch |
| `super_spine` | Clos super-spine (3- and 5-stage fabrics) |
| `server` | End host / compute node |
| `other` | Any device not covered by the above |

---

## Rack Capacity

### Management ToR Placement

When placing a `management_tor` device in a rack:

- **RU capacity**: Checked but treated as a **soft warning** — placement proceeds even if
  the rack is full. This reflects real-world scenarios where a management switch may need
  to be shoehorned into a full rack.
- **Power capacity**: Checked as a **soft warning** — placement proceeds with a warning
  if the rack's power budget would be exceeded.

This differs from front-end fabric devices where RU overflow is a hard block.

### Capacity Warning Examples

- **RU overflow**: `management switch placement: no contiguous 1U slot available in rack`
- **Power overflow**: `power capacity exceeded: 500W used + 200W new = 700W > 600W capacity`

---

## Block Management Aggregation

Each block can be assigned a management aggregation switch via the API. This aggregation
record tracks:

- Which device serves as the management agg for this block
- The total uplink port capacity (`max_ports`)
- The number of ports already allocated (`used_ports`)

### Port Capacity Enforcement

Management agg port capacity is enforced as a **hard block**. When all ports are
allocated, additional management ToR placements will fail until ports are released.

This is different from front-end fabric links and reflects how management networks are
typically designed: the agg switch is a fixed-capacity device with a known number of
downlinks.

### No Management Agg Assigned

When a `management_tor` is placed in a rack that belongs to a block with no management
agg assigned, the placement succeeds but returns a warning:

> `no management aggregation assigned to this block; management ToR has no upstream connectivity`

This lets you place management switches incrementally before finalizing the aggregation
design.

---

## Topology Visualization

The topology view includes a **Management Plane** toggle. When enabled:

- Management nodes are rendered in a **distinct color** (tertiary palette) separate from
  front-end fabric nodes (primary palette).
- Management links are rendered as a **dashed line** in the tertiary color.
- The toggle state persists within the session but is not saved across page reloads.

### Color Coding

| Element | Color |
|---------|-------|
| Management ToR | Tertiary (teal/green) |
| Management Agg | Darker tertiary |
| Management links | Dashed tertiary |
| Front-end fabric nodes | Primary (blue) |
| Front-end fabric links | Solid primary |

---

## API Reference

### Management Aggregation

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/api/blocks/{block_id}/management-agg` | Assign or update management agg for a block |
| `GET` | `/api/blocks/{block_id}/management-agg` | Get management agg for a block |
| `DELETE` | `/api/blocks/{block_id}/management-agg` | Remove management agg from a block |
| `GET` | `/api/blocks/{block_id}/aggregations` | List all aggregation records for a block |

#### Request body for `PUT /api/blocks/{block_id}/management-agg`

```json
{
  "device_id": 42,
  "max_ports": 48,
  "description": "Block A management agg"
}
```

- `device_id`: Optional. Reference to the device that serves as the management agg.
  Set to `null` to clear the device assignment while keeping the aggregation record.
- `max_ports`: Maximum number of ToR uplinks. Set to `0` to disable capacity enforcement.
- `description`: Optional free-text description.

### Device Placement with Management Roles

Use the standard device placement API (`POST /api/racks/{id}/devices`) with the
`role` field set to `management_tor` or `management_agg`:

```json
{
  "device_model_id": 5,
  "name": "mgmt-tor-rack-01",
  "role": "management_tor",
  "position": 0
}
```

Position `0` auto-places the device at the lowest available slot.

---

## Design Patterns

### Standard Management Topology

A typical datacenter block management topology:

```
[Server] [Server] [Server]
    \       |       /
  [Management ToR — rack 1]
  [Management ToR — rack 2]   <-- one per rack
  [Management ToR — rack N]
          |
  [Management Agg Switch]     <-- one per block
          |
  [Out-of-band Network / Jump Host]
```

### Port Sizing

- **Management ToR**: 1U, 24–48 1G copper ports (e.g., Cisco C9200L-24T)
- **Management Agg**: 1U, 24–48 1G ports (e.g., Cisco C9200L-48T or equivalent)

A 48-port management agg can support up to 48 racks in a block. Larger blocks may
require multiple management agg switches or a higher-density agg platform.

### Power Budget

Management switches are typically low-power (30–60W each). However, they must be
accounted for in rack power budgets, especially in high-density deployments.

Rough estimates per rack:
- Management ToR: ~40–60W
- Additional PDU capacity: typically negligible (<1% of rack power budget)

---

## Related Concepts

- [Device Catalog](device-catalog.md) — Hardware platforms for management switches
- Block Aggregation — How blocks are connected to the wider network
- Rack Modeling — RU and power capacity management
