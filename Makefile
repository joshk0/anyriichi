# Convenience targets. Node tooling runs through corepack (no global pnpm needed).
PNPM := COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm

.PHONY: install dev build test lint typecheck worker-build worker-test worker-run worker-deploy sqlc db-push rules-sync check

install:
	$(PNPM) install

dev:
	$(PNPM) --filter @anyriichi/web dev

build:
	$(PNPM) -r build

test: worker-test
	$(PNPM) -r test

lint:
	$(PNPM) -r lint

typecheck:
	$(PNPM) -r typecheck

worker-build:
	cd worker && go build -o ../bin/worker ./cmd/worker

# DB-backed tests run when TEST_DATABASE_URL points at a disposable Postgres (local or CI), e.g.
# TEST_DATABASE_URL=postgres://postgres@localhost:5432/postgres?sslmode=disable make worker-test
worker-test:
	cd worker && go vet ./... && go test ./...

worker-run:
	cd worker && go run ./cmd/worker

check: typecheck lint test

worker-deploy:
	deploy/worker/deploy.sh $(if $(ENV),--env,)

# Regenerate typed queries (worker/internal/db) from supabase/migrations + worker/internal/db/queries.
sqlc:
	cd worker && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

# Apply supabase/migrations to the linked hosted project.
db-push:
	npx -y supabase@2.62.10 db push

# Copy shared rules JSON into the Go module (go:embed cannot reach outside worker/).
rules-sync:
	cp packages/shared/rules/*.json worker/internal/rules/data/
