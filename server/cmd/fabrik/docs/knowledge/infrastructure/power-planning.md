---
title: Power Planning
category: infrastructure
tags: [power, capacity, pdu, ups, efficiency, pue]
---

# Power Planning

Power planning ensures that your datacenter has sufficient electrical capacity, redundancy, and efficiency to support the intended workloads — both now and as the deployment scales.

## Power Hierarchy

Understanding the power chain from utility to server is essential for capacity planning:

```
Utility Feed (MW)
  └── Transformer
        └── Main Distribution Board
              └── UPS (Uninterruptible Power Supply)
                    └── PDU (Power Distribution Unit)
                          └── Server PSU (A-side)
                          └── Server PSU (B-side, separate PDU/UPS)
```

Each stage has losses. The **Power Usage Effectiveness (PUE)** metric captures end-to-end efficiency.

## PUE (Power Usage Effectiveness)

$$PUE = \frac{\text{Total Facility Power}}{\text{IT Equipment Power}}$$

| PUE | Rating |
|-----|--------|
| 1.0 | Theoretical perfection (impossible) |
| 1.1 – 1.2 | Excellent (hyperscale) |
| 1.2 – 1.5 | Good (modern enterprise DC) |
| 1.5 – 2.0 | Average (legacy facilities) |
| 2.0+ | Poor (needs improvement) |

Google's hyperscale facilities average ~1.10 PUE. A typical enterprise datacenter is 1.5–1.7.

## Per-Rack Power Budget

Calculate the maximum power draw per rack before ordering equipment:

$$P_{rack} = \sum_{i} P_{device_i} \times \text{redundancy factor}$$

Rules of thumb:
- **Standard 1U server**: 200–500W typical, 700W max
- **2U server (high memory)**: 400–800W typical
- **GPU server (8× A100)**: 5–10 kW typical
- **1U ToR switch**: 200–600W
- **Patch panels / KVM**: negligible

Add 20% headroom for growth and peak loads.

## Redundancy Levels

### N+1 Redundancy

One extra UPS/PDU beyond what is needed to carry the load. If any one unit fails, capacity is maintained.

### 2N Redundancy

Full duplication — two independent paths from utility to server. Each path carries the entire load at 50% utilization during normal operation. Most critical datacenter tiers (Tier III, Tier IV) require 2N.

### 2N+1

The highest level — two complete independent paths plus one additional unit. Extremely expensive; used only for the most critical infrastructure.

## Tier Levels (Uptime Institute)

| Tier | Redundancy | Availability | Downtime/year |
|------|-----------|-------------|---------------|
| Tier I | None | 99.671% | ~28.8 hrs |
| Tier II | N+1 partial | 99.741% | ~22 hrs |
| Tier III | N+1 concurrent maintainability | 99.982% | ~1.6 hrs |
| Tier IV | 2N fault tolerant | 99.995% | ~0.4 hrs |

## Capacity Planning Calculations

### Critical Power (kW)

$$P_{critical} = P_{rack} \times N_{racks} \times \text{utilization factor}$$

Utilization factor accounts for diversity — not all racks will draw maximum power simultaneously. Use 0.7–0.85 for server racks.

### UPS Sizing

$$P_{UPS} = P_{critical} / \text{UPS efficiency} \times \text{redundancy factor}$$

A typical UPS operates at 92–96% efficiency at full load. Size UPS units to operate at 60–80% of rated capacity for efficiency and thermal reasons.

### Generator Sizing

Generators must start within 10–15 seconds of utility failure (UPS batteries provide this bridge). Generator capacity should cover:
- IT load
- Cooling systems (chillers, CRACs, fans)
- Lighting and life safety
- A 20% growth buffer

## Power Monitoring

Implement power monitoring at multiple levels:
- **Rack PDU level**: Per-outlet metering allows per-device power accounting.
- **UPS level**: Total load per UPS branch.
- **Building level**: Total facility consumption for PUE calculation.

Set alerting thresholds:
- Warning at 75% of circuit capacity
- Critical at 90% of circuit capacity

## Common Mistakes

1. **Underestimating idle power**: Servers at 10% CPU can still draw 60–70% of peak power.
2. **Not accounting for inrush current**: Servers draw 2–3× normal current at power-on; PDU breakers must handle this.
3. **Stranded capacity**: PDUs are often limited by the lowest-rated component in the chain (breaker, cable, connector).
4. **Single points of failure**: Verify A-side and B-side power paths are truly independent at every level.

## Related Topics

- [Rack Design Basics](infrastructure/rack-design-basics)
- [Clos Fabric Fundamentals](networking/clos-fabric-fundamentals)
