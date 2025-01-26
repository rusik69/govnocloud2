package k3s

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rusik69/govnocloud2/pkg/ssh"
	"github.com/rusik69/govnocloud2/pkg/types"
)

func InstallKubeVirt(host, user, key, managerHost, version string) error {
	baseURL := fmt.Sprintf("https://github.com/kubevirt/kubevirt/releases/download/%s", version)

	// Install operator and CR
	manifests := []string{"kubevirt-operator.yaml", "kubevirt-cr.yaml"}
	for _, manifest := range manifests {
		cmd := fmt.Sprintf("kubectl apply -f %s/%s --wait=true --timeout=300s", baseURL, manifest)
		log.Println(cmd)
		if out, err := ssh.Run(cmd, host, key, user, "", true, 60); err != nil {
			return fmt.Errorf("failed to apply %s: %w", manifest, err)
		} else {
			log.Println(out)
		}
	}

	// Install virtctl
	virtctlCmd := fmt.Sprintf("sudo curl --no-progress-meter -L -o /usr/local/bin/virtctl %s/virtctl-%s-linux-amd64 && sudo chmod +x /usr/local/bin/virtctl",
		baseURL, version)
	log.Println(virtctlCmd)
	if out, err := ssh.Run(virtctlCmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to install virtctl: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for KubeVirt to be ready
	time.Sleep(5 * time.Second)
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-operator -n kubevirt"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt: %w", err)
	}

	// install kubevirt manager
	if err := InstallKubeVirtManager(host, user, key); err != nil {
		return fmt.Errorf("failed to install KubeVirt Manager: %w", err)
	}

	// create ingress
	if err := CreateKubevirtManagerIngress(host, user, key, managerHost); err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	return nil
}

func InstallKubeVirtManager(host, user, key string) error {
	managerURL := "https://raw.githubusercontent.com/kubevirt-manager/kubevirt-manager/main/kubernetes/bundled.yaml"

	// Install manager
	cmd := fmt.Sprintf("kubectl apply -f %s --wait=true --timeout=300s", managerURL)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to install KubeVirt Manager: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for manager to be ready
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app=kubevirt-manager -n kubevirt-manager"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt Manager: %w", err)
	}

	return nil
}

// CreateKubevirtManagerIngress creates an ingress for kubevirt manager
func CreateKubevirtManagerIngress(host, user, key, managerHost string) error {
	ingressYaml := fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubevirt-manager-ingress
  namespace: kubevirt-manager
spec:
  ingressClassName: traefik
  rules:
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: kubevirt-manager
            port:
              number: 8080
`, managerHost)
	cmd := fmt.Sprintf("cat << 'EOF' > /tmp/kubevirt-manager-ingress.yaml\n%s\nEOF", ingressYaml)
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, key, user, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to create ingress YAML: %w", err)
	}
	log.Println(out)
	cmd = "kubectl apply -f /tmp/kubevirt-manager-ingress.yaml -n kubevirt-manager"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to apply ingress: %w", err)
	}
	log.Println(out)
	// wait for kubevirt instance types to be ready
	cmd = "kubectl wait --for=condition=ready --timeout=300s pod -l app=kubevirt-manager -n kubevirt-manager"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, "", true, 300)
	if err != nil {
		return fmt.Errorf("failed to wait for kubevirt instance types: %w %s", err, out)
	}
	log.Println(out)
	// get kubevirt instance types
	cmd = "kubectl get virtualmachineclusterinstancetypes -o jsonpath='{.items[*].metadata.name}'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to get kubevirt instance types: %w", err)
	}
	log.Println(out)
	instanceTypes := strings.Split(out, " ")
	// remove kubevirt instance types
	for _, instanceType := range instanceTypes {
		if instanceType != "" {
			cmd = fmt.Sprintf("kubectl delete virtualmachineclusterinstancetype %s", instanceType)
			log.Println(cmd)
			out, err = ssh.Run(cmd, host, key, user, "", true, 60)
			if err != nil {
				return fmt.Errorf("failed to delete kubevirt instance type: %w", err)
			}
			log.Println(out)
		}
	}
	// create virtualmachineinstancetypes based on vmsizes
	for _, vmSize := range types.VMSizes {
		cmd = fmt.Sprintf("virtctl create instancetype --name %s --cpu %d --memory %d | kubectl apply -f -", vmSize.Name, vmSize.CPU, vmSize.RAM)
		log.Println(cmd)
		out, err = ssh.Run(cmd, host, key, user, "", true, 60)
		if err != nil {
			return fmt.Errorf("failed to create virtualmachineinstancetype: %w", err)
		}
		log.Println(out)
	}
	log.Println("KubeVirt Manager is successfully installed")
	return nil
}
