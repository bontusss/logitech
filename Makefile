# Project root Makefile.
# Run make from the project root.
#
#   make s_up     start the full stack  (postgres + migrations + go server + nuxt dev)
#   make s_down   stop everything
#   make db_up    start postgres via docker compose
#   make db_down  stop postgres
#   make m_up     run migrations up (all pending)
#   make m_down   run migrations down (1 step, or N: make m_down N=2)
#   make run-be   run go server alone
#   make run-fe   run nuxt dev alone
#   make build    build embedded binary (nuxt generate -> go:embed -> go build)
#   make swagger  generate swagger / openapi documentation
#   make sqlc     generate sqlc code from queries and schema
#
# Health check: curl http://localhost:8080/health
# Swagger docs: http://localhost:8080/swagger/index.html
#
# Copy .env.example to .env to override defaults.

BE_PORT           ?= 8080
FE_PORT           ?= 3000
DATABASE_URL      ?= postgres://app:app@localhost:5432/appdb?sslmode=disable
RUN_DIR           := backend/.run

## ---- stack (s_up / s_down) ---------------------------------------------

s_up: db_up m_up ## postgres + migrations + go server (bg) + nuxt dev (bg)
	@mkdir -p $(RUN_DIR)
	@echo "==> starting go server on :$(BE_PORT) (logs: $(RUN_DIR)/be.log)"
	@(cd backend && PORT=$(BE_PORT) DATABASE_URL=$(DATABASE_URL) go run ./cmd/server > .run/be.log 2>&1 & echo $$! > .run/be.pid)
	@echo "==> starting nuxt dev on :$(FE_PORT) (logs: $(RUN_DIR)/fe.log)"
	@(cd frontend && FE_PORT=$(FE_PORT) npm run dev > ../$(RUN_DIR)/fe.log 2>&1 & echo $$! > ../$(RUN_DIR)/fe.pid)
	@$(MAKE) --no-print-directory wait-health
	@echo "   go server: $(RUN_DIR)/be.log   pid: $$(cat $(RUN_DIR)/be.pid)"
	@echo "   nuxt dev : $(RUN_DIR)/fe.log   pid: $$(cat $(RUN_DIR)/fe.pid)"
	@echo "==> stack up: Go http://localhost:$(BE_PORT) (serves /health) — Nuxt http://localhost:$(FE_PORT)"

s_down: ## stop go server, nuxt dev, and postgres
	@echo "==> stopping processes"
	@for pid_file in $(RUN_DIR)/be.pid $(RUN_DIR)/fe.pid; do \
	  if [ -f $$pid_file ]; then \
	    pkill -TERM -P $$(cat $$pid_file) >/dev/null 2>&1; kill -TERM $$(cat $$pid_file) >/dev/null 2>&1 || true; rm -f $$pid_file; \
	  fi; \
	done
	@pkill -f "go run ./cmd/server" >/dev/null 2>&1 || true
	@pkill -f "nuxt dev" >/dev/null 2>&1 || true
	@$(MAKE) --no-print-directory db_down
	@echo "==> stack down"

## ---- database (db_up / db_down) -----------------------------------------

db_up: ## start postgres via docker compose and wait until healthy
	@echo "==> starting postgres (docker compose)"
	@docker compose -f backend/docker-compose.yml up -d db
	@echo -n "==> waiting for postgres ."
	@until [ "$$(docker inspect --format='{{.State.Health.Status}}' app_postgres 2>/dev/null)" = "healthy" ]; do echo -n "."; sleep 1; done
	@echo " healthy"

db_down: ## stop postgres container (data volume is kept)
	@echo "==> stopping postgres"
	@docker compose -f backend/docker-compose.yml stop db 2>/dev/null || true

## ---- migrations (m_up / m_down) ------------------------------------------

m_up: ## apply all pending migrations
	@echo "==> migrations up"
	@cd backend && DATABASE_URL=$(DATABASE_URL) go run ./cmd/migrate up

m_down: ## revert 1 migration (or N: make m_down N=2)
	@echo "==> migrations down"
	@cd backend && DATABASE_URL=$(DATABASE_URL) go run ./cmd/migrate down $(N)

## ---- independent servers ---------------------------------------------------

run-be: ## run the go server alone (foreground)
	cd backend && PORT=$(BE_PORT) DATABASE_URL=$(DATABASE_URL) go run ./cmd/server

run-fe: ## run the nuxt dev server alone (foreground)
	cd frontend && npm run dev -- --port $(FE_PORT) --host

## ---- build (embed) ---------------------------------------------------------

build: fe-generate embed-copy ## nuxt generate -> copy into internal/web/dist -> go:embed build
	@rm -f backend/internal/web/dist/.gitkeep
	@cd backend && go build -trimpath -o bin/server ./cmd/server
	@echo "==> done: backend/bin/server"

fe-generate:
	@cd frontend && npm ci --silent 2>/dev/null || npm install
	@cd frontend && npx nuxt generate

embed-copy:
	@rm -rf backend/internal/web/dist
	@cp -R frontend/.output/public backend/internal/web/dist
	@find backend/internal/web/dist -type f -exec touch {} +

## ---- docs (swagger) ---------------------------------------------------------

swagger: ## generate swagger / openapi documentation
	@echo "==> generating swagger docs"
	@cd backend && swag init -g main.go -d cmd/server,internal/handler -o docs

## ---- sqlc -------------------------------------------------------------------

sqlc: ## generate sqlc code from queries and schema
	@echo "==> generating sqlc code"
	@cd backend && sqlc generate

clean: ## remove build artifacts and runtime files
	@rm -rf backend/bin $(RUN_DIR)
	@rm -rf backend/internal/web/dist && mkdir -p backend/internal/web/dist && touch backend/internal/web/dist/.gitkeep

wait-health:
	@echo -n "==> waiting for /health ."
	@until curl -sf http://localhost:$(BE_PORT)/health >/dev/null 2>&1; do echo -n "."; sleep 1; done
	@echo " up"

## ---- misc -----------------------------------------------------------

help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

.DEFAULT_GOAL := help

.PHONY: s_up s_down db_up db_down m_up m_down run-be run-fe build clean help swagger sqlc wait-health
