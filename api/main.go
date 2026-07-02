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
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				handlers.StatusHandler(w, r)
			case http.MethodDelete, http.MethodOptions:
				handlers.DeleteHandler(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 3 && parts[2] == "scale" {
			if r.Method == http.MethodPost || r.Method == http.MethodOptions {
				handlers.ScaleHandler(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 3 && parts[2] == "pods" {
			if r.Method == http.MethodGet || r.Method == http.MethodOptions {
				handlers.PodsHandler(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 5 && parts[2] == "pods" && parts[4] == "logs" {
			if r.Method == http.MethodGet || r.Method == http.MethodOptions {
				handlers.PodLogsHandler(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.NotFound(w, r)
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
