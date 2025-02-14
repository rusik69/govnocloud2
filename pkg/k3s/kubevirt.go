package k3s

import (
	"fmt"
	"log"
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

	// Wait for KubeVirt operator to be ready
	time.Sleep(5 * time.Second)
	waitOperatorCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-operator -n kubevirt"
	log.Println(waitOperatorCmd)
	if _, err := ssh.Run(waitOperatorCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for virt-operator: %w", err)
	}

	// Wait for KubeVirt CR to be ready
	waitKVCmd := "kubectl wait --for=condition=Available --timeout=300s kubevirt kubevirt -n kubevirt"
	log.Println(waitKVCmd)
	if _, err := ssh.Run(waitKVCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt CR: %w", err)
	}

	// Additional wait to ensure all components are created
	time.Sleep(30 * time.Second)

	// Wait for virt-api
	waitApiCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-api -n kubevirt"
	log.Println(waitApiCmd)
	if _, err := ssh.Run(waitApiCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for virt-api: %w", err)
	}

	// Wait for virt-controller
	waitControllerCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-controller -n kubevirt"
	log.Println(waitControllerCmd)
	if _, err := ssh.Run(waitControllerCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for virt-controller: %w", err)
	}

	// install kubevirt manager
	if err := InstallKubeVirtManager(host, user, key); err != nil {
		return fmt.Errorf("failed to install KubeVirt Manager: %w", err)
	}

	// Wait for kubevirt-manager deployment
	waitManagerCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app=kubevirt-manager -n kubevirt-manager"
	log.Println(waitManagerCmd)
	if _, err := ssh.Run(waitManagerCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for kubevirt-manager: %w", err)
	}

	// create ingress
	if err := CreateKubevirtManagerIngress(host, user, key, managerHost); err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	// remove default virtualmachineinstancetypes
	if err := RemoveDefaultVirtualMachineInstanceTypes(host, user, key); err != nil {
		return fmt.Errorf("failed to remove default virtualmachineinstancetypes: %w", err)
	}

	// create virtualmachineinstancetypes
	if err := CreateVirtualMachineInstanceTypes(host, user, key); err != nil {
		return fmt.Errorf("failed to create virtualmachineinstancetypes: %w", err)
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
	return nil
}

func CreateVirtualMachineInstanceTypes(host, user, key string) error {
	// create virtualmachineinstancetypes based on vmsizes
	for _, vmSize := range types.VMSizes {
		cmd := fmt.Sprintf("virtctl create instancetype --name %s --cpu %d --memory %d | kubectl apply -f -", vmSize.Name, vmSize.CPU, vmSize.RAM)
		log.Println(cmd)
		out, err := ssh.Run(cmd, host, key, user, "", true, 60)
		if err != nil {
			return fmt.Errorf("failed to create virtualmachineinstancetype: %w", err)
		}
		log.Println(out)
	}
	log.Println("KubeVirt Manager is successfully installed")
	return nil
}

func RemoveDefaultVirtualMachineInstanceTypes(host, user, key string) error {
	// remove default virtualmachineinstancetypes
	cmd := "kubectl delete virtualmachineclusterinstancetype --all"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, key, user, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to remove default virtualmachineinstancetypes: %w", err)
	}
	log.Println(out)
	return nil
}
