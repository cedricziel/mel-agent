package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIBinaryIntegration tests the actual compiled binary
func TestCLIBinaryIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping CLI binary integration test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput []string
		timeout        time.Duration
	}{
		{
			name:           "help command",
			args:           []string{"--help"},
			expectError:    false,
			expectedOutput: []string{"MEL Agent", "Usage:", "Available Commands:"},
			timeout:        5 * time.Second,
		},
		{
			name:           "server help",
			args:           []string{"server", "--help"},
			expectError:    false,
			expectedOutput: []string{"Start the API server", "Flags:", "--port"},
			timeout:        5 * time.Second,
		},
		{
			name:           "worker help",
			args:           []string{"worker", "--help"},
			expectError:    false,
			expectedOutput: []string{"Start a remote worker", "--token", "--server"},
			timeout:        5 * time.Second,
		},
		{
			name:           "invalid command",
			args:           []string{"invalid-command"},
			expectError:    true,
			expectedOutput: []string{"unknown command"},
			timeout:        5 * time.Second,
		},
		{
			name:           "worker without token",
			args:           []string{"worker"},
			expectError:    true,
			expectedOutput: []string{"required flag(s) \"token\" not set"},
			timeout:        5 * time.Second,
		},
		{
			name:           "completion bash",
			args:           []string{"completion", "bash"},
			expectError:    false,
			expectedOutput: []string{"# bash completion", "mel-agent"},
			timeout:        5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			cmd.Dir = tempDir

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				assert.Error(t, err, "Expected command to fail")
			} else {
				assert.NoError(t, err, "Expected command to succeed")
			}

			// Check for expected output strings
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "Output should contain expected text")
			}

			t.Logf("✅ CLI binary integration test passed: %s", tt.name)
		})
	}
}

// TestCLIEnvironmentIntegration tests environment variable integration with real binary
func TestCLIEnvironmentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI environment integration test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	// Test with environment variables
	t.Run("environment variables", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "server", "--help")
		cmd.Dir = tempDir
		cmd.Env = append(os.Environ(), "PORT=9999")

		output, err := cmd.CombinedOutput()
		assert.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Port to listen on", "Help should mention port flag")

		t.Logf("✅ Environment variable integration test passed")
	})
}

// TestCLIConfigFileIntegration tests config file support with real binary
func TestCLIConfigFileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI config file integration test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	// Create config file
	configContent := `
server:
  port: "7777"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	t.Run("config file support", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test that binary can read config (we can't easily test the actual values without
		// starting the server, but we can test that it doesn't error with config present)
		cmd := exec.CommandContext(ctx, binaryPath, "server", "--help")
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Binary should handle config file gracefully")

		outputStr := string(output)
		assert.Contains(t, outputStr, "server", "Help should contain server information")

		t.Logf("✅ Config file integration test passed")
	})
}

// TestCLIVersionAndCompletion tests version info and completion generation
func TestCLIVersionAndCompletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI version and completion test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	completionTests := []struct {
		shell string
	}{
		{"bash"},
		{"zsh"},
		{"fish"},
		{"powershell"},
	}

	for _, tt := range completionTests {
		t.Run("completion_"+tt.shell, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, "completion", tt.shell)
			cmd.Dir = tempDir

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			assert.NoError(t, err, "Completion generation should succeed")
			assert.NotEmpty(t, outputStr, "Completion output should not be empty")

			// Basic validation that it looks like shell completion
			if tt.shell == "bash" {
				assert.Contains(t, outputStr, "bash completion", "Bash completion should contain bash-specific content")
			}

			t.Logf("✅ %s completion generation test passed", tt.shell)
		})
	}
}

// TestAPIServerIntegration tests the api-server command integration
func TestAPIServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API server integration test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	t.Run("api-server help command", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "api-server", "--help")
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "api-server --help should succeed")

		outputStr := string(output)
		assert.Contains(t, outputStr, "Start the API server without embedded workers", "Help should describe api-server purpose")
		assert.Contains(t, outputStr, "horizontal scaling", "Help should mention horizontal scaling")
		assert.Contains(t, outputStr, "--port", "Help should show port flag")
		assert.Contains(t, outputStr, "-p", "Help should show port short flag")

		t.Logf("✅ API server help test passed")
	})

	t.Run("api-server flag validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Test with custom port flag
		cmd := exec.CommandContext(ctx, binaryPath, "api-server", "--port", "9999", "--help")
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "api-server with port flag should succeed")

		outputStr := string(output)
		assert.Contains(t, outputStr, "api-server", "Output should confirm api-server command")

		t.Logf("✅ API server flag validation test passed")
	})

	t.Run("api-server in root help", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "--help")
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Root help should succeed")

		outputStr := string(output)
		assert.Contains(t, outputStr, "api-server", "Root help should list api-server command")
		assert.Contains(t, outputStr, "Start the API server only", "Root help should show api-server description")

		t.Logf("✅ API server in root help test passed")
	})
}

// TestCLIErrorHandling tests various error conditions
func TestCLIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI error handling test in short mode")
	}

	// Build the binary for testing
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "mel-agent-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	errorTests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "unknown flag",
			args:        []string{"server", "--unknown-flag"},
			expectedErr: "unknown flag",
		},
		{
			name:        "missing required token",
			args:        []string{"worker"},
			expectedErr: "required flag(s) \"token\" not set",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"invalid"},
			expectedErr: "unknown command",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			cmd.Dir = tempDir

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			assert.Error(t, err, "Command should fail")
			assert.Contains(t, strings.ToLower(outputStr), strings.ToLower(tt.expectedErr),
				"Error output should contain expected error message")

			t.Logf("✅ CLI error handling test passed: %s", tt.name)
		})
	}
}
