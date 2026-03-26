---
title: Clos Fabric Fundamentals
category: networking
tags: [clos, fabric, spine-leaf, topology, switching]
---

# Clos Fabric Fundamentals

A Clos network is a multistage switching architecture that provides non-blocking or rearrangeably non-blocking connectivity between any input and any output port. Originally conceived by Charles Clos in 1953 for telephone switching, the topology has become the dominant architecture for modern datacenter networks.

## Core Concept

A Clos fabric replaces a single large switch (which would require an impractically large number of ports and line cards) with a network of smaller switches interconnected in a predictable, scalable pattern.

The key insight is that **every path through the fabric has equal cost and capacity**, enabling efficient load balancing via ECMP.

## Two-Tier (Leaf-Spine) Topology

The most common datacenter variant is the two-tier leaf-spine fabric:

```
Spine Layer
  [S1]  [S2]  [S3]  [S4]

Leaf Layer
  [L1]  [L2]  [L3]  [L4]  [L5]  [L6]
```

**Every leaf connects to every spine.** This is the defining property.

- **Leaf switches** (also called top-of-rack or ToR) connect directly to servers and other endpoints.
- **Spine switches** are pure transit nodes — they carry inter-leaf traffic only.
- There are no direct leaf-to-leaf links and no direct spine-to-spine links.

### Path Calculation

Between any two servers on different leaf switches, there are exactly **N spine paths** where N is the number of spine switches. Each path traverses exactly 3 hops: server → leaf → spine → leaf → server (the leaf-to-leaf path is 2 switch hops).

## Three-Tier (Super-Spine / Pod) Topology

When a two-tier fabric exceeds practical scale limits, a third tier is added:

```
Super-Spine Layer
  [SS1] [SS2] [SS3] [SS4]

Pod Spine Layer
  [PS1][PS2]    [PS3][PS4]    Pod1         Pod2

Leaf Layer
  [L1][L2][L3]    [L4][L5][L6]
```

Each **pod** is an independent leaf-spine fabric. Super-spine switches interconnect pods. This design is used in hyperscaler networks (Google Jupiter, Meta Fabric, etc.) to scale to hundreds of thousands of servers.

## Bandwidth and Oversubscription

In a leaf-spine fabric:

- **Uplink capacity** = number of spine switches × link speed
- **Downlink capacity** = number of servers × link speed
- **Oversubscription ratio** = downlink capacity / uplink capacity

A **1:1 (non-blocking)** fabric provides full bisectional bandwidth. Most datacenter fabrics use 3:1 or 4:1 oversubscription to reduce cost.

See [Oversubscription Ratios](networking/oversubscription) for detailed calculations.

## ECMP and Traffic Distribution

All paths in a Clos fabric have equal cost. Routers use **Equal-Cost Multi-Path (ECMP)** to distribute traffic across all available paths.

See [ECMP Load Balancing](networking/ecmp) for details on hash functions and flow affinity.

## Key Properties

| Property | Value |
|----------|-------|
| Diameter | 2 hops (same pod), 4 hops (cross-pod) |
| Path redundancy | N paths (N = spine count) |
| Failure domain | Single spine failure reduces capacity by 1/N |
| Scalability | Linear — add spines to increase capacity |

## Practical Sizing Example

For a leaf switch with 64 ports:
- Allocate 48 ports downlink (servers) at 25 GbE
- Allocate 16 ports uplink (spines) at 100 GbE
- Oversubscription = (48 × 25) / (16 × 100) = 1200 / 1600 = 0.75:1 (actually non-blocking)

Alternatively at 3:1:
- 48 ports downlink at 25 GbE = 1200 Gbps
- 16 ports uplink at 25 GbE = 400 Gbps
- Ratio = 1200 / 400 = 3:1

## References

- Charles Clos, "A Study of Non-Blocking Switching Networks," Bell System Technical Journal, 1953
- Arista Networks: [Clos Network Design Guide](https://www.arista.com/en/solutions/data-center-designs)
- Google: "Jupiter Rising: A Decade of Clos Topologies and Centralized Control in Google's Datacenter Network," SIGCOMM 2015
