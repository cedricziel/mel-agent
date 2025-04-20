package api

import (
   "context"
   "github.com/cedricziel/mel-agent/internal/plugin"
)

// nodeDefinitionAdapter wraps NodeDefinition into a plugin.NodePlugin.
type nodeDefinitionAdapter struct {
   def NodeDefinition
}

// Meta converts the NodeDefinition metadata into PluginMeta.
func (a nodeDefinitionAdapter) Meta() plugin.PluginMeta {
   nd := a.def.Meta()
   // Convert ParameterDefinition to plugin.ParamSpec
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
       ID:          nd.Type,
       Version:     "0.1.0",
       Categories:  []string{"node"},
       Params:      params,
       UIComponent: "",
   }
}

// Execute adapts plugin inputs and invokes the underlying NodeDefinition.
func (a nodeDefinitionAdapter) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
   agentID, _ := inputs["agent_id"].(string)
   input := inputs["input"]
   config, _ := inputs["config"].(map[string]interface{})
   // Build api.Node
   node := Node{ID: "", Type: a.def.Meta().Type, Data: config}
   // Execute
   out, err := a.def.Execute(agentID, node, input)
   if err != nil {
       return nil, err
   }
   // Normalize output
   if m, ok := out.(map[string]interface{}); ok {
       return m, nil
   }
   return map[string]interface{}{"output": out}, nil
}

// convertValidators transforms ValidatorSpec to plugin.ValidatorSpec.
func convertValidators(in []ValidatorSpec) []plugin.ValidatorSpec {
   var out []plugin.ValidatorSpec
   for _, v := range in {
       out = append(out, plugin.ValidatorSpec{Type: v.Type, Params: v.Params})
   }
   return out
}

// Register all existing node definitions as plugins.
func init() {
   for _, def := range ListNodeDefinitions() {
       plugin.Register(nodeDefinitionAdapter{def: def})
   }
}