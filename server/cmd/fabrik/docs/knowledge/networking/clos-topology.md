# Clos Topology Designer

The Clos topology designer lets you plan multi-stage folded-Clos fabrics for datacenter networks.
You configure the stage count, switch radix, and oversubscription ratio; the tool calculates
the exact switch counts and port distributions required.

## Topology parameters

### Stage count

| Stages | Common name | Use case |
|--------|-------------|----------|
| 2 | Leaf-Spine | Small to mid-scale pods |
| 3 | Leaf-Spine-SuperSpine | Multi-pod campus or medium DC |
| 5 | Extended Clos | Hyperscale / multi-PoD fabrics |

### Radix

The radix is the total number of ports on each switch in the fabric.
All switches in a given fabric design share the same radix.
The tool auto-corrects the radix to the nearest value that divides evenly
for the chosen oversubscription ratio and includes an explanation.

### Oversubscription

Oversubscription is the ratio of downlink capacity to uplink capacity on a leaf switch.
A ratio of **1.0** means the fabric is non-blocking: every host port has a dedicated
uplink path.

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

Once you have defined the topology parameters, assign real hardware switch models to each
role (leaf, spine, super-spine). The tool will surface a warning if the device's port count
differs from the configured radix, which may indicate a misconfiguration.

If no device models are in the catalog, navigate to the **Catalog** page to add them.

## Derived metrics

The designer calculates and displays:

| Metric | Description |
|--------|-------------|
| Total switches | Sum of all switches across all roles |
| Total host ports | Leaf count × leaf downlinks |
| Oversubscription ratio | As configured |
| Radix correction | Noted when radix is snapped to ensure even divisibility |

## References

- [Clos Fabric Fundamentals](clos-fabric-fundamentals.md)
- [Oversubscription](oversubscription.md)
- [ECMP](ecmp.md)
