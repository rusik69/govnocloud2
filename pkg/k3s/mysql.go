package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallMySQL installs MySQL operator using Helm
func InstallMySQL(host, user, key string) error {
	// Add MySQL operator Helm repository
	addRepoCmd := "helm repo add mysql-operator https://mysql.github.io/mysql-operator/ && helm repo update"
	log.Println(addRepoCmd)
	if out, err := ssh.Run(addRepoCmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to add MySQL operator helm repo: %w", err)
	} else {
		log.Println(out)
	}

	// Install MySQL operator
	installCmd := "helm install mysql-operator mysql-operator/mysql-operator --namespace mysql-operator --create-namespace --wait"
	log.Println(installCmd)
	if out, err := ssh.Run(installCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to install MySQL operator: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for operator to be ready
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=mysql-operator -n mysql-operator"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for MySQL operator: %w", err)
	}

	return nil
}
