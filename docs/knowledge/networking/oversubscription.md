---
title: Oversubscription Ratios
category: networking
tags: [oversubscription, bandwidth, capacity, planning, fabric]
---

# Oversubscription Ratios

Oversubscription is the ratio of the maximum possible demand from endpoints to the available capacity in the network core. Understanding oversubscription is essential for designing cost-effective fabrics that meet application performance requirements.

## Definition

$$\text{Oversubscription Ratio} = \frac{\text{Total Downlink Bandwidth}}{\text{Total Uplink Bandwidth}}$$

A ratio of **1:1** means the fabric is non-blocking — every endpoint could transmit at full line rate simultaneously without congestion. A ratio of **3:1** means the core can only handle one-third of the aggregate endpoint bandwidth simultaneously.

## Why Oversubscribe?

Servers rarely transmit at full line rate simultaneously. Application traffic is inherently bursty. In practice:

- Web/application servers: 10:1 to 20:1 is often acceptable
- Storage traffic: 3:1 to 5:1 is typical
- HPC/AI/ML workloads: 1:1 is often required (all-to-all communication patterns)
- General enterprise: 4:1 to 8:1

Oversubscription trades **theoretical peak performance** for **cost reduction**.

## Leaf Switch Calculation

For a leaf switch with total port count P:

- Let $D$ = number of downlink ports (to servers)
- Let $U$ = number of uplink ports (to spines)
- Let $d$ = downlink port speed (Gbps)
- Let $u$ = uplink port speed (Gbps)

$$\text{Oversubscription} = \frac{D \times d}{U \times u}$$

### Example: 64-port 100GbE leaf

Scenario 1 — 3:1 oversubscription:
- 48 downlinks × 100 GbE = 4800 Gbps
- 16 uplinks × 100 GbE = 1600 Gbps
- Ratio = 4800 / 1600 = **3:1**

Scenario 2 — non-blocking:
- 32 downlinks × 100 GbE = 3200 Gbps
- 32 uplinks × 100 GbE = 3200 Gbps
- Ratio = 3200 / 3200 = **1:1** (but only 32 server ports per rack!)

## Spine Layer and Bisectional Bandwidth

**Bisectional bandwidth** is the minimum bandwidth between any two equal halves of the fabric. In a leaf-spine topology:

$$\text{Bisectional BW} = \text{Spine Count} \times \text{Uplink Speed per Leaf} \times \text{Leaf Count} / 2$$

If the leaf uplink bandwidth equals the sum of all leaf downlink bandwidth, the fabric achieves **full bisectional bandwidth** (1:1 oversubscription at the fabric level).

## Multi-Tier Oversubscription

In a three-tier fabric (super-spine / pod / leaf), oversubscription compounds:

$$\text{End-to-End Ratio} = \text{Leaf Oversubscription} \times \text{Pod Oversubscription}$$

Example:
- Leaf oversubscription: 3:1
- Pod (leaf-to-pod-spine) oversubscription: 2:1
- End-to-end: 6:1

Traffic within a pod sees only 3:1; traffic crossing pods sees 6:1.

## Traffic Matrix Assumptions

Your oversubscription choice should be validated against your actual traffic matrix:

| Workload Type | East-West % | Typical Ratio |
|---------------|-------------|---------------|
| Web/3-tier app | 30-50% | 4:1 – 8:1 |
| Big data / Hadoop | 70-80% | 2:1 – 4:1 |
| AI/ML training | 90-100% | 1:1 – 2:1 |
| Block storage | 60-70% | 2:1 – 3:1 |
| General mixed | 50-60% | 3:1 – 5:1 |

## Practical Recommendations

1. **Start with 3:1** for general workloads — it is the industry baseline.
2. **Use 1:1 or 2:1** for GPU clusters, distributed training, or any workload with high all-to-all communication.
3. **Monitor actual utilization** — if spine links consistently exceed 50-60%, you are likely oversubscribed for your workload.
4. **Plan for growth** — factor in a 20-30% growth buffer when sizing uplinks.

## Related Topics

- [Clos Fabric Fundamentals](networking/clos-fabric-fundamentals)
- [ECMP Load Balancing](networking/ecmp)
- [Rack Design Basics](infrastructure/rack-design-basics)
