package transform

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func newTestEnvelope(ctx api.ExecutionContext, node api.Node, input interface{}) *api.Envelope[interface{}] {
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	return core.NewEnvelope(input, trace)
}

func TestTransformDefinition_Template(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{"expression": "Hello, {{.input.name}}!"}}

	input := map[string]interface{}{"name": "Alice"}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	if out.Data != "Hello, Alice!" {
		t.Errorf("unexpected result: %v", out.Data)
	}
	if out.DataType != "string" {
		t.Errorf("expected DataType string, got %s", out.DataType)
	}
}

func TestTransformDefinition_Variables(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "agent", RunID: "run", Variables: map[string]interface{}{"role": "admin"}}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{"expression": "{{.vars.role}}-{{.input}}"}}

	env := newTestEnvelope(ctx, node, "data")

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	if out.Data != "admin-data" {
		t.Errorf("unexpected result: %v", out.Data)
	}
}

func TestTransformDefinition_MissingExpression(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "agent", RunID: "run"}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{}}

	env := newTestEnvelope(ctx, node, "data")

	_, err := def.ExecuteEnvelope(ctx, node, env)
	if err == nil {
		t.Fatal("expected error for missing expression")
	}
	if len(env.Errors) == 0 {
		t.Error("expected error recorded on envelope")
	}
}

func TestTransformDefinition_InvalidTemplate(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "agent", RunID: "run"}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{"expression": "{{ invalid"}}

	env := newTestEnvelope(ctx, node, "data")

	_, err := def.ExecuteEnvelope(ctx, node, env)
	if err == nil {
		t.Fatal("expected error for malformed template")
	}
}
