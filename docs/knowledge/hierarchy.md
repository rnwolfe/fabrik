# Design Hierarchy Glossary

fabrik organises a datacenter network design as a five-level hierarchy.
Each level maps to a concrete entity in the data model.

## Levels

### Site

A physical location — a datacenter building, campus, or co-location facility.
A design may span multiple sites. Sites contain one or more super-blocks (pods).

### Super-Block (Pod)

A logical grouping of blocks within a site, typically corresponding to a
datacenter pod, hall, or row. Super-blocks share aggregation infrastructure
such as super-spine switches. A site contains one or more super-blocks.

### Block

The fundamental unit of a Clos fabric design. Each block is an independent
2-stage (leaf–spine) Clos fabric that serves a set of racks. Blocks are the
primary object you configure in the design canvas — you assign leaf and spine
models, tune spine count, and track rack capacity at the block level.

### Fabric

The internal switching structure of a block. A fabric has a tier (front-end
or back-end) and defines the leaf and spine switch roles, port counts, and
inter-tier cabling.

### Leaf

A top-of-rack (ToR) access switch. Leaf switches face servers on their
downlink ports and connect to spine switches on their uplink ports.

### Spine

An aggregation switch in the middle tier of a 2-stage Clos fabric. Spine
switches interconnect all leaves within a block and do not face servers
directly.

### Super-Spine

An aggregation switch that interconnects spine switches across multiple blocks
within a super-block (pod). Super-spines appear in 3-stage or 5-stage Clos
topologies when blocks must communicate at low latency within a pod.

## Visual Summary

```
Site
└── Super-Block (Pod)
    └── Block
        └── Fabric
            ├── Spine (× N)
            └── Leaf  (× M)
                └── Servers / hosts
```

## Related Topics

- [Topology Visualization](topology-visualization.md) — interactive graph view of your designs
- [Block Aggregation](block-aggregation.md) — block-level aggregation switches
- [Radix](radix.md) — how switch port count constrains fabric scale
- [Oversubscription](oversubscription.md) — leaf-to-spine bandwidth ratios
