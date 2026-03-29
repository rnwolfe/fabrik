# Clos Topology Designer

The Clos topology designer lets you plan multi-stage folded-Clos fabrics for datacenter networks.
Stages emerge automatically from your design hierarchy — each level you equip with aggregation
switches adds one tier to the Clos fabric. The tool calculates switch counts and port distributions
based on the hardware you assign.

## How stages emerge

fabrik derives topology stages from the hierarchy you build:

| Hierarchy levels with aggregation | Resulting topology |
|------------------------------------|-------------------|
| Block only (leaf + spine) | 2-stage Leaf-Spine |
| Block + Super-Block (super-spine) | 3-stage Leaf-Spine-SuperSpine |
| Block + Super-Block + Site (agg) | 5-stage Extended Clos |

You do not configure stage count directly. Assign aggregation switch models at each hierarchy
level and the topology is derived automatically. This means designs naturally grow from 2-stage
to 3-stage to 5-stage as your datacenter expands.

## Topology parameters

### Radix

The radix is the total number of ports on each switch in the fabric.
All switches in a given fabric design share the same radix.
The tool auto-corrects the radix to the nearest value that divides evenly
for the chosen oversubscription ratio and includes an explanation.

### Oversubscription

Oversubscription is the ratio of downlink capacity to uplink capacity on a leaf switch.
A ratio of **1.0** means the fabric is non-blocking: every host port has a dedicated
uplink path. Oversubscription is derived from the spine count and port group split you
configure on each block.

| Ratio | Meaning |
|-------|---------|
| 1.0 | Non-blocking — full bisection bandwidth |
| 2.0 | 2:1 — two host ports share one uplink |
| 3.0 | 3:1 — three host ports share one uplink |

## Port allocation

### 2-Stage (Leaf-Spine)

Given radix `R` and oversubscription `os`:

```
uplinks   = R / (1 + os)
downlinks = R - uplinks
spines    = uplinks
```

Each leaf connects to every spine via one uplink port.
Each spine connects to all leaves via one port per leaf.

### 3-Stage (Leaf-Spine-SuperSpine)

The 3-stage fabric adds a super-spine tier that interconnects spine pods:

```
divisor      = 1 + os
uplinks      = R / divisor
downlinks    = R - uplinks
spines       = uplinks
super-spines = R / spines
```

### 5-Stage (Extended Clos)

The 5-stage fabric adds two additional aggregation tiers above the 3-stage spine layer.
It is used in hyperscale networks where a single super-spine tier cannot provide enough
bandwidth at the required scale.

## Device model assignment

Assign real hardware switch models at each hierarchy level (block, super-block, site).
The tool surfaces a warning if the device's port count differs from the derived radix,
which may indicate a misconfiguration.

If no device models are in the catalog, navigate to the **Catalog** page to add them.

## Derived metrics

The designer calculates and displays:

| Metric | Description |
|--------|-------------|
| Total switches | Sum of all switches across all roles |
| Total host ports | Leaf count × leaf downlinks |
| Oversubscription ratio | Derived from port group split and spine count |
| Stage count | Derived from hierarchy depth |

## References

- [Clos Fabric Fundamentals](clos-fabric-fundamentals.md)
- [Oversubscription](oversubscription.md)
- [ECMP](ecmp.md)
