package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

func InstallClickhouse(host, user, key string) error {
	// Add Altinity Clickhouse Operator Helm repository
	addRepoCmd := "helm repo add altinity https://charts.altinity.com && helm repo update"
	log.Println(addRepoCmd)
	if out, err := ssh.Run(addRepoCmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to add Altinity helm repo: %w", err)
	} else {
		log.Println(out)
	}

	// Install Altinity Clickhouse Operator
	installCmd := "helm install clickhouse-operator altinity/clickhouse-operator --namespace clickhouse-system --create-namespace --wait"
	log.Println(installCmd)
	if out, err := ssh.Run(installCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to install Clickhouse operator: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for operator to be ready
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app=clickhouse-operator -n clickhouse-system"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for Clickhouse operator: %w", err)
	}

	return nil
}
