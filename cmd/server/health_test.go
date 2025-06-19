package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	body := rr.Body.String()
	assert.Contains(t, body, `"status":"ok"`)
	assert.Contains(t, body, `"timestamp"`)

	t.Logf("✅ Health check response: %s", body)
}

func TestReadinessCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ready", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(readinessCheckHandler)

	handler.ServeHTTP(rr, req)

	// The response code depends on whether the database is available
	// In tests without a real database, it should return 503
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	body := rr.Body.String()
	assert.Contains(t, body, `"status"`)
	assert.Contains(t, body, `"timestamp"`)
	assert.Contains(t, body, `"checks"`)
	assert.Contains(t, body, `"database"`)

	// Should contain either "ready" or "not_ready"
	assert.True(t, strings.Contains(body, `"ready"`) || strings.Contains(body, `"not_ready"`))

	t.Logf("✅ Readiness check response: %s", body)
}

func TestHealthCheckEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		handler        http.HandlerFunc
		expectedStatus int
		mustContain    []string
	}{
		{
			name:           "health endpoint",
			path:           "/health",
			handler:        healthCheckHandler,
			expectedStatus: http.StatusOK,
			mustContain:    []string{`"status":"ok"`, `"timestamp"`},
		},
		{
			name:    "readiness endpoint",
			path:    "/ready",
			handler: readinessCheckHandler,
			// Status depends on DB availability, so we check for either
			expectedStatus: -1, // Special case - check multiple statuses
			mustContain:    []string{`"status"`, `"checks"`, `"database"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			tt.handler.ServeHTTP(rr, req)

			if tt.expectedStatus != -1 {
				assert.Equal(t, tt.expectedStatus, rr.Code)
			} else {
				// For readiness check, accept either OK or Service Unavailable
				assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, rr.Code)
			}

			body := rr.Body.String()
			for _, mustContain := range tt.mustContain {
				assert.Contains(t, body, mustContain)
			}

			t.Logf("✅ %s test passed: %s", tt.name, body)
		})
	}
}

func TestHealthCheckResponseFormat(t *testing.T) {
	t.Run("health check JSON format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		healthCheckHandler(rr, req)

		body := rr.Body.String()

		// Verify it's valid JSON structure
		assert.True(t, strings.HasPrefix(body, `{"`))
		assert.True(t, strings.HasSuffix(body, `"}`))
		assert.Contains(t, body, `"status":"ok"`)

		// Timestamp should be in ISO format
		assert.Contains(t, body, `"timestamp":"`)
		assert.Contains(t, body, `T`) // ISO 8601 format contains 'T'
		assert.Contains(t, body, `Z`) // UTC timezone indicator
	})

	t.Run("readiness check JSON format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/ready", nil)
		rr := httptest.NewRecorder()

		readinessCheckHandler(rr, req)

		body := rr.Body.String()

		// Verify it's valid JSON structure
		assert.True(t, strings.HasPrefix(body, `{"`))
		assert.True(t, strings.HasSuffix(body, `}`))

		// Check required fields
		assert.Contains(t, body, `"status"`)
		assert.Contains(t, body, `"timestamp"`)
		assert.Contains(t, body, `"checks"`)
		assert.Contains(t, body, `"database"`)
	})
}
