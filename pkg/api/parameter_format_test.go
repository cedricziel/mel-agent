package api

import "testing"

func TestWithFormatSetsSchemaFormat(t *testing.T) {
	param := NewStringParameter("expression", "Expression", true).
		WithDescription("a template").
		WithFormat("template")

	if param.JSONSchema == nil {
		t.Fatal("expected JSONSchema to be set")
	}
	if param.JSONSchema.Format != "template" {
		t.Errorf("expected format template, got %q", param.JSONSchema.Format)
	}
	if param.JSONSchema.Type != "string" {
		t.Errorf("expected type string, got %q", param.JSONSchema.Type)
	}

	schema := param.ToJSONSchema()
	if schema.Format != "template" {
		t.Errorf("expected generated schema format template, got %q", schema.Format)
	}
	if schema.Description != "a template" {
		t.Errorf("expected description to carry over, got %q", schema.Description)
	}
}

func TestWithFormatPreservesExistingSchema(t *testing.T) {
	param := NewStringParameter("code", "Code", true)
	param.JSONSchema = &JSONSchema{Type: "string", Title: "Code"}
	param = param.WithFormat("code")

	if param.JSONSchema.Title != "Code" {
		t.Errorf("expected existing schema fields to be preserved, got %q", param.JSONSchema.Title)
	}
	if param.JSONSchema.Format != "code" {
		t.Errorf("expected format code, got %q", param.JSONSchema.Format)
	}
}

func TestWithFormatDoesNotMutateOriginal(t *testing.T) {
	base := NewStringParameter("expression", "Expression", true)
	base.JSONSchema = &JSONSchema{Type: "string"}

	derived := base.WithFormat("template")

	if base.JSONSchema.Format != "" {
		t.Errorf("expected original schema to be untouched, got %q", base.JSONSchema.Format)
	}
	if derived.JSONSchema.Format != "template" {
		t.Errorf("expected derived format template, got %q", derived.JSONSchema.Format)
	}
}
