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

	httpApi "github.com/cedricziel/mel-agent/internal/api"
	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/injector"
	"github.com/cedricziel/mel-agent/internal/triggers"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/cedricziel/mel-agent/pkg/plugin"
)

func main() {
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
		port, _ := cmd.Flags().GetString("port")
		startServer(port)
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
		serverURL, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")
		workerID, _ := cmd.Flags().GetString("id")
		concurrency, _ := cmd.Flags().GetInt("concurrency")

		if token == "" {
			log.Fatal("Worker token is required. Set MEL_WORKER_TOKEN environment variable or use --token flag")
		}

		startWorker(serverURL, token, workerID, concurrency)
	},
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(workerCmd)

	// Server command flags
	serverCmd.Flags().StringP("port", "p", getEnvOrDefault("PORT", "8080"), "Port to listen on")

	// Worker command flags
	workerCmd.Flags().StringP("server", "s", getEnvOrDefault("MEL_SERVER_URL", "http://localhost:8080"), "API server URL")
	workerCmd.Flags().StringP("token", "t", getEnvOrDefault("MEL_WORKER_TOKEN", ""), "Authentication token (required)")
	workerCmd.Flags().String("id", getEnvOrDefault("MEL_WORKER_ID", ""), "Worker ID (auto-generated if empty)")
	workerCmd.Flags().IntP("concurrency", "c", 5, "Number of concurrent workflow executions")

	// Mark required flags
	workerCmd.MarkFlagRequired("token")
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
