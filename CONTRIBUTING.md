# Contributing to MEL Agent

## Setup

1. **Prerequisites**
   - Go 1.23+
   - Docker (for tests)
   - PostgreSQL (for local development)

2. **Install dependencies**
   ```bash
   go mod download
   cd web && pnpm install
   ```

## Running Tests

### Backend Tests
```bash
# Run all tests
go test ./...

# Run tests with database (uses testcontainers)
go test ./internal/api ./pkg/execution

# Run specific test
go test ./internal/api -v -run TestWorkerRegistration
```

### Frontend Tests
```bash
cd web
pnpm test
```

## Development Commands

### Backend
- `go run ./cmd/server` - Start API server
- `go run ./cmd/server worker -token <token>` - Start worker
- `go test ./...` - Run tests
- `go vet ./...` - Lint code

### Frontend
- `cd web && pnpm dev` - Start dev server
- `cd web && pnpm build` - Build for production
- `cd web && pnpm lint` - Lint code

## Database

Tests use testcontainers with automatic migration application. No manual database setup required for testing.

For local development:
```bash
docker compose up --build
```

## Code Quality

- Tests are required for new features
- Follow existing code patterns
- Run linters before submitting PRs