package api

import (
	"net/http"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/go-chi/chi/v5"
)

// getNodeTypeSchemaHandler returns a JSON Schema for the given node type.
func getNodeTypeSchemaHandler(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	def := api.FindDefinition(typ)
	if def == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "node type not found"})
		return
	}
	meta := def.Meta()
	// Build JSON Schema
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title":   meta.Label,
		"type":    "object",
	}
	props := map[string]interface{}{}
	required := []string{}
	for _, p := range meta.Parameters {
		prop := map[string]interface{}{}
		// map basic types
		switch p.Type {
		case "string":
			prop["type"] = "string"
		case "number":
			prop["type"] = "number"
		case "boolean":
			prop["type"] = "boolean"
		case "json":
			prop["type"] = "object"
		case "enum":
			prop["type"] = "string"
		default:
			prop["type"] = p.Type
		}
		// dynamic enum for LLM connectionId
		if meta.Type == "llm" && p.Name == "connectionId" {
			// list all valid LLM connections
			rows, err := db.DB.Query(
				`SELECT c.id, c.name FROM connections c JOIN integrations i ON c.integration_id = i.id WHERE i.category = $1 AND c.status = 'valid'`,
				"llm_provider",
			)
			if err == nil {
				var ids []string
				var names []string
				for rows.Next() {
					var id, name string
					if err := rows.Scan(&id, &name); err == nil {
						ids = append(ids, id)
						names = append(names, name)
					}
				}
				rows.Close()
				prop["enum"] = ids
				prop["x-enumNames"] = names
			}
		} else if meta.Type == "baserow" && p.Name == "connectionId" {
			// list all valid Baserow connections
			rows, err := db.DB.Query(
				`SELECT c.id, c.name FROM connections c JOIN integrations i ON c.integration_id = i.id WHERE i.name = $1 AND c.status = 'valid'`,
				"baserow",
			)
			if err == nil {
				var ids []string
				var names []string
				for rows.Next() {
					var id, name string
					if err := rows.Scan(&id, &name); err == nil {
						ids = append(ids, id)
						names = append(names, name)
					}
				}
				rows.Close()
				prop["enum"] = ids
				prop["x-enumNames"] = names
			}
		} else if meta.Type == "baserow" && (p.Name == "databaseId" || p.Name == "tableId") {
			// For database and table IDs, we'll need dynamic loading via API endpoints
			// Set up the parameter to indicate it needs dynamic loading
			prop["x-dynamicEnum"] = map[string]interface{}{
				"type": p.Name,
				"dependsOn": "connectionId",
			}
			if p.Name == "tableId" {
				prop["x-dynamicEnum"].(map[string]interface{})["dependsOn"] = []string{"connectionId", "databaseId"}
			}
		} else {
			if len(p.Options) > 0 {
				prop["enum"] = p.Options
			}
		}
		// common fields
		if p.Description != "" {
			prop["description"] = p.Description
		}
		if p.Default != nil {
			prop["default"] = p.Default
		}
		if p.VisibilityCondition != "" {
			prop["x-visibilityCondition"] = p.VisibilityCondition
		}
		if len(p.Validators) > 0 {
			prop["x-validators"] = p.Validators
		}
		props[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}
	schema["properties"] = props
	if len(required) > 0 {
		schema["required"] = required
	}
	writeJSON(w, http.StatusOK, schema)
}
