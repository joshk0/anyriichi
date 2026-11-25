# Convenience targets. Node tooling runs through corepack (no global pnpm needed).
PNPM := COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm

.PHONY: install dev build test lint typecheck worker-build worker-test worker-run worker-deploy check

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

worker-test:
	cd worker && go vet ./... && go test ./...

worker-run:
	cd worker && go run ./cmd/worker

check: typecheck lint test

worker-deploy:
	deploy/worker/deploy.sh $(if $(ENV),--env,)
