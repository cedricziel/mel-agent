package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIServerCommand(t *testing.T) {
	// Create a new root command for testing
	cmd := &cobra.Command{Use: "test"}

	// Add our commands
	cmd.AddCommand(serverCmd)
	cmd.AddCommand(apiServerCmd)
	cmd.AddCommand(workerCmd)

	t.Run("api-server command exists", func(t *testing.T) {
		apiCmd, _, err := cmd.Find([]string{"api-server"})
		require.NoError(t, err)
		assert.NotNil(t, apiCmd)
		assert.Equal(t, "api-server", apiCmd.Use)
	})

	t.Run("api-server command has correct short description", func(t *testing.T) {
		apiCmd, _, _ := cmd.Find([]string{"api-server"})
		assert.Equal(t, "Start the API server only", apiCmd.Short)
	})

	t.Run("api-server command has port flag", func(t *testing.T) {
		apiCmd, _, _ := cmd.Find([]string{"api-server"})
		flag := apiCmd.Flag("port")
		require.NotNil(t, flag)
		assert.Equal(t, "p", flag.Shorthand)
		assert.Equal(t, "8080", flag.DefValue)
	})

	t.Run("api-server help mentions no embedded workers", func(t *testing.T) {
		apiCmd, _, _ := cmd.Find([]string{"api-server"})
		assert.Contains(t, apiCmd.Long, "without embedded workers")
		assert.Contains(t, apiCmd.Long, "horizontal scaling")
	})

	t.Run("api-server differs from server command", func(t *testing.T) {
		serverCmd, _, _ := cmd.Find([]string{"server"})
		apiCmd, _, _ := cmd.Find([]string{"api-server"})

		// Server command mentions embedded workers
		assert.Contains(t, serverCmd.Long, "embedded workers")

		// API server command explicitly says no embedded workers
		assert.Contains(t, apiCmd.Long, "without embedded workers")

		// They should be different commands
		assert.NotEqual(t, serverCmd.Use, apiCmd.Use)
		assert.NotEqual(t, serverCmd.Short, apiCmd.Short)
	})
}

func TestAPIServerCommandOutput(t *testing.T) {
	t.Run("help output shows all commands", func(t *testing.T) {
		// Create root command
		cmd := &cobra.Command{
			Use:   "mel-agent",
			Short: "MEL Agent - AI Agents SaaS platform",
		}
		cmd.AddCommand(serverCmd)
		cmd.AddCommand(apiServerCmd)
		cmd.AddCommand(workerCmd)

		// Capture help output
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "server")
		assert.Contains(t, output, "api-server")
		assert.Contains(t, output, "worker")
		assert.Contains(t, output, "Start the API server only")
	})

	t.Run("api-server help output", func(t *testing.T) {
		// Create root command
		cmd := &cobra.Command{
			Use:   "mel-agent",
			Short: "MEL Agent - AI Agents SaaS platform",
		}
		cmd.AddCommand(apiServerCmd)

		// Capture help output
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"api-server", "--help"})

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Start the API server without embedded workers")
		assert.Contains(t, output, "horizontal scaling")
		assert.Contains(t, output, "--port")
		assert.Contains(t, output, "-p")
	})
}

func TestCommandLineArgumentParsing(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCmd  string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "api-server with default port",
			args:         []string{"api-server"},
			expectedCmd:  "api-server",
			expectedPort: "8080",
		},
		{
			name:         "api-server with custom port",
			args:         []string{"api-server", "--port", "9090"},
			expectedCmd:  "api-server",
			expectedPort: "9090",
		},
		{
			name:         "api-server with short port flag",
			args:         []string{"api-server", "-p", "3000"},
			expectedCmd:  "api-server",
			expectedPort: "3000",
		},
		{
			name:         "server command still works",
			args:         []string{"server"},
			expectedCmd:  "server",
			expectedPort: "8080",
		},
		{
			name:        "worker command still works",
			args:        []string{"worker"},
			expectedCmd: "worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for each test
			cmd := &cobra.Command{Use: "mel-agent"}
			cmd.AddCommand(serverCmd)
			cmd.AddCommand(apiServerCmd)
			cmd.AddCommand(workerCmd)

			// Find the command
			foundCmd, _, err := cmd.Find(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCmd, foundCmd.Use)

			// Check port flag if applicable
			if tt.expectedPort != "" && (strings.Contains(tt.expectedCmd, "server")) {
				// Parse the flags
				err := foundCmd.ParseFlags(tt.args[1:])
				require.NoError(t, err)

				portFlag := foundCmd.Flag("port")
				if portFlag != nil && portFlag.Changed {
					assert.Equal(t, tt.expectedPort, portFlag.Value.String())
				}
			}
		})
	}
}
