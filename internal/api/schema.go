package api

import (
  "net/http"

  "github.com/go-chi/chi/v5"
)

// getNodeTypeSchemaHandler returns a JSON Schema for the given node type.
func getNodeTypeSchemaHandler(w http.ResponseWriter, r *http.Request) {
  typ := chi.URLParam(r, "type")
  def := FindDefinition(typ)
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
    prop := map[string]interface{}{ }
    // map basic types
    switch p.Type {
    case "string": prop["type"] = "string"
    case "number": prop["type"] = "number"
    case "boolean": prop["type"] = "boolean"
    case "json": prop["type"] = "object"
    case "enum": prop["type"] = "string"
    default: prop["type"] = p.Type
    }
    if p.Description != "" {
      prop["description"] = p.Description
    }
    if p.Default != nil {
      prop["default"] = p.Default
    }
    if len(p.Options) > 0 {
      prop["enum"] = p.Options
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