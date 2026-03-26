# Data Model ERD

This document describes the core domain model for fabrik and the relationships
between entities.

## Entity Relationship Diagram

```
Design
  ├── Fabric  (tier: frontend | backend)
  └── Site
        └── SuperBlock
              └── Block
                    └── Rack  (type: physical | logical)
                          └── Device  (role: spine | leaf | super_spine | server | other)
                                └── Port  (type: ethernet | fiber | dac | other)

DeviceModel  (referenced by Device)
```

### ERD (Mermaid)

```mermaid
erDiagram
    Design {
        int    id PK
        string name
        string description
        time   created_at
        time   updated_at
    }

    Fabric {
        int    id PK
        int    design_id FK
        string name
        string tier
        string description
        time   created_at
        time   updated_at
    }

    Site {
        int    id PK
        int    design_id FK
        string name
        string description
        time   created_at
        time   updated_at
    }

    SuperBlock {
        int    id PK
        int    site_id FK
        string name
        string description
        time   created_at
        time   updated_at
    }

    Block {
        int    id PK
        int    super_block_id FK
        string name
        string description
        time   created_at
        time   updated_at
    }

    Rack {
        int    id PK
        int    block_id FK
        string name
        string type
        int    height_u
        string description
        time   created_at
        time   updated_at
    }

    DeviceModel {
        int    id PK
        string vendor
        string model
        int    port_count
        int    height_u
        int    power_watts
        string description
        time   created_at
        time   updated_at
    }

    Device {
        int    id PK
        int    rack_id FK
        int    device_model_id FK
        string name
        string role
        int    position
        string description
        time   created_at
        time   updated_at
    }

    Port {
        int    id PK
        int    device_id FK
        string name
        string type
        int    speed_gbps
        string description
        time   created_at
        time   updated_at
    }

    Design ||--o{ Fabric : "has"
    Design ||--o{ Site : "contains"
    Site ||--o{ SuperBlock : "contains"
    SuperBlock ||--o{ Block : "contains"
    Block ||--o{ Rack : "contains"
    Rack ||--o{ Device : "houses"
    Device ||--o{ Port : "has"
    DeviceModel ||--o{ Device : "models"
```

## Hierarchy

Physical placement follows a strict containment hierarchy:

```
Site → SuperBlock → Block → Rack → Device → Port
```

- **Site**: Physical datacenter building or campus.
- **SuperBlock**: A data hall, pod, or zone within a site.
- **Block**: A row or cluster of racks within a super-block.
- **Rack**: A physical cabinet (42U standard) or logical grouping.
- **Device**: A network switch or server installed at a rack position.
- **Port**: An individual switchport or NIC port on a device.

## Fabric Model

Fabric is separate from the physical hierarchy. A Fabric represents a Clos
network tier (front-end customer-facing or back-end storage/cluster) that spans
devices across the physical hierarchy.

## Device Catalog

`DeviceModel` is a shared catalog of hardware platforms. Multiple `Device`
instances can reference the same `DeviceModel`. The model captures:

- Vendor and model string (unique pair)
- Port count and form factor (height in rack units)
- Typical power draw (watts)

## Schema Notes

- All tables use `INTEGER PRIMARY KEY AUTOINCREMENT` with SQLite.
- `ON DELETE CASCADE` is used throughout the containment hierarchy so that
  deleting a parent automatically removes all children.
- `PRAGMA foreign_keys=ON` is set at connection open time.
- `PRAGMA journal_mode=WAL` is set for better concurrent read performance.
- All timestamps default to `strftime('%Y-%m-%dT%H:%M:%fZ', 'now')` in UTC.
