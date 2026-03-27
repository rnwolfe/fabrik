# Metrics Dashboard

The metrics dashboard provides an at-a-glance view of your design's key operational
metrics: host count, switch count, bisection bandwidth, power consumption, and
per-tier oversubscription.

## Accessing the Dashboard

Navigate to **Metrics** in the top navigation. Select a design from the dropdown to
view its metrics. The dashboard auto-refreshes every 30 seconds.

## Summary Cards

| Card | What it shows |
|------|---------------|
| Total Host Ports | Sum of all host-facing ports across all fabrics |
| Total Switches | Sum of all switching nodes across all fabrics |
| Power Utilization | `total_draw / total_capacity × 100%` |
| Bisection Bandwidth | Bandwidth at the narrowest fabric tier (Gbps) |

## Oversubscription

Oversubscription is computed per tier for each fabric:

```
Leaf→Spine oversubscription = leaf_downlinks / leaf_uplinks
Spine→SuperSpine oversubscription = spine_downlinks / spine_uplinks
```

A ratio of 1.0 means non-blocking. Higher ratios indicate more hosts sharing uplink
capacity. Ratios are colour-coded:

| Range | Colour | Interpretation |
|-------|--------|----------------|
| ≤ 1.5 | Green | Healthy — low contention |
| 1.5–3.0 | Amber | Warning — shared uplinks |
| > 3.0 | Red | Critical — heavy oversubscription |

The **Chokepoint** callout highlights the single worst-case tier across all fabrics.

## Power Metrics

Shows total rack power capacity vs. device power draw from device model
`power_watts_typical` values. The gauge turns amber at 70% utilization and red at 90%.

## Resource Capacity

Aggregates compute resources across all racks and blocks in the design. Requires
devices to be placed in racks with device models that have CPU, RAM, storage, and GPU
fields populated.

## Port Utilization

Per-fabric, per-tier breakdown of total vs. allocated ports. Allocated ports are those
consumed by uplink/downlink connections within the Clos topology; remaining ports are
available for host connections or future expansion.

## Empty State

If a design has no fabrics or devices, the dashboard shows an empty state prompt.
Add racks, devices, and fabrics to populate the metrics.
