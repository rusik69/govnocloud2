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

	// Install open-iscsi and other required packages on all nodes
	cmd = "kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type==\"InternalIP\")].address}'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, keyPath, user, "", true, 0)
	if err != nil {
		return fmt.Errorf("failed to get node IPs: %w", err)
	}

	nodeIPs := strings.Fields(out)
	for _, nodeIP := range nodeIPs {
		// Install required packages
		cmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s "+
			"'sudo apt-get update && "+
			"sudo apt-get install -y open-iscsi nfs-common util-linux apache2-utils'",
			keyPath, user, nodeIP)
		log.Printf("Installing required packages on node %s", nodeIP)
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
			return fmt.Errorf("failed to install packages on node %s: %w", nodeIP, err)
		}

		// Create data directory
		cmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s "+
			"'sudo mkdir -p /var/lib/longhorn && "+
			"sudo chmod 777 /var/lib/longhorn'",
			keyPath, user, nodeIP)
		log.Printf("Creating data directory on node %s", nodeIP)
		if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
			return fmt.Errorf("failed to create data directory on node %s: %w", nodeIP, err)
		}
	}

	// Install Longhorn using Helm with specific configuration
	cmd = "helm install longhorn longhorn/longhorn " +
		"--namespace longhorn-system " +
		"--set persistence.defaultClass=true " +
		"--set defaultSettings.defaultReplicaCount=1 " +
		"--set defaultSettings.createDefaultDiskLabeledNodes=true " +
		"--set defaultSettings.defaultDataPath=/var/lib/longhorn " +
		"--set defaultSettings.defaultDataLocality=disabled " +
		"--set defaultSettings.replicaSoftAntiAffinity=true " +
		"--set defaultSettings.storageOverProvisioningPercentage=200 " +
		"--set defaultSettings.storageMinimalAvailablePercentage=10 " +
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

	// Install Longhorn dashboard
	log.Println("Installing Longhorn dashboard...")

	// Create ingress for Longhorn dashboard
	ingressYaml := `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: longhorn-ingress
  namespace: longhorn-system
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
spec:
  rules:
  - host: master.govno.cloud
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: longhorn-frontend
            port:
              number: 80
`

	// Write ingress YAML to a temporary file on the master node
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/longhorn-ingress.yaml\n%s\nEOF", ingressYaml)
	log.Println("Creating Longhorn ingress YAML")
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to create ingress YAML: %w", err)
	}

	// Apply the ingress configuration
	cmd = "kubectl apply -f /tmp/longhorn-ingress.yaml"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to apply Longhorn ingress: %s: %w", out, err)
	}
	// Create basic auth secret for dashboard access
	cmd = "htpasswd -bc /tmp/auth admin govnocloud"
	log.Println("Creating basic auth file")
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to create basic auth file: %w", err)
	}

	// Create Kubernetes secret from basic auth file
	cmd = "kubectl create secret generic basic-auth --from-file=/tmp/auth -n longhorn-system"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, keyPath, user, "", true, 0)
	if err != nil && !strings.Contains(out, "already exists") {
		return fmt.Errorf("failed to create basic auth secret: %s: %w", out, err)
	}

	// Add basic auth annotation to ingress
	cmd = "kubectl annotate ingress longhorn-ingress -n longhorn-system nginx.ingress.kubernetes.io/auth-type=basic nginx.ingress.kubernetes.io/auth-secret=basic-auth --overwrite"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		return fmt.Errorf("failed to add basic auth to ingress: %w", err)
	}

	// Wait for ingress to be ready
	log.Println("Waiting for Longhorn dashboard to be accessible...")
	maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		cmd = "kubectl get ingress -n longhorn-system longhorn-ingress"
		out, err := ssh.Run(cmd, host, keyPath, user, "", true, 0)
		if err == nil && strings.Contains(out, "master.govno.cloud") {
			log.Println("Longhorn dashboard is accessible at: http://master.govno.cloud")
			log.Println("Username: admin")
			log.Println("Password: govnocloud")
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to verify Longhorn dashboard access after %d attempts", maxRetries)
		}
		log.Printf("Waiting for ingress to be ready (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	// Cleanup temporary files
	cmd = "rm -f /tmp/longhorn-ingress.yaml /tmp/auth"
	if _, err := ssh.Run(cmd, host, keyPath, user, "", true, 0); err != nil {
		log.Printf("Warning: failed to cleanup temporary files: %v", err)
	}

	log.Println("Longhorn installation completed successfully")
	return nil
}
