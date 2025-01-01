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
		"--set defaultSettings.defaultReplicaCount=1 " +
		"--set defaultSettings.createDefaultDiskLabeledNodes=true " +
		"--set defaultSettings.defaultDataPath=/dev/sda " +
		"--set csi.attacherReplicaCount=1 " +
		"--set csi.provisionerReplicaCount=1 " +
		"--set csi.resizerReplicaCount=1 " +
		"--set csi.snapshotterReplicaCount=1 " +
		"--set csi.kubeletRootDir=/var/lib/kubelet " +
		"--set longhornManager.priorityClass=\"\" " +
		"--set longhornDriver.priorityClass=\"\""
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to install Longhorn: %w", err)
	}

	// Wait for initial pod creation
	log.Println("Waiting for initial pod creation...")
	time.Sleep(30 * time.Second)

	// Check Longhorn pod status
	maxRetries := 20 // Increased retries
	for i := 0; i < maxRetries; i++ {
		log.Printf("Checking Longhorn pods (attempt %d/%d)...", i+1, maxRetries)
		cmd = "kubectl -n longhorn-system get pods"
		out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0)
		if err == nil {
			log.Printf("Pod status:\n%s", out)
		}

		// Check if all pods are ready
		cmd = "kubectl -n longhorn-system wait --for=condition=ready pod --all --timeout=60s"
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err == nil {
			log.Println("All Longhorn pods are ready")
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to wait for Longhorn pods after %d attempts", maxRetries)
		}
		time.Sleep(30 * time.Second)
	}

	// Label nodes for Longhorn storage
	for _, nodeIP := range nodeIPs {
		cmd = fmt.Sprintf("kubectl label nodes node-%s node.longhorn.io/create-default-disk=true --overwrite", nodeIP)
		log.Printf("Labeling node %s for Longhorn storage", nodeIP)
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
			return fmt.Errorf("failed to label node %s: %w", nodeIP, err)
		}
	}

	// Check CSI driver registration with debugging
	log.Println("Checking CSI driver registration...")
	for i := 0; i < maxRetries; i++ {
		// List all CSI drivers for debugging
		cmd = "kubectl get csidrivers"
		out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0)
		if err != nil {
			return fmt.Errorf("failed to get CSI drivers: %w", err)
		}
		log.Printf("Available CSI drivers:\n%s", out)

		// Check for Longhorn CSI driver
		cmd = "kubectl get csidriver driver.longhorn.io"
		out, err = ssh.Run(cmd, host, keyPath, user, "", true, 0)
		if err == nil {
			log.Printf("CSI driver found:\n%s", out)
			log.Println("CSI driver successfully registered")
			break
		}
		if i == maxRetries-1 {
			// Get Longhorn manager logs for debugging
			cmd = "kubectl -n longhorn-system logs -l app=longhorn-manager --tail=100"
			out, _ := ssh.Run(cmd, host, keyPath, user, "", true, 0)
			log.Printf("Longhorn manager logs:\n%s", out)
			return fmt.Errorf("CSI driver not registered after %d attempts", maxRetries)
		}
		log.Printf("CSI driver not ready yet, waiting... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(30 * time.Second)
	}

	log.Println("Longhorn installation completed successfully")
	return nil
}
