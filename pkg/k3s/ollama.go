package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

func InstallOllama(host, user, key string) error {
	installCmd := "kubectl apply --server-side=true -f https://raw.githubusercontent.com/nekomeowww/ollama-operator/v0.10.1/dist/install.yaml"
	if out, err := ssh.Run(installCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to install ollama: %w", err)
	} else {
		log.Println(out)
	}
	waitCmd := "kubectl wait -n ollama-operator-system --for=jsonpath='{.status.readyReplicas}'=1 deployment/ollama-operator-controller-manager"
	if out, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for ollama: %w", err)
	} else {
		log.Println(out)
	}
	return nil
}
