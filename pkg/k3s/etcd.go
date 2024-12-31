package k3s

import (
	"fmt"
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallEtcd installs etcd into the Kubernetes cluster
func InstallEtcd(host, user, keyPath string) error {
	log.Println("Installing etcd cluster...")

	// Add the Bitnami Helm repository
	cmd := "helm repo add bitnami https://charts.bitnami.com/bitnami"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to add bitnami helm repo: %w", err)
	}

	// Update Helm repositories
	cmd = "helm repo update"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}

	// Create namespace for etcd
	cmd = "kubectl create namespace etcd"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0)
	if err != nil && !strings.Contains(out, "already exists") {
		return fmt.Errorf("failed to create etcd namespace: %s: %w", out, err)
	}

	// Install etcd using Helm
	cmd = "helm install etcd bitnami/etcd " +
		"--namespace etcd " +
		"--set replicaCount=2 " +
		"--set persistence.enabled=true " +
		"--set persistence.size=1Gi"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to install etcd: %s: %w", out, err)
	}

	// Wait for etcd pods to be ready
	cmd = "kubectl wait --namespace etcd --for=condition=ready pod -l app.kubernetes.io/name=etcd --timeout=300s"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed waiting for etcd pods: %s: %w", out, err)
	}

	log.Println("etcd installation completed successfully")
	return nil
}
