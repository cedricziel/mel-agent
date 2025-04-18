package main

// Standard library + third‑party imports
import (
   "log"
   "net/http"
   "os"

   "github.com/go-chi/chi/v5"
   "github.com/go-chi/chi/v5/middleware"

   "github.com/cedricziel/mel-agent/internal/api"
   "github.com/cedricziel/mel-agent/internal/db"
   "github.com/cedricziel/mel-agent/internal/triggers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

    // connect database (fatal on error)
    db.Connect()
    // start trigger scheduler engine
    scheduler := triggers.NewEngine()
    scheduler.Start()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

   // health endpoint
   r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

   // webhook entrypoint for external events (e.g., GitHub, Stripe)
   // POST triggers the execution; GET returns Method Not Allowed with our header for diagnostics
   r.Post("/webhooks/{provider}/{triggerID}", api.WebhookHandler)
  r.Get("/webhooks/{provider}/{triggerID}", func(w http.ResponseWriter, r *http.Request) {
      // Indicate the request hit our engine
      w.Header().Set("X-Agent-Processed", "true")
      w.Header().Set("Allow", "POST")
      http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
  })
   // mount api routes under /api
   r.Mount("/api", api.Handler())

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
