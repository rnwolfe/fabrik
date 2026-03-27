---
title: Rack Modeling
category: infrastructure
tags: [rack, physical, constraint, capacity, power, RU, placement]
---

# Rack Modeling

fabrik's rack modeling feature lets you define rack templates, create rack instances, assign devices to specific RU positions, and automatically validate physical constraints — RU capacity and power budget — in real time.

## Core Concepts

### Rack Types (Templates)

A **rack type** is a reusable template that captures the standard specifications for a rack model:

- **Height (U)** — how many rack units the rack provides (e.g., 42U, 24U)
- **Power Capacity (W)** — the total PDU power budget in watts

When you create a rack from a rack type, the specs are copied into the rack instance. The rack can then diverge from the template (e.g., a custom power limit).

### Rack Instances

A **rack** is a physical installation. It may be:

- **Standalone** — not assigned to a block
- **Block-assigned** — placed in a logical Block within a Site hierarchy

Each rack tracks:
- Its inherited or custom height (U) and power capacity (W)
- All devices currently installed (with their RU positions)
- Real-time summary of used/available RU and power

### RU Positions

RU positions are **1-indexed from the bottom** of the rack. A 42U rack has positions 1–42. Multi-RU devices (e.g., a 7U chassis) occupy consecutive positions: a device at position 1 with height 7U occupies positions 1–7.

## Constraint Validation

### RU Capacity (Hard Limit)

fabrik rejects any placement that would overflow the rack's RU capacity:

- Device height_u > remaining available U → **400 Bad Request**
- Requested position + device height − 1 > rack height → **400 Bad Request**
- Position overlaps an existing device → **400 Bad Request**

Positions must be ≥ 1. Position 0 and negative positions are rejected.

### Power Budget (Soft Limit)

Power is a **soft limit**: oversubscription is allowed, but warned:

- Placing a device that would exceed the rack's power capacity → response includes a `warning` field
- Rack summary marks a warning when utilization exceeds 80%

This allows for planning headroom and accounting for redundant PSU configurations.

## API Overview

### Rack Types

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/rack-types` | Create a rack type |
| `GET` | `/api/rack-types` | List all rack types |
| `GET` | `/api/rack-types/:id` | Get a single rack type |
| `PUT` | `/api/rack-types/:id` | Update a rack type |
| `DELETE` | `/api/rack-types/:id` | Delete (409 if racks reference it) |

### Racks

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/racks` | Create a rack (optionally from a type) |
| `GET` | `/api/racks` | List racks (filter with `?block_id=`) |
| `GET` | `/api/racks/:id` | Get rack with usage summary |
| `PUT` | `/api/racks/:id` | Update name, description, block |
| `DELETE` | `/api/racks/:id` | Delete rack (cascades to devices) |

### Device Placement

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/racks/:id/devices` | Place device (auto-suggests position if omitted) |
| `PUT` | `/api/racks/:rack_id/devices/:device_id` | Move device within the same rack |
| `PUT` | `/api/racks/:rack_id/devices/:device_id/move` | Move device to a different rack |
| `DELETE` | `/api/racks/:rack_id/devices/:device_id` | Remove device (`?compact=true` to shift others down) |

## Rack Summary Response

`GET /api/racks/:id` returns an enriched summary:

```json
{
  "id": 1,
  "name": "row-a-rack-01",
  "height_u": 42,
  "power_capacity_w": 10000,
  "used_u": 8,
  "available_u": 34,
  "used_watts": 1800,
  "warning": "power utilization at 82% (8200W / 10000W)",
  "devices": [
    {
      "id": 5,
      "name": "spine-01",
      "position": 1,
      "height_u": 2,
      "power_watts": 900,
      "model_vendor": "Arista",
      "model_name": "7050TX-64",
      "role": "spine"
    }
  ]
}
```

## Device Removal: Compact Mode

When removing a device, pass `?compact=true` to shift all devices above the removed device's slot downward, filling the gap. Use `?compact=false` (default) to leave the gap empty and preserve existing positions.

## Cross-Rack Moves

To move a device from one rack to another, `PUT /api/racks/:src_rack_id/devices/:device_id/move` with the body:

```json
{ "dest_rack_id": 2, "position": 5 }
```

The same constraint checks apply to the destination rack. If `position` is omitted (or 0), fabrik auto-suggests the lowest available slot.
