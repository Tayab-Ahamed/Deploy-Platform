package k8s

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

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

	if err := applyDeployment(c, name, image); err != nil {
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

func deploymentJSON(name, image string) string {
	return fmt.Sprintf(`{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": %q,
    "namespace": %q,
    "labels": {"app": %q, "managed-by": "deploy-platform"}
  },
  "spec": {
    "replicas": 1,
    "selector": {"matchLabels": {"app": %q}},
    "template": {
      "metadata": {"labels": {"app": %q}},
      "spec": {
        "containers": [{
          "name": %q,
          "image": %q,
          "imagePullPolicy": "Always",
          "ports": [{"containerPort": 8080, "protocol": "TCP"}]
        }]
      }
    }
  }
}`, name, Namespace, name, name, name, name, image)
}

func applyDeployment(c *K8sClient, name, image string) error {
	body := deploymentJSON(name, image)
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


