package builder

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"deploy-platform/models"
)

// newID generates a random 16-byte hex string as a job ID.
func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// JobStore is a thread-safe in-memory store for deploy jobs.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*models.DeployJob
}

var Store = &JobStore{jobs: make(map[string]*models.DeployJob)}

func (s *JobStore) Set(job *models.DeployJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *job
	logs := make([]string, len(job.Logs))
	copy(logs, job.Logs)
	cp.Logs = logs
	s.jobs[job.ID] = &cp
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
		out = append(out, &cp)
	}
	return out
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
