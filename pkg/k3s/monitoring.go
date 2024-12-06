package k3s

import (
	"fmt"
	"log"
	"os"
	"os/exec"

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
	Port   string
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
func NewMonitoringConfig(host, port, user, key string) *MonitoringConfig {
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
			Name:  "prometheus",
			Chart: "prometheus-community/prometheus",
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
		Port: port,
		User: user,
		Key:  key,
	}
}

// DeployPrometheus deploys Prometheus to k3s cluster.
func DeployPrometheus(host, port, user, key string) error {
	cfg := NewMonitoringConfig(host, port, user, key)
	return deployMonitoringStack(cfg)
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
	cmd := exec.Command("helm", "upgrade", "--install",
		cfg.Release.Name,
		cfg.Release.Chart,
		"--values", valuesFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install monitoring stack: %w", err)
	}

	return nil
}
