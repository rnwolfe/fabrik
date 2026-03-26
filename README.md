# fabrik

> Design datacenter network topologies at any scale — from a few racks to hyperscale — grounded in real hardware platforms.

**fabrik** helps datacenter network architects plan Clos fabric topologies in fine detail.
Define racks, select hardware platforms, wire up front-end and back-end fabrics, and
instantly see meaningful metrics — all through an interactive visual interface that teaches
you datacenter design as you go.

## Quick Start

```bash
npx fabrik run
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Features

- **Clos fabric design** — Define multi-stage fabrics with configurable radix and oversubscription
- **Front-end and back-end fabrics** — Model both traditional and RDMA/GPU networks
- **Real hardware catalog** — Use actual switch, server, and optic models (Dell, Cisco Nexus 9300)
- **Rack modeling** — Define racks with RU count, power capacity, and management switches
- **Resource capacity** — Track vCPU, RAM, storage, and GPU across the design
- **Power modeling** — First-class power budget tracking per rack and site-wide
- **Interactive visualization** — Topology diagrams with hover, click, and drill-down
- **Embedded knowledge base** — Contextual documentation on datacenter design principles
- **Local-first** — All data stored locally in SQLite. Works offline. No cloud dependency.
- **Portable designs** — Export/import between fabrik instances

## Development

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Node.js 22+](https://nodejs.org/)
- [Angular CLI](https://angular.dev/tools/cli)

### Build & Test

```bash
make build     # Build server and frontend
make test      # Run all tests
make lint      # Lint all code
make serve     # Start dev server
```

See [CLAUDE.md](CLAUDE.md) for architecture details and development workflow.

## License

MIT
