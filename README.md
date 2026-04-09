# fabrik

> [!WARNING]
> fabrik is in early development. APIs, data models, and the database schema change frequently and without notice. Expect breaking changes between versions.

> Design datacenter network topologies at any scale — from a few racks to hyperscale — grounded in real hardware platforms.

**fabrik** helps datacenter network architects plan Clos fabric topologies in fine detail.
Define a site hierarchy, assign hardware platforms to spine and leaf roles, wire up
front-end and back-end fabrics, and instantly see meaningful metrics — all through an
interactive visual interface that teaches you datacenter design as you go.

## Running from Source

fabrik is not yet distributed as a binary or npm package. You'll need to build it yourself.

**Prerequisites**

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 24+](https://nodejs.org/)
- [air](https://github.com/air-verse/air) — Go hot-reload for the dev server (`go install github.com/air-verse/air@latest`)

```bash
git clone https://github.com/rnwolfe/fabrik
cd fabrik
make setup   # install frontend dependencies
make serve   # start backend + frontend dev servers
```

- Backend: [http://localhost:8080](http://localhost:8080)
- Frontend: [http://localhost:4200](http://localhost:4200) (hot reload via Vite)

## Features

- **Clos fabric design** — Define multi-stage fabrics with configurable radix and oversubscription; topology emerges automatically from the hierarchy
- **Site hierarchy** — Organize designs into sites and superblocks with a tree UI; manage spine and leaf assignments per block
- **Real hardware catalog** — Use actual switch, server, and optic models (Dell, Cisco Nexus 9300) with role suggestions (leaf / spine / super-spine), port-group summaries, and port-speed filtering
- **Device creation in-flow** — Add new catalog devices without leaving the design canvas
- **Rack modeling** — Define racks with RU count, power capacity, and management switches; visualize server placement
- **Inline metrics** — Core metrics (oversubscription, port utilization, power) displayed directly on the design canvas
- **Metrics dashboard** — Per-tier oversubscription, power, and resource capacity breakdown
- **Structured validation** — Design rule violations surface as typed error codes with links back to the offending block
- **Interactive visualization** — Topology diagrams with hover, click, and drill-down
- **Embedded knowledge base** — Contextual documentation on datacenter design principles with deep-link help buttons throughout the UI
- **Local-first** — All data stored locally in SQLite. Works offline. No cloud dependency.

## Development

### Stack

- **Backend**: Go 1.25+ with SQLite (local-first storage)
- **Frontend**: React 19 + Vite

### Build & Test

```bash
make build     # Build server and frontend
make test      # Run all tests
make lint      # Lint all code
make serve     # Start dev server (air + Vite, hot reload)
```

See [CLAUDE.md](CLAUDE.md) for architecture details and development workflow.

## License

MIT
