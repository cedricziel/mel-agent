package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIAssistantChatNoAPIKey tests assistant chat without OpenAI API key
func TestOpenAPIAssistantChatNoAPIKey(t *testing.T) {
	// Ensure no API key is set
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test agent first
	agentID := getTestAgentID(t, db)

	chatRequest := AssistantChatRequest{
		Messages: []ChatMessage{
			{
				Role:    User,
				Content: "Hello, can you help me build a workflow?",
			},
		},
	}
	reqBody, _ := json.Marshal(chatRequest)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/"+agentID.String()+"/assistant/chat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "OPENAI_API_KEY")
}

// TestOpenAPIAssistantChatAgentNotFound tests assistant chat with non-existent agent
func TestOpenAPIAssistantChatAgentNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	chatRequest := AssistantChatRequest{
		Messages: []ChatMessage{
			{
				Role:    User,
				Content: "Hello, can you help me build a workflow?",
			},
		},
	}
	reqBody, _ := json.Marshal(chatRequest)

	// Use a non-existent agent ID
	req := httptest.NewRequest(http.MethodPost, "/api/agents/00000000-0000-0000-0000-000000000999/assistant/chat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "Agent not found")
}

// TestOpenAPIAssistantChatInvalidJSON tests assistant chat with malformed JSON
func TestOpenAPIAssistantChatInvalidJSON(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test agent first
	agentID := getTestAgentID(t, db)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/"+agentID.String()+"/assistant/chat", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
