# Power and Resource Capacity

fabrik tracks power consumption and compute resource totals across every level of the
datacenter hierarchy: Rack → Block → SuperBlock → Site → Design.

## Power Fields on Device Models

Each device model carries three power figures:

| Field | Description |
|-------|-------------|
| `power_watts_idle` | Idle/no-load draw |
| `power_watts_typical` | Typical operating draw (used for oversubscription checks) |
| `power_watts_max` | Maximum draw under full load (used for hard capacity enforcement) |

Zero-watt devices (patch panels, cable trays) are always placeable regardless of rack
power budget.

## Rack Power Oversubscription

Every rack has a power capacity (`power_capacity_w`) and two oversubscription thresholds:

| Field | Default | Meaning |
|-------|---------|---------|
| `power_oversub_pct_warn` | 100% | Typical draw above this % triggers a soft warning |
| `power_oversub_pct_max` | 110% | Max draw above this % blocks device placement (hard constraint) |

The defaults allow a 10 % headroom buffer over the rated capacity before placements are
rejected. Set both to 100 for strict enforcement.

### Warning vs. Hard Constraint

- **Soft warning** — Typical draw exceeds `power_oversub_pct_warn` of the capacity. The
  device is still placed but the API response includes a `warning` field describing the
  utilisation level. The frontend displays this as an amber notification.

- **Hard constraint** — Max draw would exceed `power_oversub_pct_max` of the capacity.
  The placement is rejected with HTTP 422 and an error message. This cannot be overridden
  from the UI.

## Device Model Types

| Type | Value | Typical use |
|------|-------|-------------|
| Network | `network` | Switches, routers, load balancers |
| Server | `server` | Compute nodes, storage servers, GPUs |
| Storage | `storage` | Dedicated storage appliances |
| Other | `other` | PDUs, KVMs, console servers |

Server-type models support compute resource fields: `cpu_sockets`, `cores_per_socket`,
`ram_gb`, `storage_tb`, and `gpu_count`. These fields are zero for non-server models.

## Capacity Aggregation

The `GET /api/designs/{id}/capacity` endpoint returns aggregated totals for any level
in the hierarchy using the `level` and `entity_id` query parameters.

| `level` | `entity_id` required? | Description |
|---------|----------------------|-------------|
| `design` | No | Totals for the entire design |
| `site` | Yes | Totals for a specific site |
| `superblock` | Yes | Totals for a specific super-block |
| `block` | Yes | Totals for a specific block |
| `rack` | Yes | Totals for a single rack |

### Example Requests

```
GET /api/designs/1/capacity
GET /api/designs/1/capacity?level=design
GET /api/designs/1/capacity?level=rack&entity_id=7
GET /api/designs/1/capacity?level=block&entity_id=3
```

### Response Shape

```json
{
  "level": "design",
  "id": 1,
  "name": "prod-dc-2025",
  "power_watts_idle": 250000,
  "power_watts_typical": 480000,
  "power_watts_max": 650000,
  "total_vcpu": 12288,
  "total_ram_gb": 98304,
  "total_storage_tb": 2048.5,
  "total_gpu_count": 128,
  "device_count": 342
}
```

`total_vcpu` is derived as `cpu_sockets × cores_per_socket` and is not stored separately.

## Capacity Summary Component

The `<app-capacity-summary>` Angular component renders a power + compute card for any
hierarchy level.

```html
<!-- Design-level summary -->
<app-capacity-summary [designId]="1" [level]="'design'" />

<!-- Rack-level summary -->
<app-capacity-summary [designId]="1" [level]="'rack'" [entityId]="7" />
```

The component reloads automatically when any input changes.

## Seed Data

The fabrik seed catalog ships with representative power values for common network and
server hardware. All seed models have `device_model_type` set appropriately:

- Network switches have `port_count > 0` and zero compute resource fields.
- Server models have `cpu_sockets`, `cores_per_socket`, and `ram_gb` populated.

Custom models (non-seed) can be created and edited freely through the Device Catalog.
