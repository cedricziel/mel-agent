package plugin

import (
   "context"

   "github.com/cedricziel/mel-agent/internal/api"
)

// adapter wraps an api.NodeDefinition into a plugin.NodePlugin
type nodeDefinitionAdapter struct {
   def api.NodeDefinition
}

// Meta returns PluginMeta derived from the NodeDefinition metadata
func (a nodeDefinitionAdapter) Meta() PluginMeta {
   nd := a.def.Meta()
   return PluginMeta{
       ID:          nd.Type,
       Version:     "0.1.0",
       Categories:  []string{"node"},
       Params:      nd.Parameters,
       UIComponent: "",
   }
}

// Execute adapts inputs to the underlying NodeDefinition.Execute
func (a nodeDefinitionAdapter) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
   // Extract payload and config
   payload := inputs["input"]
   config, _ := inputs["config"].(map[string]interface{})
   // Construct api.Node
   node := api.Node{Data: config}
   // Extract optional agentID
   agentID, _ := inputs["agent_id"].(string)
   // Call underlying definition
   result, err := a.def.Execute(agentID, node, payload)
   if err != nil {
       return nil, err
   }
   // Normalize result to map[string]interface{}
   if m, ok := result.(map[string]interface{}); ok {
       return m, nil
   }
   return map[string]interface{}{"output": result}, nil
}

// init registers all existing NodeDefinitions as NodePlugins
func init() {
   for _, def := range api.ListNodeDefinitions() {
       Register(nodeDefinitionAdapter{def: def})
   }
}