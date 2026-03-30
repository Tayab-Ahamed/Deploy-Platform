package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"deploy-platform/builder"
)

// StatusHandler handles GET /deploy/{id}
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Extract ID from path: /deploy/<id>
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		jsonError(w, "deploy id required", http.StatusBadRequest)
		return
	}
	id := parts[1]

	job, ok := builder.Store.Get(id)
	if !ok {
		jsonError(w, "deployment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListHandler handles GET /deploy
func ListHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	all := builder.Store.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}
