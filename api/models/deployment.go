package models

import "time"

type DeployStatus string

const (
	StatusQueued    DeployStatus = "queued"
	StatusBuilding  DeployStatus = "building"
	StatusDeploying DeployStatus = "deploying"
	StatusLive      DeployStatus = "live"
	StatusFailed    DeployStatus = "failed"
)

type DeployJob struct {
	ID        string            `json:"id"`
	AppName   string            `json:"app_name"`
	ImageTag  string            `json:"image_tag"`
	Status    DeployStatus      `json:"status"`
	Logs      []string          `json:"logs"`
	URL       string            `json:"url"`
	CreatedAt time.Time         `json:"created_at"`
	SourceDir string            `json:"-"`
	EnvVars   map[string]string `json:"env_vars,omitempty"`
	Replicas  int               `json:"replicas,omitempty"`
}
