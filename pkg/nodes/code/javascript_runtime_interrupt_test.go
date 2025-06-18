package code

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptRuntime_Execute_ContextCancellation(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"value": 42},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	t.Run("should interrupt VM execution on context cancellation", func(t *testing.T) {
		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Code that would run for a long time without interruption
		longRunningCode := `
			let count = 0;
			while (true) {
				count++;
				// This loop would run forever without VM interruption
				if (count > 1000000) {
					break; // Safety net in case interruption doesn't work
				}
			}
			return count;
		`

		// Start execution
		resultChan := make(chan struct {
			result interface{}
			err    error
		}, 1)

		go func() {
			result, err := runtime.Execute(ctx, longRunningCode, execContext)
			resultChan <- struct {
				result interface{}
				err    error
			}{result, err}
		}()

		// Cancel the context after a short delay
		time.Sleep(50 * time.Millisecond)
		cancel()

		// Wait for the result or timeout
		select {
		case res := <-resultChan:
			// Should get an error due to interruption
			assert.Error(t, res.err)
			assert.Contains(t, res.err.Error(), "execution cancelled")
		case <-time.After(2 * time.Second):
			t.Fatal("Execution should have been cancelled quickly but timed out")
		}
	})

	t.Run("should handle context cancellation before execution starts", func(t *testing.T) {
		// Create an already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		code := `return 42;`

		result, err := runtime.Execute(ctx, code, execContext)

		// Should get an error due to context cancellation
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("should complete normally when context is not cancelled", func(t *testing.T) {
		ctx := context.Background()

		code := `return input.data.value * 2;`

		result, err := runtime.Execute(ctx, code, execContext)

		require.NoError(t, err)
		assert.Equal(t, int64(84), result) // 42 * 2
	})

	t.Run("should handle timeout context", func(t *testing.T) {
		// Create a context with a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Code that takes longer than the timeout
		slowCode := `
			// Simulate some work
			let sum = 0;
			for (let i = 0; i < 1000000; i++) {
				sum += i;
			}
			return sum;
		`

		result, err := runtime.Execute(ctx, slowCode, execContext)

		// Should get an error due to timeout
		assert.Error(t, err)
		assert.Nil(t, result)
		// Error should indicate cancellation or timeout
		assert.True(t,
			err.Error() == "JavaScript execution cancelled: context deadline exceeded" ||
				err.Error() == "execution error: Error: execution cancelled",
			"Expected cancellation error, got: %v", err)
	})
}

func TestJavaScriptRuntime_Execute_NonBlockingSends(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	t.Run("should not block when context is cancelled during panic recovery", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Code that will cause a panic
		panicCode := `throw new Error("intentional panic");`

		// Cancel context immediately to test non-blocking sends
		cancel()

		// This should not hang due to non-blocking channel sends
		done := make(chan bool, 1)
		go func() {
			runtime.Execute(ctx, panicCode, execContext)
			done <- true
		}()

		select {
		case <-done:
			// Success - execution completed without hanging
		case <-time.After(1 * time.Second):
			t.Fatal("Execution hung due to blocking channel sends")
		}
	})
}
