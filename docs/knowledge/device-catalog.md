# Device Catalog

The device catalog is a library of hardware models — switches, servers, and other datacenter
equipment — that fabrik uses as the foundation for rack design and Clos fabric planning.

## Overview

Every device in a rack or fabric is based on a **device model** from the catalog. A device model
captures the key electrical and physical properties of a hardware platform:

| Field | Description |
|-------|-------------|
| Vendor | Manufacturer name (e.g., Cisco, Dell, Generic) |
| Model | Platform name (e.g., Nexus 9364C-GX2A) |
| Port Count | Number of network ports |
| Height (U) | Rack unit height (1–50 U) |
| Power (W) | Typical power draw in watts |
| Description | Optional notes about the platform |

## Seed models

Seed models ship with fabrik and represent real hardware with accurate vendor specs. They are
marked as read-only (you cannot edit or delete them directly). Seed models cover:

- **Cisco Nexus 9364C-GX2A** — 64-port 400GbE spine switch, 2 RU, ~2000 W
- **Cisco Nexus 93180YC-FX3** — 48x 25GbE + 6x 100GbE leaf switch, 1 RU, ~400 W
- **Dell PowerEdge R750** — 2-socket Intel Xeon server, 1 RU, ~800 W
- **Dell PowerEdge R6625** — 2-socket AMD EPYC server, 1 RU, ~600 W
- **Generic 48-port switch** — Placeholder switch for quick-start designs, 1 RU, 300 W
- **Generic 1RU server** — Placeholder server for quick-start designs, 1 RU, 500 W

## Generic templates

Generic models (vendor = "Generic") let you start a design without committing to specific
hardware. Replace them with vendor-specific models as your design matures.

## Custom models

Click **Add Device** to define a custom hardware model with your own specs. Custom models
can be edited, archived, or duplicated at any time.

## Duplicating models

The **Duplicate** action creates an editable copy of any model — including read-only seed
models. Use this to customise a seed model (e.g., override port count or power draw for a
specific SKU variant) without losing the original.

## Archiving

Deleting a custom model archives it rather than permanently removing it. Archived models
are hidden from the catalog browser by default but remain referenced by any existing rack
or fabric design. They can be revealed with the **include archived** filter.

## Using the catalog in designs

The **Device Picker** panel appears in the rack and fabric views. Drag a device model from
the picker into a rack slot, or click to select it for assignment. The picker supports
live search by vendor or model name.
