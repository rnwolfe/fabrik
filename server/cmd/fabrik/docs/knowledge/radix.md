---
title: Switch Radix
category: networking
tags: [radix, switch, clos, port-count, design]
---

# Switch Radix

## What is radix

Radix is the total number of ports on a network switch. The term comes from mathematics (the base of a number system) and, in network engineering, describes how many connections a switch can simultaneously support. A 64-port Ethernet switch has a radix of 64; each port represents one potential connection to a server, another switch, or a storage device.

Radix is distinct from bandwidth: a high-radix switch might have many ports running at moderate speeds (e.g., 64 × 100 GbE), while a low-radix device might have fewer ports at very high speeds (e.g., 16 × 400 GbE). The right choice depends on whether your bottleneck is the number of things you need to connect or the bandwidth those connections require.

In fabrik, the radix of each device model is read directly from the hardware catalog. Port groups allow a model to express that some ports run at different speeds than others — which is common for leaf switches that have a mix of server-facing and uplink ports.

## Radix in Clos fabrics

Clos fabrics are fundamentally limited and shaped by the radix of their constituent switches. In a 2-stage (leaf-spine) Clos:

- **Spine radix** sets the maximum number of leaves that can connect to a single spine. If a spine has 64 ports and each leaf takes one uplink port per spine, you can have at most 64 leaves.
- **Leaf uplink count** (a fraction of leaf radix) determines how many spines can be connected and therefore how much oversubscription the fabric carries.
- **Leaf downlink count** (the remaining ports) sets how many servers each leaf can serve.

For a 3-stage (leaf-spine-super-spine) Clos, the math extends upward: spine switches split their radix between leaf-facing downlinks and super-spine-facing uplinks, introducing a second tier of oversubscription. The total scalability of the fabric is determined by the product of radix values across all stages.

A useful rule of thumb: with switches of equal radix *k*, a 2-stage Clos can connect up to *k²/4* servers non-blocking. With a 3-stage Clos the ceiling rises to *k³/4*. This is why network architects care deeply about switch radix — it multiplies.

## Choosing the right radix

Higher radix is generally better from a topology standpoint: more ports per switch means fewer switches to build the same size fabric, fewer cables, lower latency (fewer hops), and simpler operations. However, higher-radix ASICs are typically more expensive per port, consume more power, and may impose engineering constraints on the PCB and chassis.

Practical radix choices in 2024–2026 data centers:

| Switch class | Typical radix | Common use |
|---|---|---|
| ToR leaf | 48–64 | Server access, 25/100 GbE downlinks |
| Aggregation leaf | 32–64 | Spine-facing, 100/400 GbE |
| Spine | 32–128 | Leaf aggregation, 100/400 GbE |
| Super-spine | 64–128 | Pod aggregation, 400 GbE |

When choosing a leaf model in fabrik, the radix directly determines the maximum spine count (capped at the number of uplink ports defined in the port groups) and the per-leaf host port capacity. Higher radix leaves allow more spines — and therefore lower oversubscription — without swapping out hardware.
