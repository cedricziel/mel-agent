package transform

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func TestTransformDefinition_Template(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{"expression": "Hello, {{ .input.name }}!"}}

	input := map[string]interface{}{"name": "Alice"}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	if out.Data != "Hello, Alice!" {
		t.Errorf("unexpected result: %v", out.Data)
	}
}

func TestTransformDefinition_Variables(t *testing.T) {
	def := transformDefinition{}
	ctx := api.ExecutionContext{AgentID: "agent", RunID: "run", Variables: map[string]interface{}{"role": "admin"}}
	node := api.Node{ID: "transform", Type: "transform", Data: map[string]interface{}{"expression": "{{ .vars.role }}-{{ .input }}"}}

	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}("data"), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	if out.Data != "admin-data" {
		t.Errorf("unexpected result: %v", out.Data)
	}
}
