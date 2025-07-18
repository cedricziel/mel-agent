package adapters

import (
	"context"
	"time"

	api "github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/plugin"
)

// NodeDefinitionAdapter wraps a NodeDefinition into a plugin.NodePlugin.
type NodeDefinitionAdapter struct {
	Def api.NodeDefinition
}

// convertValidators transforms internalapi.ValidatorSpec into plugin.ValidatorSpec.
func convertValidators(in []api.ValidatorSpec) []plugin.ValidatorSpec {
	var out []plugin.ValidatorSpec
	for _, v := range in {
		out = append(out, plugin.ValidatorSpec{Type: v.Type, Params: v.Params})
	}
	return out
}

// Meta returns the PluginMeta for the underlying NodeDefinition.
func (a NodeDefinitionAdapter) Meta() plugin.PluginMeta {
	nd := a.Def.Meta()
	// Convert api.ParameterDefinition to plugin.ParamSpec
	var params []plugin.ParamSpec
	for _, p := range nd.Parameters {
		params = append(params, plugin.ParamSpec{
			Name:                p.Name,
			Label:               p.Label,
			Type:                p.Type,
			Required:            p.Required,
			Default:             p.Default,
			Group:               p.Group,
			VisibilityCondition: p.VisibilityCondition,
			Options:             p.Options,
			Validators:          convertValidators(p.Validators),
			Description:         p.Description,
		})
	}
	return plugin.PluginMeta{
		ID:         nd.Type,
		Version:    "0.1.0",
		Categories: []string{"node"},
		Params:     params,
	}
}

// Execute adapts plugin inputs and invokes the underlying NodeDefinition.
func (a NodeDefinitionAdapter) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	agentID, _ := inputs["agent_id"].(string)
	input := inputs["input"]
	config, _ := inputs["config"].(map[string]interface{})
	node := api.Node{ID: "", Type: a.Def.Meta().Type, Data: config}
	execCtx := api.ExecutionContext{
		AgentID: agentID,
		Mel:     api.NewMel(), // Provide platform utilities
	}

	// Create envelope for the new interface
	envelope := &api.Envelope[interface{}]{
		ID:       "plugin-" + node.ID,
		IssuedAt: time.Now(),
		Version:  1,
		DataType: "unknown",
		Data:     input,
		Trace: api.Trace{
			AgentID: execCtx.AgentID,
			RunID:   execCtx.RunID,
			NodeID:  node.ID,
			Step:    node.ID,
			Attempt: 1,
		},
		Variables: execCtx.Variables,
	}

	result, err := a.Def.ExecuteEnvelope(execCtx, node, envelope)
	if err != nil {
		return nil, err
	}

	out := result.Data
	if m, ok := out.(map[string]interface{}); ok {
		return m, nil
	}
	return map[string]interface{}{"output": out}, nil
}

// ProvideNodePlugins returns a slice of plugin.NodePlugin wrapping
// both core and builder NodeDefinitions.
func ProvideNodePlugins() []plugin.NodePlugin {
	core := api.AllCoreDefinitions()
	builder := api.ListNodeDefinitions()
	var out []plugin.NodePlugin
	for _, def := range core {
		out = append(out, NodeDefinitionAdapter{Def: def})
	}
	for _, def := range builder {
		out = append(out, NodeDefinitionAdapter{Def: def})
	}
	return out
}

func init() {
	for _, p := range ProvideNodePlugins() {
		plugin.Register(p)
	}
}
