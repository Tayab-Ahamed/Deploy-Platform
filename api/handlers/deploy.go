package handlers

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"deploy-platform/builder"
	"deploy-platform/models"
)

var validAppName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)

func DeployHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		jsonError(w, "request too large or not multipart", http.StatusBadRequest)
		return
	}

	appName := strings.TrimSpace(r.FormValue("app_name"))
	if !validAppName.MatchString(appName) {
		jsonError(w, "app_name: 3-30 lowercase letters/numbers/hyphens, cannot start or end with hyphen", http.StatusBadRequest)
		return
	}

	envVarsJSON := strings.TrimSpace(r.FormValue("env_vars"))
	var envVars map[string]string
	if envVarsJSON != "" {
		if err := json.Unmarshal([]byte(envVarsJSON), &envVars); err != nil {
			jsonError(w, "invalid env_vars JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	file, _, err := r.FormFile("source")
	if err != nil {
		jsonError(w, "source zip file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	jobID := builder.NewJobID()
	tmpDir, err := os.MkdirTemp("", "deploy-"+jobID)
	if err != nil {
		jsonError(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	zipPath := filepath.Join(tmpDir, "source.zip")
	dst, err := os.Create(zipPath)
	if err != nil {
		jsonError(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = io.Copy(dst, file); err != nil {
		dst.Close()
		jsonError(w, "server error writing file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dst.Close()

	srcDir := filepath.Join(tmpDir, "src")
	if err = os.MkdirAll(srcDir, 0755); err != nil {
		jsonError(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err = unzip(zipPath, srcDir); err != nil {
		jsonError(w, "failed to extract zip: "+err.Error(), http.StatusBadRequest)
		return
	}

	job := &models.DeployJob{
		ID:        jobID,
		AppName:   appName,
		SourceDir: srcDir,
		Status:    models.StatusQueued,
		Logs:      []string{"Job queued — waiting for build worker..."},
		CreatedAt: time.Now(),
		EnvVars:   envVars,
		Replicas:  1,
	}
	builder.Store.Set(job)
	builder.BuildQueue <- job

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"deploy_id": jobID,
		"status":    string(models.StatusQueued),
	})
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rel := filepath.Clean(f.Name)
		if strings.HasPrefix(rel, "..") {
			continue // zip-slip guard
		}
		target := filepath.Join(dest, rel)
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
