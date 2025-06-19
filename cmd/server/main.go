package main

// Standard library + third‑party imports
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	httpApi "github.com/cedricziel/mel-agent/internal/api"
	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/injector"
	"github.com/cedricziel/mel-agent/internal/triggers"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/cedricziel/mel-agent/pkg/plugin"
)

func main() {
	initConfig()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mel-agent",
	Short: "MEL Agent - AI Agents SaaS platform",
	Long: `MEL Agent is a platform for building and running AI agent workflows.

It provides a visual workflow builder with support for various node types,
triggers, and integrations. You can run it as an API server or as a distributed
worker for horizontal scaling.`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the API server",
	Long: `Start the API server with embedded workers.

The server will:
- Connect to PostgreSQL database and run migrations
- Load and register node plugins
- Start embedded workflow workers
- Start trigger scheduler
- Serve API endpoints at /api/*
- Handle webhooks at /webhooks/{provider}/{triggerID}
- Provide health check at /health`,
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetString("server.port")
		startServer(port)
	},
}

var apiServerCmd = &cobra.Command{
	Use:   "api-server",
	Short: "Start the API server only",
	Long: `Start the API server without embedded workers.

The api-server will:
- Connect to PostgreSQL database and run migrations
- Load and register node plugins
- Start trigger scheduler
- Serve API endpoints at /api/*
- Handle webhooks at /webhooks/{provider}/{triggerID}
- Provide health check at /health

This mode is designed for horizontal scaling of API servers
separate from worker processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetString("server.port")
		startAPIServer(port)
	},
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start a workflow worker",
	Long: `Start a remote worker process that connects to an API server.

The worker will:
- Connect to the specified API server
- Authenticate using the provided token
- Process workflow tasks with specified concurrency
- Auto-generate worker ID if not provided`,
	Run: func(cmd *cobra.Command, args []string) {
		serverURL := viper.GetString("worker.server")
		token := viper.GetString("worker.token")
		workerID := viper.GetString("worker.id")
		concurrency := viper.GetInt("worker.concurrency")

		if token == "" {
			log.Fatal("Worker token is required. Set MEL_WORKER_TOKEN environment variable or use --token flag")
		}

		startWorker(serverURL, token, workerID, concurrency)
	},
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(workerCmd)

	// Server command flags
	serverCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))

	// API Server command flags (same as server)
	apiServerCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	viper.BindPFlag("server.port", apiServerCmd.Flags().Lookup("port"))

	// Worker command flags
	workerCmd.Flags().StringP("server", "s", "http://localhost:8080", "API server URL")
	workerCmd.Flags().StringP("token", "t", "", "Authentication token (required)")
	workerCmd.Flags().String("id", "", "Worker ID (auto-generated if empty)")
	workerCmd.Flags().IntP("concurrency", "c", 5, "Number of concurrent workflow executions")

	// Bind worker flags to viper
	viper.BindPFlag("worker.server", workerCmd.Flags().Lookup("server"))
	viper.BindPFlag("worker.token", workerCmd.Flags().Lookup("token"))
	viper.BindPFlag("worker.id", workerCmd.Flags().Lookup("id"))
	viper.BindPFlag("worker.concurrency", workerCmd.Flags().Lookup("concurrency"))

	// Mark required flags
	workerCmd.MarkFlagRequired("token")
}

// initConfig initializes Viper configuration
func initConfig() {
	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config file search paths
	viper.AddConfigPath(".")                // Current directory
	viper.AddConfigPath("$HOME/.mel-agent") // User home directory
	viper.AddConfigPath("/etc/mel-agent")   // System-wide config

	// Environment variable configuration
	viper.SetEnvPrefix("MEL") // Prefix for environment variables
	viper.AutomaticEnv()      // Automatically read env vars

	// Support legacy environment variables for backward compatibility
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("worker.server", "MEL_SERVER_URL")
	viper.BindEnv("worker.token", "MEL_WORKER_TOKEN")
	viper.BindEnv("worker.id", "MEL_WORKER_ID")
	viper.BindEnv("database.url", "DATABASE_URL")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("worker.server", "http://localhost:8080")
	viper.SetDefault("worker.concurrency", 5)
	viper.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable")

	// Try to read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore and use defaults/env vars
		} else {
			// Config file was found but another error was produced
			log.Printf("Error reading config file: %v", err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func startServer(port string) {
	// connect database (fatal on error)
	db.Connect()

	// load connection plugins from the database
	plugin.RegisterConnectionPlugins()
	// register node plugins via injector (core + builder)
	for _, p := range injector.InitializeNodePlugins() {
		plugin.Register(p)
	}

	// initialize MEL instance for durable workflow execution
	mel := api.NewMel()

	// create workflow engine factory function
	workflowEngineFactory := httpApi.InitializeWorkflowEngine(db.DB, mel)

	// create durable workflow execution engine
	workflowEngine := execution.NewDurableExecutionEngine(db.DB, mel, "api-server")

	// create cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start durable workflow workers
	workerConfig := execution.DefaultWorkerConfig()
	worker := execution.NewWorker(db.DB, mel, workerConfig)
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Printf("Workflow worker error: %v", err)
		}
	}()

	// start trigger scheduler engine
	scheduler := triggers.NewEngine()
	scheduler.Start(ctx)

	// Note: Legacy runner disabled since we dropped agent_runs table
	// The new durable workflow execution system handles all workflow processing

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// webhook entrypoint for external events (e.g., GitHub, Stripe) – accept all HTTP methods
	r.HandleFunc("/webhooks/{provider}/{triggerID}", httpApi.WebhookHandler)

	// Create an efficient API handler that routes without response buffering
	// Route based on path analysis since we know the exact route patterns
	mainAPIHandler := httpApi.Handler()
	workflowHandler := workflowEngineFactory(workflowEngine)

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Chi Mount passes the full path including /api prefix
		// Workflow engine only handles /api/workflow-runs* routes
		// Everything else goes to main API - this is more efficient than buffering
		if strings.HasPrefix(r.URL.Path, "/api/workflow-runs") {
			// Strip /api prefix for workflow handler since it expects /workflow-runs
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			workflowHandler.ServeHTTP(w, r)
		} else {
			// Strip /api prefix for main API handler as well
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			mainAPIHandler.ServeHTTP(w, r)
		}
	})

	r.Mount("/api", apiHandler)

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func startAPIServer(port string) {
	// connect database (fatal on error)
	db.Connect()

	// load connection plugins from the database
	plugin.RegisterConnectionPlugins()
	// register node plugins via injector (core + builder)
	for _, p := range injector.InitializeNodePlugins() {
		plugin.Register(p)
	}

	// initialize MEL instance for durable workflow execution
	mel := api.NewMel()

	// create workflow engine factory function
	workflowEngineFactory := httpApi.InitializeWorkflowEngine(db.DB, mel)

	// create durable workflow execution engine
	workflowEngine := execution.NewDurableExecutionEngine(db.DB, mel, "api-server")

	// create cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// NOTE: No embedded workers started in api-server mode
	log.Printf("Starting API server only (no embedded workers)")

	// start trigger scheduler engine
	scheduler := triggers.NewEngine()
	scheduler.Start(ctx)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// webhook entrypoint for external events (e.g., GitHub, Stripe) – accept all HTTP methods
	r.HandleFunc("/webhooks/{provider}/{triggerID}", httpApi.WebhookHandler)

	// Create an efficient API handler that routes without response buffering
	// Route based on path analysis since we know the exact route patterns
	mainAPIHandler := httpApi.Handler()
	workflowHandler := workflowEngineFactory(workflowEngine)

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Chi Mount passes the full path including /api prefix
		// Workflow engine only handles /api/workflow-runs* routes
		// Everything else goes to main API - this is more efficient than buffering
		if strings.HasPrefix(r.URL.Path, "/api/workflow-runs") {
			// Strip /api prefix for workflow handler since it expects /workflow-runs
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			workflowHandler.ServeHTTP(w, r)
		} else {
			// Strip /api prefix for main API handler as well
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			mainAPIHandler.ServeHTTP(w, r)
		}
	})

	r.Mount("/api", apiHandler)

	log.Printf("API server listening on :%s (no embedded workers)", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func startWorker(serverURL, token, workerID string, concurrency int) {
	// Generate worker ID if not provided
	if workerID == "" {
		workerID = generateWorkerID()
	}

	log.Printf("Starting worker %s connecting to %s", workerID, serverURL)

	// Initialize MEL instance for workflow execution
	mel := api.NewMel()

	// Create a remote worker that connects to the API server
	remoteWorker := execution.NewRemoteWorker(serverURL, token, workerID, mel, concurrency)

	// Create cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the remote worker
	log.Printf("Worker %s starting with concurrency %d", workerID, concurrency)
	if err := remoteWorker.Start(ctx); err != nil {
		log.Fatalf("Worker failed to start: %v", err)
	}
}

func generateWorkerID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("worker-%d", time.Now().Unix())
	}
	return fmt.Sprintf("worker-%s", hex.EncodeToString(bytes))
}
