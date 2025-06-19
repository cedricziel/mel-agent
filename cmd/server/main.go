package main

// Standard library + third‑party imports
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpApi "github.com/cedricziel/mel-agent/internal/api"
	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/injector"
	"github.com/cedricziel/mel-agent/internal/triggers"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/cedricziel/mel-agent/pkg/plugin"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		runServer()
	case "worker":
		runWorker()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: mel-agent <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  server    Start the API server")
	fmt.Println("  worker    Start a workflow worker")
	fmt.Println("  help      Show this help message")
	fmt.Println("")
	fmt.Println("Use 'mel-agent <command> -h' for command-specific help")
}

func runServer() {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	port := serverCmd.String("port", getEnvOrDefault("PORT", "8080"), "Port to listen on")

	serverCmd.Parse(os.Args[2:])

	startServer(*port)
}

func runWorker() {
	workerCmd := flag.NewFlagSet("worker", flag.ExitOnError)
	serverURL := workerCmd.String("server", getEnvOrDefault("MEL_SERVER_URL", "http://localhost:8080"), "API server URL")
	token := workerCmd.String("token", getEnvOrDefault("MEL_WORKER_TOKEN", ""), "Authentication token")
	workerID := workerCmd.String("id", getEnvOrDefault("MEL_WORKER_ID", ""), "Worker ID (auto-generated if empty)")
	concurrency := workerCmd.Int("concurrency", 5, "Number of concurrent workflow executions")

	workerCmd.Parse(os.Args[2:])

	if *token == "" {
		log.Fatal("Worker token is required. Set MEL_WORKER_TOKEN environment variable or use -token flag")
	}

	startWorker(*serverURL, *token, *workerID, *concurrency)
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
