package k3s

import (
	"fmt"
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
	return nil
}
