package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
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

	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
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

	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "OPENAI_API_KEY")
}

// TestOpenAPIAssistantChatEmptyMessages tests assistant chat with empty messages
func TestOpenAPIAssistantChatEmptyMessages(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	// Create handlers with test API key
	handlers := NewOpenAPIHandlers(db, mockEngine, "test-api-key")
	strictHandler := NewStrictHandler(handlers, nil)
	
	// Create router manually for test
	r := chi.NewRouter()
	HandlerFromMux(strictHandler, r)
	router := r

	chatRequest := AssistantChatRequest{
		Messages: []ChatMessage{}, // Empty messages
	}
	reqBody, _ := json.Marshal(chatRequest)

	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "openai")
}

// TestOpenAPIAssistantChatInvalidJSON tests assistant chat with malformed JSON
func TestOpenAPIAssistantChatInvalidJSON(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
