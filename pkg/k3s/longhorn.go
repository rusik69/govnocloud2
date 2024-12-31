package k3s

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallLonghorn installs Longhorn storage system into the Kubernetes cluster
func InstallLonghorn(host, user, keyPath string) error {
	log.Println("Installing Longhorn storage system...")

	// Add the Longhorn Helm repository
	cmd := "helm repo add longhorn https://charts.longhorn.io"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to add Longhorn helm repo: %w", err)
	}

	// Update Helm repositories
	cmd = "helm repo update"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}

	// Create namespace for Longhorn
	cmd = "kubectl create namespace longhorn-system"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0)
	if err != nil && !strings.Contains(out, "already exists") {
		return fmt.Errorf("failed to create longhorn namespace: %s: %w", out, err)
	}

	// Install open-iscsi on all nodes
	cmd = "kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type==\"InternalIP\")].address}'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, keyPath, user, "", true, 0)
	if err != nil {
		return fmt.Errorf("failed to get node IPs: %w", err)
	}

	nodeIPs := strings.Fields(out)
	for _, nodeIP := range nodeIPs {
		cmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s 'sudo apt-get update && sudo apt-get install -y open-iscsi'",
			keyPath, user, nodeIP)
		log.Printf("Installing open-iscsi on node %s", nodeIP)
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
			return fmt.Errorf("failed to install open-iscsi on node %s: %w", nodeIP, err)
		}
	}

	// Install Longhorn using Helm with specific configuration
	cmd = "helm install longhorn longhorn/longhorn " +
		"--namespace longhorn-system " +
		"--set persistence.defaultClass=true " +
		"--set csi.attacherReplicaCount=1 " +
		"--set csi.provisionerReplicaCount=1 " +
		"--set csi.resizerReplicaCount=1 " +
		"--set csi.snapshotterReplicaCount=1"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to install Longhorn: %w", err)
	}

	// Wait for Longhorn pods to be ready with retries
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		log.Printf("Waiting for Longhorn pods (attempt %d/%d)...", i+1, maxRetries)
		cmd = "kubectl -n longhorn-system wait --for=condition=ready pod --all --timeout=60s"
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err == nil {
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to wait for Longhorn pods after %d attempts", maxRetries)
		}
		time.Sleep(30 * time.Second)
	}

	// Wait for CSI driver to be registered with retries
	log.Println("Waiting for CSI driver registration...")
	for i := 0; i < maxRetries; i++ {
		cmd = "kubectl get csidriver driver.longhorn.io -o name"
		if out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err == nil && strings.Contains(out, "driver.longhorn.io") {
			log.Println("CSI driver successfully registered")
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("CSI driver not registered after %d attempts", maxRetries)
		}
		log.Printf("CSI driver not ready yet, waiting... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(30 * time.Second)
	}

	log.Println("Longhorn installation completed successfully")
	return nil
}
