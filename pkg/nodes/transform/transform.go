package transform

import (
	"bytes"
	"text/template"

	api "github.com/cedricziel/mel-agent/pkg/api"
)

// transformDefinition provides the built-in "Transform" node.
type transformDefinition struct{}

// Meta returns metadata for the Transform node.
func (transformDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "transform",
		Label:    "Transform",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("expression", "Expression", true).WithGroup("Settings").WithDescription("Transform input via expression"),
		},
	}
}

// ExecuteEnvelope applies the expression to the input envelope (currently passthrough).
func (d transformDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	expr, _ := node.Data["expression"].(string)
	if expr == "" {
		return envelope, api.NewNodeError(node.ID, node.Type, "expression parameter required")
	}

	tmpl, err := template.New("transform").Parse(expr)
	if err != nil {
		envelope.AddError(node.ID, "template parse failed", err)
		return envelope, err
	}

	data := map[string]interface{}{
		"input": envelope.Data,
		"vars":  ctx.Variables,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		envelope.AddError(node.ID, "template execute failed", err)
		return envelope, err
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = buf.String()
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
