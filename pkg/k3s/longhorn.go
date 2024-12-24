package k3s

import (
	"fmt"
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallLonghorn installs Longhorn storage system into the Kubernetes cluster
func InstallLonghorn(host, user, keyPath string) error {
	log.Println("Installing Longhorn storage system...")

	// Add the Longhorn Helm repository
	if _, err := ssh.Run("helm repo add longhorn https://charts.longhorn.io", host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to add Longhorn helm repo: %w", err)
	}

	// Update Helm repositories
	if _, err := ssh.Run("helm repo update", host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}

	// Create namespace for Longhorn
	out, err := ssh.Run("kubectl create namespace longhorn-system", host, keyPath, user, "", true, 0)
	if err != nil && !strings.Contains(out, "already exists") {
		return fmt.Errorf("failed to create longhorn namespace: %s: %w", out, err)
	}

	// Install Longhorn using Helm
	if _, err := ssh.Run("helm install longhorn longhorn/longhorn --namespace longhorn-system --set persistence.defaultClass=true", host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to install Longhorn: %w", err)
	}

	// Wait for Longhorn to be ready
	if _, err := ssh.Run("kubectl wait --for=condition=ready --timeout=300s pod -l app=longhorn-manager -n longhorn-system", host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to wait for Longhorn to be ready: %w", err)
	}

	log.Println("Longhorn installation completed successfully")
	return nil
}
