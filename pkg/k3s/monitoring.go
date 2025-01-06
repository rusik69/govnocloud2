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
	Values           MonitoringValues
	Host             string
	User             string
	Key              string
	GrafanaHost      string
	PrometheusHost   string
	AlertmanagerHost string
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
func NewMonitoringConfig(host, user, key string, grafanaHost, prometheusHost, alertmanagerHost string) *MonitoringConfig {
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
			Name:      "monitoring",
			Chart:     "prometheus-community/kube-prometheus-stack",
			Namespace: "monitoring",
		},
		Values: MonitoringValues{
			Prometheus: PrometheusConfig{
				Service: MonitoringServiceConfig{
					Type: "ClusterIP",
				},
			},
			Grafana: GrafanaConfig{
				Service: MonitoringServiceConfig{
					Type: "ClusterIP",
				},
			},
		},
		Host:             host,
		User:             user,
		Key:              key,
		GrafanaHost:      grafanaHost,
		PrometheusHost:   prometheusHost,
		AlertmanagerHost: alertmanagerHost,
	}
}

// DeployPrometheus deploys Prometheus Operator stack to k3s cluster
func DeployPrometheus(host, user, key string, grafanaHost, prometheusHost, alertmanagerHost string) error {

	cfg := NewMonitoringConfig(host, user, key, grafanaHost, prometheusHost, alertmanagerHost)

	if err := createMonitoringNamespace(cfg); err != nil {
		return fmt.Errorf("failed to create monitoring namespace: %w", err)
	}

	return deployMonitoringStack(cfg)
}

// createMonitoringNamespace ensures the monitoring namespace exists
func createMonitoringNamespace(cfg *MonitoringConfig) error {
	cmd := "kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -"
	log.Println(cmd)
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

	if err := installMonitoringChart(cfg, valuesFile); err != nil {
		return err
	}

	// Wait for pods to be ready
	log.Println("Waiting for monitoring pods to be ready...")
	cmd := "kubectl -n monitoring wait --for=condition=ready pod --all --timeout=300s"
	if _, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 0); err != nil {
		log.Printf("Warning: Some pods are not ready: %v", err)
	}

	// Create ingresses
	if err := createMonitoringIngresses(cfg); err != nil {
		return err
	}

	return nil
}

// addHelmRepo adds the Prometheus Helm repository
func addHelmRepo(cfg *MonitoringConfig) error {
	cmd := fmt.Sprintf("helm repo add %s %s", cfg.HelmRepo.Name, cfg.HelmRepo.URL)
	log.Println(cmd)
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
	log.Println(cmd)
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
grafana:
  persistence:
    enabled: true
    size: 5Gi
  service:
    type: ClusterIP
prometheus:
  prometheusSpec:
    retention: 30d
  persistence:
    enabled: true
    size: 5Gi
alertmanager:
  alertmanagerSpec: {}
  persistence:
    enabled: true
    size: 5Gi
prometheusOperator:
  admissionWebhooks:
    enabled: false
  tls:
    enabled: false
defaultRules:
  create: true
  rules:
    alertmanager: true
    etcd: true
    configReloaders: true
    general: true
    k8s: true
    kubeApiserver: true
    kubePrometheusNodeAlerting: true
    kubePrometheusNodeRecording: true
    kubernetesAbsent: true
    kubernetesApps: true
    kubernetesResources: true
    kubernetesStorage: true
    kubernetesSystem: true
    kubeScheduler: true
    network: true
    node: true
    prometheus: true
    prometheusOperator: true
    time: true
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
	cmd := fmt.Sprintf("helm upgrade --install %s %s --namespace %s --create-namespace --values %s --wait --timeout 600s",
		cfg.Release.Name,
		cfg.Release.Chart,
		cfg.Release.Namespace,
		valuesFile,
	)
	log.Println(cmd)
	_, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to install monitoring stack: %w", err)
	}

	return nil
}

// createMonitoringIngresses creates ingress resources for monitoring components
func createMonitoringIngresses(cfg *MonitoringConfig) error {
	ingressYaml := fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: monitoring-ingress
  namespace: monitoring
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
spec:
  rules:
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: monitoring-grafana
            port:
              number: 80
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: monitoring-kube-prometheus-prometheus
            port:
              number: 9090
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: monitoring-kube-prometheus-alertmanager
            port:
              number: 9093
`, cfg.GrafanaHost, cfg.PrometheusHost, cfg.AlertmanagerHost)

	// Write and apply ingress configuration
	cmd := fmt.Sprintf("cat << 'EOF' > /tmp/monitoring-ingress.yaml\n%s\nEOF", ingressYaml)
	log.Println("Creating monitoring ingress YAML")
	if _, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 0); err != nil {
		return fmt.Errorf("failed to create ingress YAML: %w", err)
	}

	cmd = "kubectl apply -f /tmp/monitoring-ingress.yaml -n monitoring"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 0); err != nil {
		return fmt.Errorf("failed to apply monitoring ingress: %s: %w", out, err)
	}

	log.Println("Monitoring stack is accessible at:")
	log.Printf("- Grafana: http://%s", cfg.GrafanaHost)
	log.Printf("- Prometheus: http://%s", cfg.PrometheusHost)
	log.Printf("- Alertmanager: http://%s", cfg.AlertmanagerHost)

	return nil
}
