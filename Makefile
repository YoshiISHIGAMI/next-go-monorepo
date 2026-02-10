.PHONY: dev down logs test lint generate clean

# Development
dev:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

rebuild:
	docker compose up -d --build

# Testing
test: test-next test-go

test-next:
	pnpm --filter next-app test:run

test-go:
	cd apps/go-api && go test -v ./...

# Linting
lint: lint-next lint-go

lint-next:
	pnpm biome:check

lint-go:
	cd apps/go-api && gofmt -l . && go vet ./...

# Code generation
generate:
	pnpm --filter next-app generate:types

# Formatting
fmt: fmt-next fmt-go

fmt-next:
	pnpm biome:fix

fmt-go:
	cd apps/go-api && go fmt ./...

# Database
db-reset:
	docker compose down -v
	docker compose up -d db
	@echo "Waiting for DB..."
	@sleep 3
	docker compose up -d api

# Cleanup
clean:
	docker compose down -v
	rm -rf apps/next-app/.next
	rm -rf apps/next-app/node_modules
	rm -rf apps/go-api/tmp

# Help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development:"
	@echo "  dev       Start Docker (API + DB)"
	@echo "  down      Stop Docker"
	@echo "  logs      Show logs"
	@echo "  rebuild   Rebuild and start Docker"
	@echo ""
	@echo "Testing:"
	@echo "  test      Run all tests"
	@echo "  test-next Run Next.js tests"
	@echo "  test-go   Run Go tests"
	@echo ""
	@echo "Linting:"
	@echo "  lint      Run all linters"
	@echo "  lint-next Run Biome check"
	@echo "  lint-go   Run go fmt + go vet"
	@echo ""
	@echo "Other:"
	@echo "  generate  Generate TypeScript types from OpenAPI"
	@echo "  fmt       Format all code"
	@echo "  db-reset  Reset database"
	@echo "  clean     Remove all generated files"
