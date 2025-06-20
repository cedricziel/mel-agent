package merge

import (
	"reflect"
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func TestMergeDefinition_ConcatArrays(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "concat"}}

	input := []interface{}{[]interface{}{1, 2}, []interface{}{3}, 4}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output")
	}
	expected := []interface{}{1, 2, 3, 4}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_UnionMaps(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "union"}}

	input := []interface{}{
		map[string]interface{}{"a": 1, "b": 2},
		map[string]interface{}{"b": 3, "c": 4},
	}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output")
	}
	expected := map[string]interface{}{"a": 1, "b": 3, "c": 4}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_UnionArrays(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "union"}}

	input := []interface{}{[]interface{}{"a", "b"}, []interface{}{"b", "c"}}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output")
	}
	expected := []interface{}{"a", "b", "c"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_DeepMerge(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "deep"}}

	input := []interface{}{
		map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		map[string]interface{}{"b": map[string]interface{}{"d": 3}},
	}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output")
	}
	expected := map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2, "d": 3}}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_IntersectionArrays(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "intersection"}}

	input := []interface{}{[]interface{}{"a", "b", "c"}, []interface{}{"b", "c"}, []interface{}{"b", "d"}}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output")
	}
	expected := []interface{}{"b"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_IntersectionMaps(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "intersection"}}

	input := []interface{}{
		map[string]interface{}{"a": 1, "b": 2},
		map[string]interface{}{"b": 3, "c": 4},
		map[string]interface{}{"b": 5},
	}
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	env := core.NewEnvelope(interface{}(input), trace)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output")
	}
	expected := map[string]interface{}{"b": 5}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}
