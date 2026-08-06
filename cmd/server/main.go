package main

// Standard library + third‑party imports
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Import credential definitions to register them
	_ "github.com/cedricziel/mel-agent/pkg/credentials"
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

	// health endpoint for load balancers
	r.Get("/health", healthCheckHandler)

	// readiness endpoint for Kubernetes
	r.Get("/ready", readinessCheckHandler)

	// webhook entrypoint is now handled by OpenAPI at /api/webhooks/{token}

	// Use combined OpenAPI + Legacy router for gradual migration
	combinedAPIHandler := httpApi.NewCombinedRouter(db.DB, workflowEngine)

	r.Mount("/", combinedAPIHandler)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cancel context to stop workers and scheduler
	cancel()

	// Give the server a maximum of 30 seconds to finish ongoing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
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

	// health endpoint for load balancers
	r.Get("/health", healthCheckHandler)

	// readiness endpoint for Kubernetes
	r.Get("/ready", readinessCheckHandler)

	// webhook entrypoint is now handled by OpenAPI at /api/webhooks/{token}

	// Use combined OpenAPI + Legacy router for gradual migration
	combinedAPIHandler := httpApi.NewCombinedRouter(db.DB, workflowEngine)

	r.Mount("/", combinedAPIHandler)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API server listening on :%s (no embedded workers)", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API server...")

	// Cancel context to stop scheduler
	cancel()

	// Give the server a maximum of 30 seconds to finish ongoing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("API server forced to shutdown: %v", err)
	} else {
		log.Println("API server exited gracefully")
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
	remoteWorker, err := execution.NewRemoteWorker(serverURL, token, workerID, mel, concurrency)
	if err != nil {
		log.Fatalf("Failed to create remote worker: %v", err)
	}

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
		log.Printf("Warning: Failed to generate random worker ID, using timestamp fallback: %v", err)
		return fmt.Sprintf("worker-%d", time.Now().Unix())
	}
	return fmt.Sprintf("worker-%s", hex.EncodeToString(bytes))
}

// healthCheckHandler provides a basic health check for load balancers
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// readinessCheckHandler provides a comprehensive readiness check including database connectivity
func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	type HealthStatus struct {
		Status    string                 `json:"status"`
		Timestamp string                 `json:"timestamp"`
		Checks    map[string]interface{} `json:"checks"`
	}

	checks := make(map[string]interface{})
	overallStatus := "ready"

	// Check database connectivity
	if db.DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.DB.PingContext(ctx); err != nil {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			overallStatus = "not_ready"
		} else {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status": "not_initialized",
		}
		overallStatus = "not_ready"
	}

	response := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")

	if overallStatus == "ready" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Safe JSON marshaling with proper error handling
	responseBytes, err := json.Marshal(response)
	if err != nil {
		// Fallback to simple response if marshaling fails
		log.Printf("Warning: Failed to marshal readiness response: %v", err)
		fallbackResponse := fmt.Sprintf(`{"status":"%s","timestamp":"%s","error":"marshaling_failed"}`,
			overallStatus, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(fallbackResponse))
		return
	}

	w.Write(responseBytes)
}
