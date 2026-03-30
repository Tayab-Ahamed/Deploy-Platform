// Package k8s calls the Kubernetes REST API directly using only Go stdlib.
// No external dependencies required.
package k8s

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// K8sClient holds connection details for the Kubernetes API server.
type K8sClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient builds a K8sClient from the kubeconfig file.
func NewClient() (*K8sClient, error) {
	path := kubeConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %s: %w", path, err)
	}

	server, caData, certData, keyData, err := parseKubeconfig(data)
	if err != nil {
		return nil, err
	}

	caBytes, err := base64.StdEncoding.DecodeString(caData)
	if err != nil {
		return nil, fmt.Errorf("decode CA: %w", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caBytes)

	certBytes, err := base64.StdEncoding.DecodeString(certData)
	if err != nil {
		return nil, fmt.Errorf("decode client cert: %w", err)
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return nil, fmt.Errorf("decode client key: %w", err)
	}

	pair, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse client keypair: %w", err)
	}

	return &K8sClient{
		BaseURL: server,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      caPool,
					Certificates: []tls.Certificate{pair},
				},
			},
		},
	}, nil
}

// parseKubeconfig does a minimal line-by-line parse of the kubeconfig YAML.
// It only extracts the four fields the deployer needs.
func parseKubeconfig(data []byte) (server, caData, certData, keyData string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		k, v := splitKV(line)
		switch k {
		case "server":
			server = v
		case "certificate-authority-data":
			caData = v
		case "client-certificate-data":
			certData = v
		case "client-key-data":
			keyData = v
		}
	}
	if server == "" {
		err = fmt.Errorf("kubeconfig: server not found")
	}
	return
}

func splitKV(line string) (key, val string) {
	line = strings.TrimSpace(line)
	i := strings.Index(line, ":")
	if i < 0 {
		return "", ""
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:])
}

func kubeConfigPath() string {
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		return kc
	}
	if _, e := os.Stat("/etc/rancher/k3s/k3s.yaml"); e == nil {
		return "/etc/rancher/k3s/k3s.yaml"
	}
	home, _ := os.UserHomeDir()
	return home + "/.kube/config"
}
