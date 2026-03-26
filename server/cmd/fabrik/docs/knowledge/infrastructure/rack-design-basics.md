---
title: Rack Design Basics
category: infrastructure
tags: [rack, physical, cabling, power, airflow, density]
---

# Rack Design Basics

Rack design bridges the logical network topology with physical datacenter constraints. A well-designed rack layout simplifies cabling, improves airflow, eases maintenance, and maximizes density.

## Standard Rack Form Factors

| Form Factor | Height | Notes |
|------------|--------|-------|
| Full rack | 42U | Standard datacenter rack |
| Half rack | 21U | Smaller footprint, less common |
| Open-frame | 42–47U | Better airflow, no door |
| High-density | 52U+ | Used for blade/GPU deployments |

"U" (rack unit) = 1.75 inches = 44.45 mm.

## Typical Rack Population

A standard 42U rack in a leaf-spine fabric commonly contains:

```
U42-U40  [ToR Switch 1]        ← leaf switch (redundant pair)
U39-U37  [ToR Switch 2]        ← leaf switch (redundant pair)
U36      [Patch panel]
U35-U02  [Servers × 16]        ← 2U servers
U01      [1U KVM/console]
```

High-density 1U servers can fit 36–40 in a 42U rack (leaving room for switches and patch panels).

## Airflow Design

Almost all datacenter equipment uses **front-to-back** airflow (cold air in front, hot air exhaust at back).

### Hot Aisle / Cold Aisle Layout

Racks are arranged in alternating rows facing each other:
- **Cold aisle**: rack fronts face each other — cold supply air enters here.
- **Hot aisle**: rack backs face each other — hot exhaust collected here.

This arrangement prevents hot and cold air mixing, improving cooling efficiency.

### Containment

**Hot aisle containment** seals the hot aisle with roof panels and end-of-row doors, capturing hot exhaust directly into the return air plenum. This is more effective than cold aisle containment for most designs.

## Power Distribution

Each rack requires dedicated power circuits. Key considerations:

- **Per-rack power budget**: Typically 10–20 kW for standard compute, 30–50+ kW for GPU/AI racks.
- **Dual power feeds**: All production servers should have dual PSUs on separate PDUs, fed from separate circuits (ideally separate UPS/ATS units).
- **Power density**: Higher power density requires more cooling infrastructure and closer attention to floor weight ratings.

### Power Hierarchy

```
Utility → Transformer → UPS → PDU → Server PSU
```

A-side and B-side feeds come from independent UPS units. PDUs in the same rack should not share a UPS.

## Cabling

### Structured vs. Point-to-Point

- **Structured cabling** uses patch panels and horizontal/vertical cable managers. More organized, easier to manage, but adds latency (typically negligible).
- **Direct attach** (DAC/AOC from switch to server) reduces cost and latency but can be messier.

### Cable Length Planning

For leaf-spine fabrics, all leaf-to-spine uplinks should be **equal length** to ensure symmetric ECMP performance. Plan cable runs from ToR switches to end-of-row or top-of-pod patch panels before ordering cables.

### Fiber vs. Copper

| Type | Max Distance | Speed | Cost |
|------|-------------|-------|------|
| DAC (copper) | 3–5m | Up to 400G | Lowest |
| AOC (active optical) | 100m | Up to 400G | Low |
| SR4 fiber | 100m | 40G/100G | Medium |
| LR4 fiber | 10km | 40G/100G | High |

For within-rack and top-of-rack connections, DAC cables are the most cost-effective. Cross-row and inter-pod connections typically use fiber.

## Weight and Floor Loading

- Standard 42U rack: 700–1500 lbs when fully populated.
- Raised floor load ratings: typically 1000–2000 lbs/sq ft for server rooms.
- High-density GPU racks can exceed 3000 lbs — verify floor ratings before deployment.

## Rack Identification and Labeling

Use a consistent naming scheme tied to the physical location:
- `{datacenter}-{room}-{row}{rack}` — e.g., `SJC01-A101-R03`
- Label the top-U of each rack with the rack ID.
- Patch panels should label each port with the connected device and port.

## Related Topics

- [Clos Fabric Fundamentals](networking/clos-fabric-fundamentals)
- [Power Planning](infrastructure/power-planning)
- [Oversubscription Ratios](networking/oversubscription)
