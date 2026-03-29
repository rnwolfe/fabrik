.PHONY: build test lint serve setup clean

# Top-level targets
build: build-server build-frontend
test: test-server test-frontend
lint: lint-server lint-frontend

# First-time setup: install all dependencies
setup:
	cd frontend && npm install

# Server (Go)
build-server: copy-knowledge-docs
	cd server && go build -o ../bin/fabrik ./cmd/fabrik

# Copy knowledge docs into server module for embedding.
copy-knowledge-docs:
	rm -rf server/cmd/fabrik/docs/knowledge
	mkdir -p server/cmd/fabrik/docs
	cp -r docs/knowledge server/cmd/fabrik/docs/knowledge

test-server: copy-knowledge-docs
	cd server && go test ./...

lint-server:
	cd server && go vet ./...

# Frontend (React + Vite)
build-frontend:
	cd frontend && npm run build

test-frontend:
	cd frontend && npm test -- --run

lint-frontend:
	cd frontend && npm run lint

# E2E tests (requires running server)
test-e2e:
	cd frontend && npx playwright test

# Development — hot reload via air (Go) + Vite (frontend)
serve: setup
	@echo "Starting fabrik dev server..."
	@echo "  Backend:  http://localhost:8080  (hot reload via air)"
	@echo "  Frontend: http://localhost:4200  (hot reload via Vite)"
	@trap 'kill 0' INT; \
	  air & \
	  (cd frontend && npm run dev) & \
	  wait

clean:
	rm -rf bin/ dist/ frontend/dist/ server/cmd/fabrik/docs/
