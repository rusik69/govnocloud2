package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

func InstallClickhouse(host, user, key string) error {
	// Install Altinity Clickhouse Operator
	installCmd := "kubectl apply -f https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle.yaml --wait"
	log.Println(installCmd)
	if out, err := ssh.Run(installCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to install Clickhouse operator: %w", err)
	} else {
		log.Println(out)
	}
	return nil
}
