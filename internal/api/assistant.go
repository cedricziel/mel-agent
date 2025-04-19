package api

import (
	"encoding/json"
	"net/http"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// assistantChatHandler proxies chat messages to OpenAI, supporting function calls for builder tools.
func assistantChatHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages []openai.ChatCompletionMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "OPENAI_API_KEY not set"})
		return
	}
	client := openai.NewClient(apiKey)
	// Define available functions for assistant to call
	// Define available functions for assistant to call
	// Define available functions for assistant to call
	funcs := []openai.FunctionDefinition{
       {
           Name:        "list_node_types",
           Description: "List available node types and their metadata",
           // Empty object schema: must include properties field even if empty
           Parameters: map[string]interface{}{
               "type":       "object",
               "properties": map[string]interface{}{},
           },
       },
		{
			Name:        "get_node_type_schema",
			Description: "Get the JSON Schema for a given node type",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{"type": "string", "description": "Node type name"},
				},
				"required": []string{"type"},
			},
		},
		{
			Name:        "add_node",
			Description: "Add a node to the workflow",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type":       map[string]interface{}{"type": "string", "description": "Node type to add"},
					"label":      map[string]interface{}{"type": "string", "description": "Label for the node"},
					"parameters": map[string]interface{}{"type": "object", "description": "Node parameters"},
					"position": map[string]interface{}{
						"type":        "object",
						"description": "Position for the new node",
						"properties": map[string]interface{}{
							"x": map[string]interface{}{"type": "number", "description": "X coordinate"},
							"y": map[string]interface{}{"type": "number", "description": "Y coordinate"},
						},
						"required": []string{"x", "y"},
					},
				},
				"required": []string{"type"},
			},
		},
		{
			Name:        "connect_nodes",
			Description: "Connect two nodes in the workflow",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source_id": map[string]interface{}{"type": "string", "description": "ID of source node"},
					"target_id": map[string]interface{}{"type": "string", "description": "ID of target node"},
				},
				"required": []string{"source_id", "target_id"},
			},
		},
	}
	// Call OpenAI ChatCompletion
	resp, err := client.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model:        openai.GPT4oMini,
		Messages:     req.Messages,
		Functions:    funcs,
		FunctionCall: "auto",
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
