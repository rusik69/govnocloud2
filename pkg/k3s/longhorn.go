package k3s

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallLonghorn installs Longhorn storage system into the Kubernetes cluster
func InstallLonghorn(master string, nodeIPs []string, user, keyPath, ingressHost, disk string) error {
	log.Println("Installing Longhorn storage system...")

	// Add the Longhorn Helm repository
	cmd := "helm repo add longhorn https://charts.longhorn.io"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to add Longhorn helm repo: %w", err)
	}

	// Update Helm repositories
	cmd = "helm repo update"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}

	// Create namespace for Longhorn
	cmd = "kubectl create namespace longhorn-system"
	log.Println(cmd)
	out, err := ssh.Run(cmd, master, keyPath, user, "", true, 0)
	if err != nil && !strings.Contains(out, "already exists") {
		return fmt.Errorf("failed to create longhorn namespace: %s: %w", out, err)
	}

	cmd = "sudo apt-get update ; " +
		"sudo apt-get install -y open-iscsi nfs-common util-linux apache2-utils ; " +
		"sudo modprobe dm_crypt ; " +
		"sudo systemctl disable --now multipathd.socket ; " +
		"sudo systemctl disable --now multipathd.service ; " +
		"sudo systemctl enable --now iscsid ; " +
		"sudo umount /mnt ; sudo umount /var/lib/longhorn ; " +
		"sudo dd if=/dev/zero of=/dev/" + disk + " bs=1M count=100 ; " +
		"sudo blockdev --rereadpt /dev/" + disk + " ; " +
		"sudo mkfs.ext4 /dev/" + disk + " ; " +
		"sudo mkdir -p /var/lib/longhorn ; " +
		"sudo mount /dev/" + disk + " /var/lib/longhorn"

	for _, nodeIP := range nodeIPs {
		// Install required packages
		cmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s "+
			"'"+cmd+"'",
			keyPath, user, nodeIP)
		log.Println(cmd)
		if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
			return fmt.Errorf("failed to prepare node %s: %w", nodeIP, err)
		}
	}

	// Get list of nodes
	cmd = "kubectl get nodes -o jsonpath='{.items[*].metadata.name}'"
	out, err = ssh.Run(cmd, master, keyPath, user, "", true, 0)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	nodes := strings.Fields(out)

	log.Printf("Nodes: %+v", nodes)

	// Install Longhorn using Helm with filesystem configuration
	cmd = "helm install longhorn longhorn/longhorn " +
		"--namespace longhorn-system " +
		"--set persistence.defaultClass=true " +
		"--set defaultSettings.defaultReplicaCount=1 " +
		"--set defaultSettings.defaultDataLocality=disabled " +
		"--set defaultSettings.replicaSoftAntiAffinity=true " +
		"--set defaultSettings.storageOverProvisioningPercentage=200 " +
		"--set defaultSettings.storageMinimalAvailablePercentage=10 " +
		"--set defaultSettings.storageReservedPercentageForDefaultDisk=1 " +
		"--set defaultSettings.disableSchedulingOnCordonedNode=true " +
		"--set defaultSettings.nodeDownPodDeletionPolicy=delete-both-statefulset-and-deployment-pod " +
		"--set defaultSettings.allowNodeDrainWithLastHealthyReplica=true " +
		"--set defaultSettings.autoCleanupSystemGeneratedSnapshot=true " +
		"--set defaultSettings.concurrentAutomaticEngineUpgrade=3 " +
		"--set defaultSettings.backingImageCleanupWaitInterval=600 " +
		"--set csi.attacherReplicaCount=1 " +
		"--set csi.provisionerReplicaCount=1 " +
		"--set csi.resizerReplicaCount=1 " +
		"--set csi.snapshotterReplicaCount=1 " +
		"--set csi.kubeletRootDir=/var/lib/kubelet " +
		"--set longhornManager.priorityClass=\"\" " +
		"--set longhornDriver.priorityClass=\"\""

	log.Println(cmd)
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to install Longhorn: %w", err)
	}

	// Wait for initial pod creation
	log.Println("Waiting for initial pod creation...")
	time.Sleep(30 * time.Second)

	// Check CSI driver registration with debugging
	maxRetries := 20
	log.Println("Checking CSI driver registration...")
	for i := 0; i < maxRetries; i++ {
		// List all CSI drivers for debugging
		cmd = "kubectl get csidrivers"
		log.Println(cmd)
		out, err := ssh.Run(cmd, master, keyPath, user, "", true, 0)
		if err != nil {
			return fmt.Errorf("failed to get CSI drivers: %w", err)
		}
		log.Printf("Available CSI drivers:\n%s", out)

		// Check for Longhorn CSI driver
		cmd = "kubectl get csidriver driver.longhorn.io"
		log.Println(cmd)
		out, err = ssh.Run(cmd, master, keyPath, user, "", true, 0)
		if err == nil {
			log.Printf("CSI driver found:\n%s", out)
			log.Println("CSI driver successfully registered")
			break
		}
		if i == maxRetries-1 {
			// Get Longhorn manager logs for debugging
			cmd = "kubectl -n longhorn-system logs -l app=longhorn-manager --tail=100"
			log.Println(cmd)
			out, _ := ssh.Run(cmd, master, keyPath, user, "", true, 0)
			log.Printf("Longhorn manager logs:\n%s", out)
			return fmt.Errorf("CSI driver not registered after %d attempts", maxRetries)
		}
		log.Printf("CSI driver not ready yet, waiting... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(30 * time.Second)
	}

	// Install Longhorn dashboard
	log.Println("Installing Longhorn dashboard...")

	// Create ingress for Longhorn dashboard
	ingressYaml := fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: longhorn-ingress
  namespace: longhorn-system
spec:
  rules:
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: longhorn-frontend
            port:
              number: 80
`, ingressHost)

	// Write ingress YAML to a temporary file on the master node
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/longhorn-ingress.yaml\n%s\nEOF", ingressYaml)
	log.Println("Creating Longhorn ingress YAML")
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to create ingress YAML: %w", err)
	}

	// Apply the ingress configuration
	cmd = "kubectl apply -f /tmp/longhorn-ingress.yaml"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to apply Longhorn ingress: %s: %w", out, err)
	}

	// Cleanup temporary files
	cmd = "rm -f /tmp/longhorn-ingress.yaml /tmp/auth"
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		log.Printf("Warning: failed to cleanup temporary files: %v", err)
	}

	// wait for all pods in longhorn-system namespace to be ready
	log.Println("Waiting for all pods in longhorn-system namespace to be ready...")
	cmd = "kubectl wait --for=condition=ready pod --all -n longhorn-system --timeout=1800s"
	if _, err := ssh.Run(cmd, master, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to wait for pods to be ready: %w", err)
	}

	log.Println("Longhorn installation completed successfully")
	return nil
}
