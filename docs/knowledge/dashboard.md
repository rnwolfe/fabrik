---
title: "Dashboard & Design Templates"
category: "getting-started"
tags: [dashboard, templates, clos, design]
---

# Dashboard & Design Templates

The fabrik dashboard is your home screen. From here you can create and open
designs, browse the hardware catalog, and access the metrics and knowledge base.

## Creating a Design

Click **New Design** to open the creation dialog. Give your design a name and
an optional description, then choose a starting template.

## Design Templates

Templates scaffold an initial site → superblock → block hierarchy so you can
start configuring a realistic topology immediately. Every template creates
**named** hierarchy objects (e.g. "Site 1", "Pod A") and does **not**
pre-assign any leaf or spine device models — those are set per-block after
creation.

### Blank

No hierarchy is created. You get an empty canvas and can add sites, pods
(super-blocks), and blocks manually. This is equivalent to the original
behavior before templates were introduced.

**When to use:** You have a bespoke topology in mind, or you want full control
over every naming and structural decision from the start.

### 2-stage Clos

Creates the simplest production-grade Clos topology:

```
Site 1
└── Pod A
    └── Block 1   (leaf–spine fabric)
```

A 2-stage Clos is a single-tier spine layer connecting all leaves. It is
the most common design for a single-pod, single-plane datacenter.

**When to use:** One pod, one failure domain, moderate scale.

### 3-stage Clos

Creates the same structural skeleton as a 2-stage Clos but signals that a
super-spine (aggregation) tier will be configured at the superblock level:

```
Site 1
└── Pod A  (super-spine aggregation pod)
    └── Block 1
```

After creation, assign aggregation models to the Pod to introduce the
super-spine tier and connect multiple blocks through it.

**When to use:** You need to scale beyond a single leaf–spine plane, or you
anticipate adding more blocks that must communicate at low latency within
the pod.

### Pod-based fabric

Creates a multi-pod topology with two independent pods and four blocks total:

```
Site 1
├── Pod A
│   ├── Block 1
│   └── Block 2
└── Pod B
    ├── Block 3
    └── Block 4
```

Each pod is an isolated failure domain. Pods can be interconnected via
DCI or a separate super-spine layer later.

**When to use:** Large-scale datacenters where fault isolation between
pods is a first-class requirement.

## Partial Scaffold Failures

If a template scaffold fails partway through (for example, the API is
unreachable), fabrik will:

1. Navigate you to the design view with whatever hierarchy was successfully
   created.
2. Show a non-blocking notification describing the partial failure.

You can manually add the missing hierarchy objects from the design view.

## Related Topics

- [Design Hierarchy](hierarchy.md) — the five-level site → block model
- [Clos Topology](networking/clos-topology.md) — Clos fabric fundamentals
- [Block Aggregation](block-aggregation.md) — assigning spine/leaf models to blocks
