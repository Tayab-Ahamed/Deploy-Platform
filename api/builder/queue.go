package builder

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"deploy-platform/models"
)

// newID generates a random 16-byte hex string as a job ID.
func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var dataDir = "/data/jobs"

// JobStore is a thread-safe persistent store for deploy jobs.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*models.DeployJob
}

var Store *JobStore

func init() {
	if os.PathSeparator == '\\' {
		dataDir = "./data/jobs"
	}
	Store = &JobStore{jobs: make(map[string]*models.DeployJob)}
	Store.loadAll()
}

func (s *JobStore) loadAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dataDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var job models.DeployJob
		if err := json.Unmarshal(data, &job); err != nil {
			continue
		}
		s.jobs[job.ID] = &job
	}
}

func (s *JobStore) save(job *models.DeployJob) {
	path := filepath.Join(dataDir, job.ID+".json")
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}

func (s *JobStore) Set(job *models.DeployJob) {
	s.mu.Lock()
	cp := *job
	logs := make([]string, len(job.Logs))
	copy(logs, job.Logs)
	cp.Logs = logs

	if job.EnvVars != nil {
		envCopy := make(map[string]string, len(job.EnvVars))
		for k, v := range job.EnvVars {
			envCopy[k] = v
		}
		cp.EnvVars = envCopy
	}

	s.jobs[job.ID] = &cp
	s.mu.Unlock()

	s.save(&cp)
}

func (s *JobStore) Get(id string) (*models.DeployJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	logs := make([]string, len(j.Logs))
	copy(logs, j.Logs)
	cp.Logs = logs

	if j.EnvVars != nil {
		envCopy := make(map[string]string, len(j.EnvVars))
		for k, v := range j.EnvVars {
			envCopy[k] = v
		}
		cp.EnvVars = envCopy
	}

	return &cp, true
}

func (s *JobStore) All() []*models.DeployJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.DeployJob, 0, len(s.jobs))
	for _, j := range s.jobs {
		cp := *j
		logs := make([]string, len(j.Logs))
		copy(logs, j.Logs)
		cp.Logs = logs

		if j.EnvVars != nil {
			envCopy := make(map[string]string, len(j.EnvVars))
			for k, v := range j.EnvVars {
				envCopy[k] = v
			}
			cp.EnvVars = envCopy
		}

		out = append(out, &cp)
	}
	return out
}

func (s *JobStore) Delete(id string) {
	s.mu.Lock()
	delete(s.jobs, id)
	s.mu.Unlock()

	path := filepath.Join(dataDir, id+".json")
	_ = os.Remove(path)
}

// NewJobID returns a fresh unique ID.
func NewJobID() string { return newID() }

// BuildQueue is a buffered channel acting as the work queue.
var BuildQueue = make(chan *models.DeployJob, 50)

// StartWorkers launches n goroutines that consume BuildQueue.
func StartWorkers(n int) {
	for i := 0; i < n; i++ {
		go func() {
			for job := range BuildQueue {
				RunBuild(job)
			}
		}()
	}
}
