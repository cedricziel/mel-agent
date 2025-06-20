package api

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// AssistantChat provides AI-powered assistance for workflow building
func (h *OpenAPIHandlers) AssistantChat(ctx context.Context, request AssistantChatRequestObject) (AssistantChatResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return AssistantChat500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return AssistantChat400JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		errorMsg := "configuration error"
		message := "OPENAI_API_KEY not set"
		return AssistantChat500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert request messages to OpenAI format
	var messages []openai.ChatCompletionMessage
	for _, msg := range request.Body.Messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
		if msg.Name != nil {
			openaiMsg.Name = *msg.Name
		}
		if msg.FunctionCall != nil {
			openaiMsg.FunctionCall = &openai.FunctionCall{
				Name:      *msg.FunctionCall.Name,
				Arguments: *msg.FunctionCall.Arguments,
			}
		}
		messages = append(messages, openaiMsg)
	}

	// Define available functions for assistant to call
	funcs := []openai.FunctionDefinition{
		{
			Name:        "list_node_types",
			Description: "List available node types and their metadata",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_workflow",
			Description: "Get the current workflow graph (nodes and edges)",
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
		{
			Name:        "get_node_definition",
			Description: "Retrieve metadata for a specific node type, including default parameter values",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"type": map[string]interface{}{"type": "string", "description": "Node type name"}},
				"required":   []string{"type"},
			},
		},
	}

	// Create OpenAI client and make request
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:        openai.GPT4oMini,
		Messages:     messages,
		Functions:    funcs,
		FunctionCall: "auto",
	})
	if err != nil {
		errorMsg := "openai error"
		message := err.Error()
		return AssistantChat500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert OpenAI response to our format
	var choices []ChatChoice
	for _, choice := range resp.Choices {
		apiChoice := ChatChoice{
			Index: &choice.Index,
		}

		// Convert message
		msg := ChatMessage{
			Role:    ChatMessageRole(choice.Message.Role),
			Content: choice.Message.Content,
		}
		if choice.Message.Name != "" {
			msg.Name = &choice.Message.Name
		}
		if choice.Message.FunctionCall != nil {
			msg.FunctionCall = &FunctionCall{
				Name:      &choice.Message.FunctionCall.Name,
				Arguments: &choice.Message.FunctionCall.Arguments,
			}
		}
		apiChoice.Message = &msg

		if choice.FinishReason != "" {
			reason := ChatChoiceFinishReason(choice.FinishReason)
			apiChoice.FinishReason = &reason
		}

		choices = append(choices, apiChoice)
	}

	// Build response
	response := AssistantChatResponse{
		Id:      &resp.ID,
		Object:  &resp.Object,
		Created: func() *int { i := int(resp.Created); return &i }(),
		Model:   &resp.Model,
		Choices: &choices,
	}

	if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 || resp.Usage.TotalTokens > 0 {
		response.Usage = &ChatUsage{
			PromptTokens:     &resp.Usage.PromptTokens,
			CompletionTokens: &resp.Usage.CompletionTokens,
			TotalTokens:      &resp.Usage.TotalTokens,
		}
	}

	return AssistantChat200JSONResponse(response), nil
}
