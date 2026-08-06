package api

import (
	"database/sql"

	"github.com/cedricziel/mel-agent/pkg/execution"
)

// OpenAPIHandlers implements the StrictServerInterface for type-safe OpenAPI handling
type OpenAPIHandlers struct {
	db            *sql.DB
	engine        execution.ExecutionEngine
	openAIAPIKey  string
}

// NewOpenAPIHandlers creates a new OpenAPI handlers instance
func NewOpenAPIHandlers(database *sql.DB, engine execution.ExecutionEngine, openAIAPIKey string) *OpenAPIHandlers {
	return &OpenAPIHandlers{
		db:           database,
		engine:       engine,
		openAIAPIKey: openAIAPIKey,
	}
}
