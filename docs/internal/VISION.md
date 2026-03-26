# fabrik — Vision

> A design-time tool for planning datacenter network topologies at any scale — from a few racks to hyperscale — grounded in real hardware platforms with beautiful, metrics-rich visualization.

## Identity

**What it is**: fabrik helps datacenter network architects plan Clos fabric topologies in fine detail. Users define racks, select (or create) hardware platforms, wire up front-end and back-end fabrics, and instantly see meaningful metrics like oversubscription ratios, port utilization, power consumption, and aggregate resource capacity — all through an intuitive, richly interactive visual interface. If NetBox captures "what is", fabrik captures "how it should be".

**Who it's for**: Datacenter network architects designing high-scale Clos fabrics — including GPU-heavy AI/ML clusters with back-end RDMA networks — who need to root their designs in the realities of real network, server, and storage platforms and quickly evaluate design tradeoffs visually rather than in spreadsheets.

**Why it exists**: The design phase of datacenter networking lives in hand-drawn diagrams, spreadsheets, and proprietary modeling systems. Many tools capture existing state or define a source of truth, but none support engineers in *building* topologies with constraints, best practices, and guided decision-making baked in.

## Design Principles

1. **Visual-first**: The UI is the product. Topology diagrams, rack elevations, and metrics dashboards are the primary way users think and make decisions. Every feature has a clear visual representation. All diagrams and visuals are richly interactive — hover, click, drill-down, and contextual detail are expected, not optional.

2. **Constraint-aware**: Designs are validated against real-world limits — port counts, power budgets, cable reach, oversubscription ratios, rack unit capacity (including management switch overhead). The tool warns you before you draw something that can't be built.

3. **Platform-grounded**: Tied to real hardware. Users pick actual switch models, server platforms, and optics — not just abstract boxes. This keeps designs honest and makes bill-of-materials generation possible.

4. **Scale-agnostic**: Works equally well for a 4-rack lab and a 100k-server hyperscale pod. The data model and UI don't assume a particular size. A small Clos fabric and a 5-stage Clos are both first-class citizens.

5. **Design-time focus**: Optimized for the planning phase. It doesn't deploy, monitor, or reconcile with live state. It produces artifacts (topology docs, part lists, rack elevations) that feed into other tools.

6. **Opinionated guidance**: Nudges users toward proven patterns (e.g., proper oversubscription ratios, recommended radix choices) without being rigid. Suggests best practices but lets you override them.

7. **Progressive disclosure**: Simple designs stay simple. You don't need to specify every detail upfront. Start with a topology sketch, then layer in platform choices, cabling, power, etc. as the design matures.

8. **Export-oriented**: The design isn't trapped in the tool. Clean export to design docs, topology diagrams, rack elevations, BOMs, and structured data that can feed NetBox, ordering systems, or other fabrik instances.

9. **Documentation as product**: fabrik is not just a tool — it is an education platform. Every concept in the UI links to embedded documentation that explains not only how to use the feature, but *why* the underlying datacenter design principle matters. The goal is for fabrik's documentation to become the world's foremost open reference on datacenter infrastructure design.

## Out of Scope

- **Not a NetBox replacement.** fabrik does not track live inventory or operational state.
- **Not a monitoring tool.** No runtime telemetry, alerting, or traffic analysis.
- **Not a config generator.** It does not produce switch configs, Ansible playbooks, or deployment artifacts.
- **Design artifacts only.** Output is design documentation, topologies, part lists, and rack elevations — not operational tooling.

## Personality

Helpful and guiding — the tool should drive users toward good design choices through constraints and suggestions, not just passively accept inputs. Otherwise, clean and professional. The UX should feel polished and purposeful, built on Material Design components.

## Core Concepts

### Fabric Types

- **Front-end fabric**: Traditional Clos topology connecting servers to the data center network (leaf/spine/super-spine).
- **Back-end fabric (RDMA)**: GPU-to-GPU network for AI/ML workloads. Initially scoped to a single pod (e.g., 8 racks with 2 leaf switches and 2 back-end switches forming a simple pod-scoped pair).
- Both fabric types are first-class and designed independently within the same topology.

### Logical vs. Physical Racks

High-power GPU racks may be "ToR-less" — a single leaf pair serves multiple physical racks. For example, 8 physical racks housing GPUs may appear as a single logical group since they share one leaf pair. fabrik models both the physical reality (8 racks) and the logical grouping (1 leaf pair serving the group).

### Management Network

Out-of-band management networks with in-rack management switches are modeled explicitly. Management switches consume rack units and power, contributing to rack capacity limits alongside compute and network gear.

### Resource Capacity

Aggregate resource capacity is a first-class metric: total vCPU, RAM, storage, and GPU count — viewable at every level of the hierarchy (rack, block, super-block, site). This gives architects an immediate answer to "how much compute does this design provide?"

### Embedded Knowledge Base

Every concept in fabrik — oversubscription, Clos stages, ECMP, RDMA fabrics, power budgets, rack design — links to embedded documentation that lives inside the app. This is not a help menu; it is a comprehensive, structured reference on datacenter infrastructure design. Contextual help buttons throughout the UI link directly to the relevant section. The knowledge base ships with the app and works offline, just like the rest of fabrik.

## Roadmap

### Phase 1 — Foundation
- Define Clos fabrics (stage count, radix, oversubscription) for both front-end and back-end networks
- Back-end (RDMA) fabric support: pod-scoped GPU clusters with dedicated back-end switch pairs
- ToR-less rack modeling: logical rack groups sharing a leaf pair
- Out-of-band management network modeling (management switches contributing to rack capacity)
- Device abstraction layer: generic devices (e.g., "48-port switch") for quick starts, plus user-defined custom models with specific specs (e.g., Cisco Nexus 9364C-GX2A)
- Rack abstraction: user-defined racks with RU count and power capacity (e.g., "42RU / 17kW")
- First-class power modeling: track power budget per rack and across the design
- First-class resource capacity tracking: vCPU, RAM, storage, GPU — aggregated per rack, block, super-block, and site
- Select real switch/optic/server platforms from a user-managed catalog
- Interactive topology visualization (hover, click, drill-down for contextual detail)
- Key metrics dashboard (port utilization, oversubscription, host count, power, resource capacity)
- Export design summary and artifacts
- Initial hardware catalog: Dell servers, Cisco Nexus 9300 switches
- UI built on latest Material Design components
- Embedded knowledge base with contextual help linking UI concepts to datacenter design documentation

### Phase 2 — Growth
- Rack elevation views
- Multi-site / multi-datacenter designs
- Datacenter floor modeling (rack placement, hot/cold aisle layout)
- Preset rack templates (e.g., "Base Rack 1" with predefined elevation and hardware)
- Back-end fabric expansion beyond single-pod scope
- Expanded hardware vendor catalog

### Phase 3 — Maturity
- Collaboration features
- NetBox export integration
- Advanced constraint modeling (cable reach, cooling zones)
- Community-shared hardware catalogs and rack templates
- Knowledge base grows into the definitive open reference on datacenter infrastructure design

## Prior Art & Constraints

- **Must be a web app** that runs completely locally (e.g., `npx fabrik run`) and works offline.
- **Local-first storage**: All state saved locally (SQLite or similar). No cloud dependency.
- **Portable designs**: Export/import between fabrik instances — ideally via a base64-encoded shareable link that imports in one click.
- **Frontend**: Angular with Angular Material (Material Design components). Angular chosen to enable potential internal use at Google (forked to work within their tooling). This is a strategic constraint.
- **Backend**: Go. Compiles to a single binary, serves the Angular frontend as static files, and aligns with Google's internal ecosystem. SQLite for local storage with migration tooling.
- **Distribution**: npm package wrapping pre-compiled Go binaries per platform. `npx fabrik run` extracts the right binary and starts the server.
- **Initial hardware support**: Dell servers and Cisco Nexus 9300 series switches as the first vendor catalog entries.
