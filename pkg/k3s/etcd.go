package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallEtcd installs etcd on the cluster
func InstallEtcd(host, user, key string) error {
	// Add etcd helm repo
	cmd := "sudo apt install -y etcd-server etcd-client"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("error adding etcd helm repo: %w", err)
	} else {
		log.Println(out)
	}

	return nil
}
