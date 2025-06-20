package transform

import (
	"github.com/aymerick/raymond"

	api "github.com/cedricziel/mel-agent/pkg/api"
)

// transformDefinition provides the built-in "Transform" node.
type transformDefinition struct{}

// Meta returns metadata for the Transform node.
func (transformDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "transform",
		Label:    "Transform",
		Icon:     "🔄",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("expression", "Expression", true).WithGroup("Settings").WithDescription("Transform input via expression"),
		},
	}
}

// ExecuteEnvelope applies the expression to the input envelope (currently passthrough).
func (d transformDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
-	val, ok := node.Data["expression"]
-	expr, ok := val.(string)
+	expr, ok := node.Data["expression"].(string)
	if !ok || expr == "" {
		err := api.NewNodeError(node.ID, node.Type, "expression parameter required")
		envelope.AddError(node.ID, "expression parameter required", err)
		return envelope, err
	}

	tmpl, err := raymond.Parse(expr)
	if err != nil {
		envelope.AddError(node.ID, "template parse failed", err)
		return envelope, err
	}

	data := map[string]interface{}{
		"input": envelope.Data,
		"vars":  ctx.Variables,
	}

	out, err := tmpl.Exec(data)
	if err != nil {
		envelope.AddError(node.ID, "template execute failed", err)
		return envelope, err
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = out
	result.DataType = "string"
	return result, nil
}

func (transformDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(transformDefinition{})
}

// assert that transformDefinition implements the interface
var _ api.NodeDefinition = (*transformDefinition)(nil)
