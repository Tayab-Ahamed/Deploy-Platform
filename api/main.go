package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"deploy-platform/builder"
	"deploy-platform/handlers"
)

func main() {
	builder.StartWorkers(5)
	log.Println("✓ 5 build workers started")

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Route /deploy and /deploy/<id>
	mux.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodOptions:
			handlers.DeployHandler(w, r)
		case http.MethodGet:
			handlers.ListHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/deploy/", func(w http.ResponseWriter, r *http.Request) {
		// Reject deeper paths like /deploy/id/extra
		if strings.Count(strings.Trim(r.URL.Path, "/"), "/") > 1 {
			http.NotFound(w, r)
			return
		}
		handlers.StatusHandler(w, r)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Minute, // large uploads need time
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("✓ API listening on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
