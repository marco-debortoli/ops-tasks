# Uses a local Go toolchain when there is one, otherwise runs Go in Docker.
DOCKER_GO = docker run --rm $(if $(shell test -t 0 && echo y),-it) -u $$(id -u):$$(id -g) --network host \
	-v $(CURDIR)/backend:/src -v $(CURDIR)/.cache:/cache -w /src \
	-e GOCACHE=/cache/go-build -e GOMODCACHE=/cache/gomod -e GOFLAGS=-modcacherw \
	-e DATABASE_URL -e APP_TIMEZONE -e ADDR -e STATIC_DIR \
	golang:1.27-alpine go
GO ?= $(if $(shell command -v go 2>/dev/null),cd backend && go,mkdir -p .cache && $(DOCKER_GO))

export DATABASE_URL ?= postgres://ops:ops@localhost:5432/ops_tasks?sslmode=disable

.PHONY: db api web test check build up down

db: ## start Postgres in the background
	docker compose up -d db

api: ## run the Go API on :8080 against the compose database
	$(GO) run ./cmd/server

web: ## run the SvelteKit dev server on :5173 (proxies /api to :8080)
	cd web && npm run dev

test: ## backend unit tests
	$(GO) test ./...

check: ## vet, test and type-check everything
	$(GO) vet ./...
	$(GO) test ./...
	cd web && npm run check

build: ## build the production image
	docker compose build

up: ## run the whole stack (app on :8080)
	docker compose up -d --build

down:
	docker compose down
