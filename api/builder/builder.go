package builder

import (
	"fmt"
	"os/exec"
	"strings"

	"deploy-platform/k8s"
	"deploy-platform/models"
)

const RegistryHost = "localhost:5000"

func RunBuild(job *models.DeployJob) {
	log := func(line string) {
		job.Logs = append(job.Logs, line)
		Store.Set(job)
	}

	job.Status = models.StatusBuilding
	Store.Set(job)

	imageTag := fmt.Sprintf("%s/%s:%s", RegistryHost, job.AppName, job.ID[:8])
	job.ImageTag = imageTag

	// ── docker build ──────────────────────────────────────────────────
	log("▶ docker build starting...")
	buildCmd := exec.Command("docker", "build", "-t", imageTag, job.SourceDir)
	buildOut, err := buildCmd.CombinedOutput()
	for _, line := range strings.Split(string(buildOut), "\n") {
		if strings.TrimSpace(line) != "" {
			log("  " + line)
		}
	}
	if err != nil {
		log("✗ Build failed: " + err.Error())
		job.Status = models.StatusFailed
		Store.Set(job)
		return
	}
	log("✓ Build succeeded")

	// ── docker push ───────────────────────────────────────────────────
	log("▶ Pushing image to registry...")
	pushCmd := exec.Command("docker", "push", imageTag)
	pushOut, err := pushCmd.CombinedOutput()
	for _, line := range strings.Split(string(pushOut), "\n") {
		if strings.TrimSpace(line) != "" {
			log("  " + line)
		}
	}
	if err != nil {
		log("✗ Push failed: " + err.Error())
		job.Status = models.StatusFailed
		Store.Set(job)
		return
	}
	log("✓ Image pushed: " + imageTag)

	// ── kubernetes deploy ─────────────────────────────────────────────
	job.Status = models.StatusDeploying
	Store.Set(job)
	log("▶ Deploying to Kubernetes...")

	url, err := k8s.Deploy(job)
	if err != nil {
		log("✗ Kubernetes deploy failed: " + err.Error())
		job.Status = models.StatusFailed
		Store.Set(job)
		return
	}

	job.URL = url
	job.Status = models.StatusLive
	log("✓ Live at: " + url)
	Store.Set(job)
}
