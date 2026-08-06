package merge

import (
	"reflect"
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func newTestEnvelope(ctx api.ExecutionContext, node api.Node, input interface{}) *api.Envelope[interface{}] {
	trace := api.Trace{AgentID: ctx.AgentID, RunID: ctx.RunID, NodeID: node.ID, Step: node.ID, Attempt: 1}
	return core.NewEnvelope(input, trace)
}

func TestMergeDefinition_ConcatArrays(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "concat"}}

	input := []interface{}{[]interface{}{1, 2}, []interface{}{3}, 4}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output, got %T", out.Data)
	}
	expected := []interface{}{1, 2, 3, 4}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
	if out.DataType != "array" {
		t.Errorf("expected DataType array, got %s", out.DataType)
	}
}

func TestMergeDefinition_DefaultStrategy(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{}}

	input := []interface{}{[]interface{}{1}, []interface{}{2}}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	expected := []interface{}{1, 2}
	if !reflect.DeepEqual(out.Data, expected) {
		t.Errorf("expected %v got %v", expected, out.Data)
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
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output, got %T", out.Data)
	}
	expected := map[string]interface{}{"a": 1, "b": 3, "c": 4}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
	if out.DataType != "object" {
		t.Errorf("expected DataType object, got %s", out.DataType)
	}
}

func TestMergeDefinition_UnionArrays(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "union"}}

	input := []interface{}{[]interface{}{"a", "b"}, []interface{}{"b", "c"}}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output, got %T", out.Data)
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
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output, got %T", out.Data)
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
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice output, got %T", out.Data)
	}
	expected := []interface{}{"b"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_IntersectionArraysPreservesOrder(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "intersection"}}

	input := []interface{}{
		[]interface{}{"a", "b", "c", "d"},
		[]interface{}{"d", "c", "b", "a"},
	}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	expected := []interface{}{"a", "b", "c", "d"}
	if !reflect.DeepEqual(out.Data, expected) {
		t.Errorf("expected %v got %v", expected, out.Data)
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
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	got, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map output, got %T", out.Data)
	}
	expected := map[string]interface{}{"b": 5}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v got %v", expected, got)
	}
}

func TestMergeDefinition_NonArrayPassthrough(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "test-agent", RunID: "test-run"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "concat"}}

	input := map[string]interface{}{"a": 1}
	env := newTestEnvelope(ctx, node, input)

	out, err := def.ExecuteEnvelope(ctx, node, env)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}
	if !reflect.DeepEqual(out.Data, input) {
		t.Errorf("expected passthrough %v got %v", input, out.Data)
	}
}

func TestMergeDefinition_InvalidStrategy(t *testing.T) {
	def := mergeDefinition{}
	ctx := api.ExecutionContext{AgentID: "a", RunID: "r"}
	node := api.Node{ID: "merge", Type: "merge", Data: map[string]interface{}{"strategy": "bogus"}}

	input := []interface{}{1, 2}
	env := newTestEnvelope(ctx, node, input)

	_, err := def.ExecuteEnvelope(ctx, node, env)
	if err == nil {
		t.Fatal("expected error for invalid strategy")
	}
	if len(env.Errors) == 0 {
		t.Error("expected error recorded on envelope")
	}
}
