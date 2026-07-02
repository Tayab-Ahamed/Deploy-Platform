package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"deploy-platform/builder"
	"deploy-platform/k8s"
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

// DeleteHandler handles DELETE /deploy/{id}
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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

	// Delete from Kubernetes
	if err := k8s.Delete(job.AppName); err != nil {
		jsonError(w, "failed to delete from kubernetes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from store
	builder.Store.Delete(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// ScaleHandler handles POST /deploy/{id}/scale
func ScaleHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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

	replicasStr := r.URL.Query().Get("replicas")
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil || replicas < 0 {
		jsonError(w, "invalid replicas count", http.StatusBadRequest)
		return
	}

	job.Replicas = replicas
	builder.Store.Set(job)

	// Re-deploy with new replica count (k8s will scale it)
	if _, err := k8s.Deploy(job); err != nil {
		jsonError(w, "failed to scale in kubernetes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// PodsHandler handles GET /deploy/{id}/pods
func PodsHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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

	pods, err := k8s.GetPods(job.AppName)
	if err != nil {
		jsonError(w, "failed to get pods: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods)
}

// PodLogsHandler handles GET /deploy/{id}/pods/{podName}/logs
func PodLogsHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[1] == "" || parts[3] == "" {
		jsonError(w, "deploy id and pod name required", http.StatusBadRequest)
		return
	}
	podName := parts[3]

	logs, err := k8s.GetPodLogs(podName)
	if err != nil {
		jsonError(w, "failed to get pod logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"logs": logs})
}
