# fabrik — Operating Manual

> This is the agentic knowledge base for fabrik. It captures architecture decisions,
> patterns, and development practices. It is the source of truth for how to work on
> this project.

## Project Overview

**fabrik** — A design-time tool for planning datacenter network topologies at any scale,
grounded in real hardware platforms with beautiful, metrics-rich visualization.

- **Frontend**: Angular 19+ with Angular Material (Material Design 3)
- **Backend**: Go 1.23+ with SQLite (local-first storage)
- **Distribution**: npm package wrapping pre-compiled Go binaries (`npx fabrik run`)
- **API**: REST (Go backend serves Angular static files + API endpoints)

## Build & Test

```bash
make build          # Build both server and frontend
make test           # Run all tests (Go + Angular)
make lint           # Lint all code (go vet + eslint)
make serve          # Run the dev server (Go backend + Angular dev server with proxy)
```

Individual targets:

```bash
# Server (Go)
cd server && go build ./cmd/fabrik    # Build server binary
cd server && go test ./...            # Run server tests
cd server && go vet ./...             # Lint server code

# Frontend (Angular)
cd frontend && npm run build          # Build frontend
cd frontend && npm test               # Run unit tests (Jest)
cd frontend && npm run lint           # Lint frontend (ESLint)
cd frontend && npx playwright test    # Run e2e tests
```

- ALWAYS run tests after code changes
- ALWAYS run build before committing
- NEVER commit if tests fail

## File Organization

```
server/                    # Go backend
  cmd/fabrik/              # Main entrypoint — starts HTTP server, serves frontend + API
  internal/
    api/                   # HTTP handlers (REST endpoints)
      handlers/            # Request handlers grouped by domain
      middleware/          # HTTP middleware (logging, CORS, error handling)
      routes.go            # Route registration
    models/                # Domain types — the core data model
    store/                 # SQLite repository layer (CRUD operations)
    migrations/            # SQL migration files (sequential, numbered)
    service/               # Business logic layer between handlers and store
  go.mod
  go.sum

frontend/                  # Angular application
  src/
    app/
      core/                # Singleton services, guards, interceptors, app-wide config
      shared/              # Reusable components, pipes, directives (no business logic)
      features/            # Feature modules, one per major UI area:
        topology/          # Clos fabric designer and topology visualization
        catalog/           # Hardware catalog management (devices, racks, optics)
        metrics/           # Metrics dashboard (oversubscription, power, capacity)
        knowledge/         # Embedded knowledge base / documentation viewer
      models/              # TypeScript interfaces mirroring server domain types
    assets/                # Static assets (icons, images)
    environments/          # Angular environment configs
  angular.json
  package.json
  tsconfig.json

docs/                      # Project documentation
  internal/                # Internal docs (vision, decisions, pipeline)
  knowledge/               # Datacenter design knowledge base content (Markdown)

dist/                      # Built artifacts (gitignored)
```

Rules:
- Go code follows standard Go project layout. All application code lives under `internal/`.
- Angular features are self-contained modules under `features/`. Each feature has its own
  components, services, and routes.
- Shared UI components go in `shared/`. They must not import from `features/`.
- `core/` contains app-wide singletons. It is imported once by `AppModule`.
- Domain types are defined in Go (`server/internal/models/`) and mirrored as TypeScript
  interfaces (`frontend/src/app/models/`). Keep them in sync.
- Knowledge base content lives as Markdown files in `docs/knowledge/`, rendered by the
  `knowledge` feature module.
- Keep files under 500 lines. If a file grows beyond this, split it.

## Architecture Patterns

1. **Layered backend**: Handlers → Service → Store → SQLite. Handlers parse HTTP requests
   and call services. Services contain business logic. Store handles database operations.
   No direct database access from handlers.

2. **Feature-based frontend**: Each major UI area is an Angular feature module with lazy
   loading. Features own their routes, components, and services. Cross-feature communication
   goes through core services.

3. **Domain model alignment**: Go structs in `models/` are the canonical data model.
   TypeScript interfaces in `frontend/src/app/models/` mirror them. API responses use
   JSON serialization of the Go structs. When the model changes, both sides must be updated.

4. **Migration-first schema changes**: Every database schema change requires a numbered
   migration file in `server/internal/migrations/`. Never modify the database schema
   directly. Migrations must be reversible (up + down).

5. **Contextual documentation**: Every user-facing concept links to a knowledge base
   article. When adding a new feature, add or update the corresponding `docs/knowledge/`
   article. Help buttons in the UI route to the knowledge viewer.

6. **Constraint validation**: Design constraints (port counts, power budgets, RU capacity)
   are enforced in the service layer, not in handlers or the frontend. The frontend
   displays warnings/errors from the service layer. This ensures constraints are consistent
   regardless of how the API is called.

## Testing Standards

### Go (server)
- Test files live next to the code they test: `foo.go` → `foo_test.go`
- Use table-driven tests for functions with multiple input/output combinations
- Use `testing.T` directly — no external test frameworks
- Mock external dependencies via interfaces, not concrete types
- Database tests use an in-memory SQLite instance
- Aim for high coverage of the service and store layers; handlers are tested via HTTP

### Angular (frontend)
- Unit tests use Jest with Angular Testing Library
- Test files live next to components: `foo.component.ts` → `foo.component.spec.ts`
- Test behavior, not implementation — query by role/label, not CSS selectors
- E2e tests use Playwright and live in `frontend/e2e/`
- E2e tests run against a real backend with a test database
- Both unit and e2e tests are required for new features

## Error Handling

### Go
- Wrap errors with context: `fmt.Errorf("failed to load design %d: %w", id, err)`
- Service layer returns domain errors (e.g., `ErrNotFound`, `ErrConstraintViolation`)
- Handlers map domain errors to HTTP status codes
- Use structured logging (slog) with consistent fields: `slog.Error("msg", "err", err, "designID", id)`

### Angular
- HTTP errors are caught by a global error interceptor in `core/`
- User-facing errors display as Material snackbar notifications
- Form validation errors display inline using Angular Material form field errors
- Console logging in development only; no `console.log` in production code

## Development Workflow

- **main is sacred.** All changes go through PRs. No direct pushes.
- Branch naming: `feat/`, `fix/`, `chore/`, `docs/` prefixes
- Conventional commits: `type: description` format (lowercase, no scope, no trailing period)
- PRs require CI passing (both `server` and `frontend` jobs)
- Keep PRs focused — one feature or fix per PR

## Autonomous Development Workflow

An event-driven GitHub Actions pipeline that autonomously implements issues end-to-end.
For a comprehensive architecture deep-dive with diagrams, see
[docs/internal/autodev-pipeline.md](docs/internal/autodev-pipeline.md).

### How it works

Four workflows form the core loop, plus a weekly audit:

1. **`autodev-dispatch`** — Runs on a configurable cron. Picks the highest-priority
   `backlog/ready` issue, labels it `agent/implementing`, and triggers the implement workflow.
2. **`autodev-implement`** — Checks out the base branch, creates a feature branch, runs
   the agent to implement the issue, pushes, and opens a PR. After creating the PR, the
   workflow polls for Copilot review and dispatches `autodev-review-fix`.
3. **`autodev-review-fix`** — Phased review pipeline: Copilot phase (up to N iterations)
   → Claude phase → done.
4. **`claude-code-review`** — Triggered by `agent/review-claude` label or `@claude` mention.
5. **`autodev-audit`** — Weekly pipeline health report filed as a GitHub issue.

### Labels

| Label | Meaning |
|-------|---------|
| `backlog/ready` | Issue is ready for autonomous implementation |
| `agent/implementing` | Issue is currently being implemented by an agent |
| `agent/review-copilot` | Agent is addressing Copilot review feedback |
| `agent/review-claude` | Agent is addressing Claude review feedback |
| `human/blocked` | Agent hit a limit and needs human intervention |
| `via/actions` | PR created by GitHub Actions pipeline |
| `via/autodev` | PR created by /autodev CLI skill |

### Secrets required

| Secret | Purpose |
|--------|---------|
| `CLAUDE_CODE_OAUTH_TOKEN` | OAuth token for Claude Code agent execution |
| `APP_ID` + `APP_PRIVATE_KEY` | GitHub App credentials for push/PR operations |

## GitHub Issue Workflow

When creating a PR that implements a GitHub issue:

1. Read the original issue and extract acceptance criteria
2. Verify each criterion is satisfied by the implementation
3. Document verification in the PR body under "Acceptance Criteria"
4. Use closing keywords (`Closes #N`, `Fixes #N`) for auto-close on merge

## Key Files

| File | Purpose |
|------|---------|
| `CLAUDE.md` | This file — project operating manual |
| `forge.toml` | Pipeline configuration |
| `docs/internal/VISION.md` | Product vision and design principles |
| `docs/internal/DECISIONS.md` | Architectural decision records |
| `server/cmd/fabrik/main.go` | Go server entrypoint |
| `server/internal/api/routes.go` | API route registration |
| `server/internal/models/` | Canonical domain types |
| `server/internal/migrations/` | Database migration files |
| `frontend/src/app/app.routes.ts` | Angular route definitions |
| `frontend/angular.json` | Angular CLI configuration |
| `Makefile` | Top-level build orchestration |
