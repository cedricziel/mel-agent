# MEL Agent Project Context

## Project Overview
AI Agents SaaS platform with Go backend and React frontend. This is a monorepo that provides a visual workflow builder for AI agents with support for various node types, triggers, and integrations.

## Architecture
- **Backend**: Go with Chi router, PostgreSQL database
- **Frontend**: React + Vite + Tailwind CSS
- **Database**: PostgreSQL with migrations
- **Containerization**: Docker Compose for local development

## Development Commands

### Backend (Go)
- **Run server**: `go run ./cmd/server`
- **Run with Docker**: `docker compose up --build`
- **Test**: `go test ./...`
- **Build**: `go build ./cmd/server`
- **Lint**: `go vet ./...`
- **Run go fmt from time**: `go fmt ./...`

### Frontend (React)
- **Install dependencies**: `cd web && pnpm install`
- **Start dev server**: `cd web && pnpm dev`
- **Build**: `cd web && pnpm build`
- **Lint**: `cd web && pnpm lint`
- **Reformat the javascript using pnpm format from time to time**

### Database
- **Connection**: `postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable`
- **Migrations**: Located in `migrations/` directory
- **Docker setup**: Postgres runs on localhost:5432

## Development Principles
- Our goal is to create a well-tested project. Always be mindful of creating reasonably well-sized components and testing them. Create e2e tests with Cypress where applicable
- Remember to keep components at a reasonable size and add new tests for new components and functionality

## Development Guidelines
- When committing code via git, never mention claude