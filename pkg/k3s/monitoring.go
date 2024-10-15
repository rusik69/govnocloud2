package k3s

import (
	"fmt"
	"os"
	"os/exec"
)

// DeployPrometheus deploys Prometheus to k3s cluster.
func DeployPrometheus() error {
	cmd := "helm repo add prometheus-community https://prometheus-community.github.io/helm-charts"
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error adding prometheus-community repo: %w", err)
	}
	command = exec.Command("helm", "repo", "update")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error updating helm repos: %w", err)
	}
	valuesYaml := `
prometheus:
  service:
    type: NodePort
grafana:
  service:
    type: NodePort
`
	valueYamlFile, err := os.CreateTemp("", "values-*.yaml")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(valueYamlFile.Name())
	if _, err := valueYamlFile.WriteString(valuesYaml); err != nil {
		return fmt.Errorf("error writing to temporary file: %w", err)
	}
	command = exec.Command("helm", "upgrade", "--install", "prometheus", "prometheus-community/prometheus", "--values", valueYamlFile.Name())
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing Prometheus: %w", err)
	}
	return nil
}
