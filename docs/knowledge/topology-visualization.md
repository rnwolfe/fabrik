# Topology Visualization

The topology visualization provides an interactive graph view of your Clos fabric designs,
rendered alongside the fabric designer in a resizable split pane.

## Overview

The visualization renders each fabric as a layered directed graph using
[Cytoscape.js](https://js.cytoscape.org/) with the `dagre` hierarchical layout.
Nodes represent switches (or groups of switches) and edges represent inter-layer
interconnects. Multiple fabrics can be visualized simultaneously — front-end and
back-end fabrics use distinct color schemes so they are easy to distinguish at a glance.

## Layout

The graph uses a bottom-up hierarchical layout:

```
Super-Spine  (top)
     │
   Spine
     │
    Leaf
     │
   Hosts  (bottom)
```

For 5-stage topologies the additional aggregation layers appear above Super-Spine.

## Collapsed vs. Expanded Nodes

By default each fabric tier is shown as a **collapsed summary node** that represents
the entire layer (e.g. "Leaf ×32"). This keeps the graph readable for large fabrics.

Click the per-fabric chip in the toolbar to toggle between collapsed and expanded views.
In expanded mode each individual switch becomes its own node. Use the **Expand All** /
**Collapse All** button in the visualization header to control all fabrics at once.

> **Performance note:** Expanding fabrics with more than 500 total nodes may cause
> noticeable rendering delays. A warning is shown when this threshold is exceeded.

## Color Coding

### Fabric tier

| Color | Meaning |
|-------|---------|
| Blue family | Front-end fabric |
| Green family | Back-end fabric |
| Teal accents | Super-spine / aggregation layers |

### Utilization

Node borders are color-coded by utilization level:

| Border style | Threshold | Meaning |
|---|---|---|
| Thin, normal | < 70% | Healthy |
| Orange, 3 px | 70 – 89% | Warning |
| Red, 4 px | ≥ 90% | Critical |

## Interactivity

### Hover
Hover over a node to see a visual highlight. Click the node to open the **Detail Panel**.

### Node Detail Panel
Clicking a node opens the detail panel on the right side of the visualization area.
It shows:

- Node name and role (Leaf / Spine / Super-Spine / …)
- Fabric name and tier
- Port utilization with a color bar
- Hardware details when a device model is assigned (vendor, model, port count, height, power draw)

### Edge Detail Panel
Clicking a link between nodes shows:

- Source and destination node IDs
- Link speed and port assignment (when available)

### Zoom and Pan
Use the mouse wheel to zoom and drag to pan the graph. The **Fit** button (fit-screen icon)
resets the view to show all nodes.

## Split Pane

The visualization lives in a horizontal split pane alongside the fabric designer.

- **Drag** the divider handle to resize the two panes.
- **Keyboard:** Focus the divider and use **Arrow Left / Arrow Right** to resize in 2%
  increments.
- Pane sizes are **persisted in session storage**, so they survive page navigation within
  the same browser session.

## Utilization Data

Port utilization values shown in the visualization are currently derived from
API-computed metrics. In the current version, representative utilization values are
shown for illustration; as metering data becomes available from the management plane,
these values will reflect live or sampled traffic statistics.

## Related Topics

- [Device Catalog](device-catalog.md) — assign hardware models to fabric roles
- [Management Network](management-network.md) — out-of-band network overlay
- [Power & Capacity](power-capacity.md) — per-rack power budgets and RU utilization
- [Block Aggregation](block-aggregation.md) — block-level aggregation switches
