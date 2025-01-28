package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

func InstallCNPG(host, user, key string) error {
	// Add CloudNativePG Helm repository
	addRepoCmd := "helm repo add cnpg https://cloudnative-pg.github.io/charts && helm repo update"
	log.Println(addRepoCmd)
	if out, err := ssh.Run(addRepoCmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to add CNPG helm repo: %w", err)
	} else {
		log.Println(out)
	}

	// Install CloudNativePG operator
	installCmd := "helm install cnpg cnpg/cloudnative-pg --namespace cnpg-system --create-namespace --wait"
	log.Println(installCmd)
	if out, err := ssh.Run(installCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to install CNPG operator: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for operator to be ready
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=cloudnative-pg -n cnpg-system"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for CNPG operator: %w", err)
	}

	return nil
}
