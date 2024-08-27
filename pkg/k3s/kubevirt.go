package k3s

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// KubeVirtVersion is the version of KubeVirt.
const KubeVirtVersion = "v1.3.1"

// InstallKubeVirt installs KubeVirt to k3s cluster.
func InstallKubeVirt() error {
	operator := "https://github.com/kubevirt/kubevirt/releases/download/" + KubeVirtVersion + "/kubevirt-operator.yaml"
	command := exec.Command("kubectl", "apply", "-f", operator)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing KubeVirt operator: %w", err)
	}

	cr := "https://github.com/kubevirt/kubevirt/releases/download/" + KubeVirtVersion + "/kubevirt-cr.yaml"
	command = exec.Command("kubectl", "apply", "-f", cr)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing KubeVirt CR: %w", err)
	}

	log.Println("waiting for KubeVirt to be ready")
	command = exec.Command("kubectl", "-n", "kubevirt", "wait", "kv", "kubevirt", "--for", "condition=Available")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error waiting for KubeVirt to be ready: %w", err)
	}

	return nil
}
