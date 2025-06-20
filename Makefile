.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## Generate code from OpenAPI spec
	cd internal/api && go generate

.PHONY: build
build: ## Build the server binary
	go build -o bin/server ./cmd/server

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: lint
lint: ## Run linters
	go vet ./...
	go fmt ./...

.PHONY: run
run: ## Run the server
	go run ./cmd/server server

.PHONY: run-api-server
run-api-server: ## Run API server only (no embedded workers)
	go run ./cmd/server api-server

.PHONY: run-worker
run-worker: ## Run a remote worker
	go run ./cmd/server worker --token $(MEL_WORKER_TOKEN)

.PHONY: docker-build
docker-build: ## Build Docker image
	docker compose build

.PHONY: docker-up
docker-up: ## Start services with Docker Compose
	docker compose up --build

.PHONY: docker-down
docker-down: ## Stop Docker Compose services
	docker compose down

.PHONY: openapi-validate
openapi-validate: ## Validate OpenAPI 3.0 specification
	@if command -v redocly 2>&1 >/dev/null; then \
		redocly lint api/openapi.yaml; \
	elif command -v npx 2>&1 >/dev/null; then \
		npx @redocly/cli lint api/openapi.yaml; \
	else \
		echo "OpenAPI validator not found. Install with: npm install -g @redocly/cli"; \
		echo "Or use npx: npx @redocly/cli lint api/openapi.yaml"; \
	fi

.PHONY: openapi-validate-strict
openapi-validate-strict: ## Validate OpenAPI spec with strict rules (fail on warnings)
	@if command -v redocly 2>&1 >/dev/null; then \
		redocly lint --max-problems 0 api/openapi.yaml; \
	elif command -v npx 2>&1 >/dev/null; then \
		npx @redocly/cli lint --max-problems 0 api/openapi.yaml; \
	else \
		echo "OpenAPI validator not found. Install with: npm install -g @redocly/cli"; \
		echo "Or use npx: npx @redocly/cli lint --max-problems 0 api/openapi.yaml"; \
	fi

.PHONY: openapi-docs
openapi-docs: ## Serve OpenAPI 3.0 documentation
	@if command -v redocly 2>&1 >/dev/null; then \
		redocly preview-docs api/openapi.yaml; \
	elif command -v npx 2>&1 >/dev/null; then \
		npx @redocly/cli preview-docs api/openapi.yaml; \
	else \
		echo "OpenAPI docs server not found. Install with: npm install -g @redocly/cli"; \
		echo "Or use npx: npx @redocly/cli preview-docs api/openapi.yaml"; \
	fi