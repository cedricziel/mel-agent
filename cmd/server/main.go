package main

// Standard library + third‑party imports
import (
   "log"
   "net/http"
   "os"

   "github.com/go-chi/chi/v5"
   "github.com/go-chi/chi/v5/middleware"

   "github.com/cedricziel/mel-agent/pkg/api"
   _ "github.com/cedricziel/mel-agent/pkg/plugin/adapters" // register node plugins adapters
   _ "github.com/cedricziel/mel-agent/internal/api/nodes"  // import legacy builder node definitions
   "github.com/cedricziel/mel-agent/internal/db"
   "github.com/cedricziel/mel-agent/pkg/plugin"
   "github.com/cedricziel/mel-agent/internal/triggers"
   "github.com/cedricziel/mel-agent/internal/runs"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

    // connect database (fatal on error)
    db.Connect()
    // load connection plugins from the database
    plugin.RegisterConnectionPlugins()
    // start trigger scheduler engine
    scheduler := triggers.NewEngine()
    scheduler.Start()
    // start run processor to execute enqueued workflows
    runner := runs.NewRunner()
    runner.Start()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

   // health endpoint
   r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

   // webhook entrypoint for external events (e.g., GitHub, Stripe) – accept all HTTP methods
   r.HandleFunc("/webhooks/{provider}/{triggerID}", api.WebhookHandler)
   // mount api routes under /api
   r.Mount("/api", api.Handler())

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
