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
		Icon:     "🔄",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("expression", "Expression", true).
				WithGroup("Settings").
				WithFormat("template").
				WithDescription("Go template applied to the input, e.g. 'Hello, {{.input.name}}!'. The input data is available as .input and workflow variables as .vars"),
		},
	}
}

// ExecuteEnvelope renders the configured template expression against the input
// envelope. The input data is exposed as .input and execution variables as
// .vars. The rendered string becomes the output envelope's data.
func (d transformDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	expr, ok := node.Data["expression"].(string)
	if !ok || expr == "" {
		err := api.NewNodeError(node.ID, node.Type, "expression parameter required")
		envelope.AddError(node.ID, "expression parameter required", err)
		return envelope, err
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
