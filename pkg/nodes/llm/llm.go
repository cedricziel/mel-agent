package llm

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/sashabaranov/go-openai"
)

type llmDefinition struct{}

func (llmDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "llm",
		Label:    "LLM",
		Icon:     "🧠",
		Category: "AI",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("connectionId", "Model Connection", true).
				WithGroup("Settings").
				WithDescription("ID of the model provider connection"),
			api.NewStringParameter("model", "Model", false).
				WithDefault("gpt-3.5-turbo").
				WithGroup("Settings").
				WithDescription("Model name or ID"),
			api.NewStringParameter("systemPrompt", "System Prompt", false).
				WithGroup("Prompts").
				WithDescription("Optional system prompt for the model"),
		},
	}
}

// ExecuteEnvelope performs LLM completion using envelopes
func (d llmDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// Extract input from envelope
	input := envelope.Data

	// Perform existing logic
	// Resolve connection
	connID, ok := node.Data["connectionId"].(string)
	if !ok || connID == "" {
		return nil, errors.New("llm: missing connectionId parameter")
	}
	// Load connection secret and config
	var secretJSON, configJSON []byte
	err := db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, connID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("llm: connection %s not found", connID)
		}
		return nil, fmt.Errorf("llm: load connection error: %w", err)
	}
	// Parse secret (expecting {"apiKey": "..."})
	var sec struct {
		ApiKey string `json:"apiKey"`
	}
	if err := json.Unmarshal(secretJSON, &sec); err != nil {
		return nil, fmt.Errorf("llm: invalid connection secret: %w", err)
	}
	if sec.ApiKey == "" {
		return nil, errors.New("llm: apiKey missing in connection secret")
	}
	// Determine model
	model, _ := node.Data["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	// Build messages
	var msgs []openai.ChatCompletionMessage
	if sp, ok := node.Data["systemPrompt"].(string); ok && sp != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: sp})
	}
	switch v := input.(type) {
	case string:
		msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: v})
	case []interface{}:
		for _, mi := range v {
			if m, ok := mi.(map[string]interface{}); ok {
				role, _ := m["role"].(string)
				content, _ := m["content"].(string)
				msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: content})
			}
		}
	default:
		// marshal anything else
		raw, _ := json.Marshal(v)
		msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: string(raw)})
	}
	// Call OpenAI
	client := openai.NewClient(sec.ApiKey)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    model,
		Messages: msgs,
	})
	if err != nil {
		return nil, fmt.Errorf("llm: chat completion error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("llm: no response choices")
	}

	// Create result envelope
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = resp.Choices[0].Message.Content
	return result, nil
}

func (llmDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(llmDefinition{})
}

// assert that llmDefinition implements both interfaces
var _ api.NodeDefinition = (*llmDefinition)(nil)
