package k3s

import (
	"fmt"
	"log"
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// MonitoringConfig holds configuration for monitoring deployment
type MonitoringConfig struct {
	HelmRepo struct {
		Name string
		URL  string
	}
	Release struct {
		Name      string
		Chart     string
		Namespace string
	}
	Values MonitoringValues
	Host   string
	User   string
	Key    string
}

// MonitoringValues represents the Helm values for monitoring
type MonitoringValues struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Grafana    GrafanaConfig    `yaml:"grafana"`
}

// PrometheusConfig represents Prometheus specific configuration
type PrometheusConfig struct {
	Service MonitoringServiceConfig `yaml:"service"`
}

// GrafanaConfig represents Grafana specific configuration
type GrafanaConfig struct {
	Service MonitoringServiceConfig `yaml:"service"`
}

// MonitoringServiceConfig represents Kubernetes service configuration
type MonitoringServiceConfig struct {
	Type string `yaml:"type"`
}

// NewMonitoringConfig creates a default monitoring configuration
func NewMonitoringConfig(host, user, key string) *MonitoringConfig {
	return &MonitoringConfig{
		HelmRepo: struct {
			Name string
			URL  string
		}{
			Name: "prometheus-community",
			URL:  "https://prometheus-community.github.io/helm-charts",
		},
		Release: struct {
			Name      string
			Chart     string
			Namespace string
		}{
			Name:      "prometheus",
			Chart:     "prometheus-community/prometheus",
			Namespace: "monitoring",
		},
		Values: MonitoringValues{
			Prometheus: PrometheusConfig{
				Service: MonitoringServiceConfig{
					Type: "NodePort",
				},
			},
			Grafana: GrafanaConfig{
				Service: MonitoringServiceConfig{
					Type: "NodePort",
				},
			},
		},
		Host: host,
		User: user,
		Key:  key,
	}
}

// DeployPrometheus deploys Prometheus to k3s cluster.
func DeployPrometheus(host, user, key string) error {
	cfg := NewMonitoringConfig(host, user, key)

	// Create monitoring namespace first
	if err := createMonitoringNamespace(cfg); err != nil {
		return fmt.Errorf("failed to create monitoring namespace: %w", err)
	}

	// Wait for monitoring namespace to be ready
	if _, err := ssh.Run("kubectl wait --for=condition=ready --timeout=300s pod -l app=prometheus -n monitoring", cfg.Host, cfg.Key, cfg.User, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for monitoring namespace to be ready: %w", err)
	}

	return deployMonitoringStack(cfg)
}

// createMonitoringNamespace ensures the monitoring namespace exists
func createMonitoringNamespace(cfg *MonitoringConfig) error {
	cmd := "kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -"
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to create monitoring namespace: %w", err)
	}
	log.Println(out)
	return nil
}

// deployMonitoringStack handles the actual deployment of monitoring components
func deployMonitoringStack(cfg *MonitoringConfig) error {
	if err := addHelmRepo(cfg); err != nil {
		return err
	}

	if err := updateHelmRepos(cfg); err != nil {
		return err
	}

	valuesFile, err := createValuesFile(cfg.Values)
	if err != nil {
		return err
	}
	defer os.Remove(valuesFile)

	return installMonitoringChart(cfg, valuesFile)
}

// addHelmRepo adds the Prometheus Helm repository
func addHelmRepo(cfg *MonitoringConfig) error {
	cmd := fmt.Sprintf("helm repo add %s %s", cfg.HelmRepo.Name, cfg.HelmRepo.URL)
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to add Helm repository: %w", err)
	}
	log.Println(out)

	return nil
}

// updateHelmRepos updates all Helm repositories
func updateHelmRepos(cfg *MonitoringConfig) error {
	cmd := "helm repo update"
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to update Helm repositories: %w", err)
	}
	log.Println(out)

	return nil
}

// createValuesFile creates a temporary values file for Helm
func createValuesFile(values MonitoringValues) (string, error) {
	valuesYaml := `
prometheus:
  service:
    type: NodePort
grafana:
  service:
    type: NodePort
  persistence:
    enabled: true
    size: 10Gi
alertmanager:
  service:
    type: NodePort
  persistence:
    enabled: true
    size: 2Gi
server:
  persistentVolume:
    enabled: true
    size: 50Gi
`
	file, err := os.CreateTemp("", "values-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary values file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(valuesYaml); err != nil {
		return "", fmt.Errorf("failed to write values file: %w", err)
	}

	return file.Name(), nil
}

// installMonitoringChart installs the Prometheus chart using Helm
func installMonitoringChart(cfg *MonitoringConfig, valuesFile string) error {
	cmd := fmt.Sprintf("helm upgrade --install %s %s --namespace %s --create-namespace --values %s",
		cfg.Release.Name,
		cfg.Release.Chart,
		cfg.Release.Namespace,
		valuesFile,
	)
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to install monitoring stack: %w", err)
	}
	log.Println(out)

	return nil
}
