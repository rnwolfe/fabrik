.PHONY: build test lint serve clean

# Top-level targets
build: build-server build-frontend
test: test-server test-frontend
lint: lint-server lint-frontend

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

# Frontend (Angular)
build-frontend:
	cd frontend && npm run build

test-frontend:
	cd frontend && npm test -- --watch=false

lint-frontend:
	cd frontend && npm run lint

# E2E tests (requires running server)
test-e2e:
	cd frontend && npx playwright test

# Development
serve:
	@echo "Starting fabrik dev server..."
	@echo "TODO: run Go backend + Angular dev server with proxy"

clean:
	rm -rf bin/ dist/ frontend/dist/ frontend/.angular/ server/cmd/fabrik/docs/
