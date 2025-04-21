package api_test

import (
   "testing"
   // trigger plugins (webhook, schedule) and connection adapters register via internal/plugin
   _ "github.com/cedricziel/mel-agent/internal/plugin"
   // node adapters register via pkg/plugin/adapters
   _ "github.com/cedricziel/mel-agent/pkg/plugin/adapters"
   // builder node definitions register via internal/api/nodes
   _ "github.com/cedricziel/mel-agent/internal/api/nodes"

   "github.com/cedricziel/mel-agent/internal/plugin"
)

// TestPluginRegistry ensures core and builder node plugins, and trigger plugins, are registered.
func TestPluginRegistry(t *testing.T) {
   // Core node plugins expected
   coreIDs := []string{"timer", "schedule", "webhook", "slack", "http_request", "if", "switch", "agent", "llm", "inject"}
   for _, id := range coreIDs {
       if _, ok := plugin.GetNodePlugin(id); !ok {
           t.Errorf("expected core NodePlugin %q to be registered", id)
       }
   }
   // Trigger plugins expected
   triggerIDs := []string{"schedule", "webhook"}
   for _, id := range triggerIDs {
       if _, ok := plugin.GetTriggerPlugin(id); !ok {
           t.Errorf("expected TriggerPlugin %q to be registered", id)
       }
   }
}