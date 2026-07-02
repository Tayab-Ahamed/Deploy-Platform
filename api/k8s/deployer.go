package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"deploy-platform/models"
)

const (
	Namespace  = "deployments"
	BaseDomain = "localhost"
)

// Deploy creates or updates a Deployment, Service, and Ingress for the given job.
func Deploy(job *models.DeployJob) (string, error) {
	c, err := NewClient()
	if err != nil {
		return "", fmt.Errorf("k8s client: %w", err)
	}

	name := job.AppName
	image := job.ImageTag

	if err := applyDeployment(c, name, image, job.Replicas, job.EnvVars); err != nil {
		return "", fmt.Errorf("deployment: %w", err)
	}
	if err := applyService(c, name); err != nil {
		return "", fmt.Errorf("service: %w", err)
	}
	if err := applyIngress(c, name); err != nil {
		return "", fmt.Errorf("ingress: %w", err)
	}

	return fmt.Sprintf("http://%s.%s", name, BaseDomain), nil
}

// ── Deployment ────────────────────────────────────────────────────────

func deploymentJSON(name, image string, replicas int, envVars map[string]string) string {
	if replicas <= 0 {
		replicas = 1
	}
	var envBlock string
	if len(envVars) > 0 {
		var envList []string
		for k, v := range envVars {
			envList = append(envList, fmt.Sprintf(`{"name": %q, "value": %q}`, k, v))
		}
		envBlock = fmt.Sprintf(`, "env": [%s]`, strings.Join(envList, ","))
	}

	return fmt.Sprintf(`{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": %q,
    "namespace": %q,
    "labels": {"app": %q, "managed-by": "deploy-platform"}
  },
  "spec": {
    "replicas": %d,
    "selector": {"matchLabels": {"app": %q}},
    "template": {
      "metadata": {"labels": {"app": %q}},
      "spec": {
        "containers": [{
          "name": %q,
          "image": %q,
          "imagePullPolicy": "Always",
          "ports": [{"containerPort": 8080, "protocol": "TCP"}]
          %s
        }]
      }
    }
  }
}`, name, Namespace, name, replicas, name, name, name, image, envBlock)
}

func applyDeployment(c *K8sClient, name, image string, replicas int, envVars map[string]string) error {
	body := deploymentJSON(name, image, replicas, envVars)
	collURL := fmt.Sprintf("%s/apis/apps/v1/namespaces/%s/deployments", c.BaseURL, Namespace)
	resURL := collURL + "/" + name
	return applyResource(c, collURL, resURL, body)
}

// ── Service ───────────────────────────────────────────────────────────

func serviceJSON(name string) string {
	return fmt.Sprintf(`{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "name": %q,
    "namespace": %q,
    "labels": {"app": %q, "managed-by": "deploy-platform"}
  },
  "spec": {
    "selector": {"app": %q},
    "ports": [{"port": 80, "targetPort": 8080, "protocol": "TCP"}],
    "type": "ClusterIP"
  }
}`, name, Namespace, name, name)
}

func applyService(c *K8sClient, name string) error {
	body := serviceJSON(name)
	collURL := fmt.Sprintf("%s/api/v1/namespaces/%s/services", c.BaseURL, Namespace)
	resURL := collURL + "/" + name
	return applyResource(c, collURL, resURL, body)
}

// ── Ingress ───────────────────────────────────────────────────────────

func ingressJSON(name string) string {
	host := fmt.Sprintf("%s.%s", name, BaseDomain)
	return fmt.Sprintf(`{
  "apiVersion": "networking.k8s.io/v1",
  "kind": "Ingress",
  "metadata": {
    "name": %q,
    "namespace": %q,
    "labels": {"managed-by": "deploy-platform"},
    "annotations": {"nginx.ingress.kubernetes.io/rewrite-target": "/"}
  },
  "spec": {
    "ingressClassName": "nginx",
    "rules": [{
      "host": %q,
      "http": {
        "paths": [{
          "path": "/",
          "pathType": "Prefix",
          "backend": {"service": {"name": %q, "port": {"number": 80}}}
        }]
      }
    }]
  }
}`, name, Namespace, host, name)
}

func applyIngress(c *K8sClient, name string) error {
	body := ingressJSON(name)
	collURL := fmt.Sprintf("%s/apis/networking.k8s.io/v1/namespaces/%s/ingresses", c.BaseURL, Namespace)
	resURL := collURL + "/" + name
	return applyResource(c, collURL, resURL, body)
}

// ── HTTP helpers ──────────────────────────────────────────────────────

// applyResource checks if the resource exists (GET resURL).
// If 404 → POST to collURL (create). Otherwise → PUT to resURL (update).
func applyResource(c *K8sClient, collURL, resURL, body string) error {
	resp, err := c.HTTPClient.Get(resURL)
	if err != nil {
		return fmt.Errorf("GET %s: %w", resURL, err)
	}
	resp.Body.Close()

	var method, targetURL string
	if resp.StatusCode == http.StatusNotFound {
		method, targetURL = http.MethodPost, collURL
	} else {
		method, targetURL = http.MethodPut, resURL
	}

	req, err := http.NewRequest(method, targetURL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, targetURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%s %s → HTTP %d: %s", method, targetURL, res.StatusCode, string(b))
	}
	return nil
}

type PodInfo struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Restarts  int       `json:"restarts"`
	Age       string    `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// Delete cleans up the Deployment, Service, and Ingress for an app.
func Delete(name string) error {
	c, err := NewClient()
	if err != nil {
		return fmt.Errorf("k8s client: %w", err)
	}

	// Delete Ingress
	ingressURL := fmt.Sprintf("%s/apis/networking.k8s.io/v1/namespaces/%s/ingresses/%s", c.BaseURL, Namespace, name)
	_ = deleteResource(c, ingressURL)

	// Delete Service
	serviceURL := fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s", c.BaseURL, Namespace, name)
	_ = deleteResource(c, serviceURL)

	// Delete Deployment
	deployURL := fmt.Sprintf("%s/apis/apps/v1/namespaces/%s/deployments/%s", c.BaseURL, Namespace, name)
	_ = deleteResource(c, deployURL)

	return nil
}

func deleteResource(c *K8sClient, url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// GetPods queries Kubernetes API to list pods matching the app label.
func GetPods(appName string) ([]PodInfo, error) {
	c, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("k8s client: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/pods?labelSelector=app%%3D%s", c.BaseURL, Namespace, appName)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get pods (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name              string    `json:"name"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					RestartCount int `json:"restartCount"`
					State        struct {
						Waiting *struct {
							Reason string `json:"reason"`
						} `json:"waiting"`
					} `json:"state"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&podList); err != nil {
		return nil, err
	}

	var pods []PodInfo
	for _, item := range podList.Items {
		restarts := 0
		status := item.Status.Phase
		if len(item.Status.ContainerStatuses) > 0 {
			restarts = item.Status.ContainerStatuses[0].RestartCount
			if item.Status.ContainerStatuses[0].State.Waiting != nil {
				status = item.Status.ContainerStatuses[0].State.Waiting.Reason
			}
		}

		age := time.Since(item.Metadata.CreationTimestamp).Round(time.Second).String()

		pods = append(pods, PodInfo{
			Name:      item.Metadata.Name,
			Status:    status,
			Restarts:  restarts,
			Age:       age,
			CreatedAt: item.Metadata.CreationTimestamp,
		})
	}

	return pods, nil
}

// GetPodLogs gets the last 100 log lines of a specific pod.
func GetPodLogs(podName string) (string, error) {
	c, err := NewClient()
	if err != nil {
		return "", fmt.Errorf("k8s client: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/log?tailLines=100", c.BaseURL, Namespace, podName)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get pod logs (HTTP %d): %s", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}


