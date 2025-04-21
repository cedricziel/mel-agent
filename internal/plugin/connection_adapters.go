package plugin

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cedricziel/mel-agent/internal/db"
)

// connectionPlugin wraps a persisted connection instance as a ConnectionPlugin.
type connectionPlugin struct {
	id     string
	config map[string]interface{}
}

// Meta returns the plugin metadata for this connection.
func (c *connectionPlugin) Meta() PluginMeta {
	// No UI component; connections are managed via API/UI separately
	return PluginMeta{
		ID:          c.id,
		Version:     "0.1.0",
		Categories:  []string{"connection"},
		Params:      []ParamSpec{},
		UIComponent: "",
	}
}

// Connect returns the connection configuration as a resource handle.
func (c *connectionPlugin) Connect(ctx context.Context, cfg map[string]interface{}) (interface{}, error) {
	// Here, simply return the stored config. Real implementations could
	// establish DB pools, HTTP clients, etc.
	return c.config, nil
}

// RegisterConnectionPlugins loads all connections from the database and registers them as plugins.
func RegisterConnectionPlugins() {
   rows, err := db.DB.Query(`SELECT id, config FROM connections`)
   if err != nil {
       log.Printf("connection plugin: failed to load connections: %v", err)
       return
   }
   defer rows.Close()
   for rows.Next() {
       var id string
       var raw json.RawMessage
       if err := rows.Scan(&id, &raw); err != nil {
           log.Printf("connection plugin: scan error: %v", err)
           continue
       }
       cfg := map[string]interface{}{}
       if err := json.Unmarshal(raw, &cfg); err != nil {
           log.Printf("connection plugin: unmarshal config error for %s: %v", id, err)
           continue
       }
       plugin := &connectionPlugin{id: id, config: cfg}
       Register(plugin)
   }
}
