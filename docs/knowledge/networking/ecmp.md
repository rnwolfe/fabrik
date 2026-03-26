---
title: ECMP Load Balancing
category: networking
tags: [ecmp, load-balancing, hashing, multipath, fabric]
---

# ECMP Load Balancing

Equal-Cost Multi-Path (ECMP) is the mechanism that enables a Clos fabric to distribute traffic across all available parallel paths. Without effective ECMP, a multi-path topology provides no practical benefit over a single path.

## How ECMP Works

When a router has multiple next-hops with equal cost to a destination, it uses a hash function to select which next-hop to use for each flow.

The hash is computed over packet header fields — typically the **5-tuple**:
1. Source IP address
2. Destination IP address
3. Source port
4. Destination port
5. IP protocol (TCP/UDP/ICMP)

All packets in the same flow hash to the same value, ensuring **in-order delivery** within a flow. Different flows hash to different values, distributing load across paths.

## Hash Polarization

**Hash polarization** is a critical problem in multi-tier Clos fabrics. If all switches use the same hash algorithm and seed, flows that hash to the same path at tier 1 will also hash to the same path at tier 2.

Result: Only a fraction of the available paths carry traffic — the rest are idle.

### Solution: Per-Hop Hash Entropy

Each switch should use a **different hash seed or algorithm** so that flow distribution at each tier is independent. Most modern ASICs (Broadcom, Cisco, Arista) support configuring per-switch hash seeds.

Example configuration approach:
- Leaf switches: seed = 0x1234
- Spine switches: seed = 0x5678
- Super-spine: seed = 0x9ABC

## Flow Affinity and Rebalancing

ECMP is typically **static** — the hash result does not change unless the route table changes. When a path fails:

1. The failed path is removed from the ECMP group.
2. All flows are rehashed across remaining paths.
3. **All flows are disrupted**, not just flows that were using the failed path.

This "full shuffle" behavior causes a brief disruption to all TCP connections during a failure event.

### Resilient ECMP

**Resilient ECMP** (supported on most modern platforms) minimizes disruption by:
- Maintaining a hash bucket table that maps hash values to next-hops.
- On failure, only the buckets assigned to the failed next-hop are reassigned.
- Flows using healthy paths are **not** disrupted.

Resilient ECMP is strongly recommended for production fabrics.

## Elephant Flows

**Elephant flows** are large, long-lived flows (e.g., backup jobs, VM migrations) that can overwhelm a single ECMP path while other paths sit idle.

Standard ECMP cannot detect or react to elephant flows. Solutions include:

| Approach | Description | Tradeoff |
|----------|-------------|----------|
| Flowlet switching | Burst-level load balancing exploiting gaps between bursts | Requires hardware support |
| Adaptive routing | Per-packet or per-burst rebalancing based on congestion signals | Disrupts flow affinity |
| ECMP with more paths | More paths dilutes the impact of any single elephant | Requires larger spine |
| Application-level striping | Application opens multiple parallel TCP sessions | Requires application changes |

## ECMP Group Sizes

Most hardware has limits on ECMP group sizes:

- Typical maximum: 64, 128, or 256 next-hops per group
- In practice: spine count per leaf is rarely above 32

For most designs, hardware limits are not a concern. In very large three-tier fabrics (100+ pod spines), verify your hardware's ECMP table capacity.

## BGP Unnumbered and ECMP

Modern leaf-spine fabrics commonly use **BGP unnumbered** (RFC 5549) to advertise /32 or /128 host routes. In this model:

- Each link uses link-local IPv6 addresses for BGP peering.
- Host routes are redistributed via BGP into the fabric.
- ECMP is achieved natively through BGP multipath.

BGP unnumbered simplifies IP address management and provides flexible ECMP across the full fabric.

## Related Topics

- [Clos Fabric Fundamentals](networking/clos-fabric-fundamentals)
- [Oversubscription Ratios](networking/oversubscription)
