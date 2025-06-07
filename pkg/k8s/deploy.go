package k8s

import (
	"fmt"
	"log"
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

type GovnocloudServiceConfig struct {
	Name        string
	Description string
	ExecStart   string
	User        string
}

// createSystemdService creates a systemd service file
func createSystemdService(config GovnocloudServiceConfig) (string, error) {
	log.Println("Generating service file")
	serviceBody := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
ExecStart=%s
Restart=on-failure
User=%s
[Install]
WantedBy=multi-user.target
`, config.Description, config.ExecStart, config.User)
	log.Println(serviceBody)
	file, err := os.CreateTemp("", config.Name+".service")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(serviceBody); err != nil {
		return "", fmt.Errorf("failed to write service file: %w", err)
	}

	return file.Name(), nil
}

// Deploy deploys the server.
func Deploy(host, serverHost, webHost, serverPort, webPort, user, password, key, webPath string) error {
	const (
		binaryPath = "bin/govnocloud2-linux-amd64"
		destPath   = "/usr/local/bin/govnocloud2"
	)

	log.Printf("Copying govnocloud2 to %s", host)
	if err := ssh.Copy(binaryPath, destPath, host, "root", key); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Sync web files
	if err := ssh.Rsync("pkg/web/templates/*", webPath+"/templates", host, "root", key); err != nil {
		return fmt.Errorf("failed to sync web files: %w", err)
	}

	// Make binary executable
	cmd := fmt.Sprintf("sudo chmod +x %s", destPath)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, password, false, 5); err != nil {
		return fmt.Errorf("failed to make binary executable: %s", out)
	}

	// Create and deploy server service
	serverConfig := GovnocloudServiceConfig{
		Name:        "govnocloud2",
		Description: "govnocloud2 server",
		ExecStart:   fmt.Sprintf("%s server --port %s --host %s", destPath, serverPort, serverHost),
		User:        "root",
	}

	serverServicePath, err := createSystemdService(serverConfig)
	if err != nil {
		return fmt.Errorf("failed to create server service: %w", err)
	}
	defer os.Remove(serverServicePath)

	log.Printf("Copying govnocloud2 service from %s to %s", serverServicePath, host)
	if err := ssh.Copy(serverServicePath, "/etc/systemd/system/govnocloud2.service", host, "root", key); err != nil {
		return fmt.Errorf("failed to copy server service: %w", err)
	}

	// Create and deploy web service
	webConfig := GovnocloudServiceConfig{
		Name:        "govnocloud2-web",
		Description: "govnocloud2 web",
		ExecStart:   fmt.Sprintf("%s web --port %s --host %s --webpath %s", destPath, webPort, webHost, webPath),
		User:        "root",
	}

	webServicePath, err := createSystemdService(webConfig)
	if err != nil {
		return fmt.Errorf("failed to create web service: %w", err)
	}
	defer os.Remove(webServicePath)

	log.Printf("Copying govnocloud2-web service from %s to %s", webServicePath, host)
	if err := ssh.Copy(webServicePath, "/etc/systemd/system/govnocloud2-web.service", host, "root", key); err != nil {
		return fmt.Errorf("failed to copy web service: %w", err)
	}

	// Install etcd
	cmd = "sudo apt install -y etcd-server"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, password, false, 30); err != nil {
		return fmt.Errorf("failed to install etcd: %s", out)
	}

	// Reload systemd and enable services
	commands := []struct {
		cmd  string
		desc string
	}{
		{"sudo systemctl daemon-reload", "reload systemd"},
		{"sudo systemctl enable --now govnocloud2", "enable server service"},
		{"sudo systemctl enable --now govnocloud2-web", "enable web service"},
		{"sudo systemctl enable --now etcd", "enable etcd service"},
	}

	for _, command := range commands {
		log.Printf("Running: %s", command.cmd)
		if out, err := ssh.Run(command.cmd, host, key, user, password, false, 5); err != nil {
			return fmt.Errorf("failed to %s: %s", command.desc, out)
		}
	}

	// Install K9s after K3s is deployed
	if err := InstallK9s(host, user, key); err != nil {
		return fmt.Errorf("failed to install k9s: %w", err)
	}

	return nil
}

// Wol wakes on lan.
func Wol(master, user, key, ip string, macs []string) {
	for _, mac := range macs {
		cmd := fmt.Sprintf("wakeonlan -i %s %s", ip, mac)
		log.Println(cmd)

		if out, err := ssh.Run(cmd, master, key, user, "", false, 5); err != nil {
			log.Printf("Failed to wake %s: %v\nOutput: %s", mac, err, out)
			continue
		}
	}
}

// Suspend suspends the servers
func Suspend(ips []string, master, user, password, key string) {
	for _, ip := range ips {
		log.Printf("Suspending server: %s", ip)
		cmd := fmt.Sprintf("ssh %s@%s 'sudo systemctl suspend'", user, ip)
		log.Println(cmd)

		if out, err := ssh.Run(cmd, master, key, user, password, false, 5); err != nil {
			log.Printf("Failed to suspend %s: %v\nOutput: %s", ip, err, out)
			continue
		}
	}
}

// InstallK9s installs the K9s terminal UI for Kubernetes
func InstallK9s(host, user, key string) error {
	// Download K9s binary
	version := "v0.32.7"
	cmd := fmt.Sprintf("curl -LO https://github.com/derailed/k9s/releases/download/%s/k9s_Linux_amd64.tar.gz && "+
		"tar xzf k9s_Linux_amd64.tar.gz && "+
		"sudo mv k9s /usr/local/bin/ && "+
		"rm k9s_Linux_amd64.tar.gz LICENSE README.md", version)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to install k9s: %v\nOutput: %s", err, out)
	}

	log.Printf("K9s installed successfully on %s", host)
	return nil
}

// SetupNat setups NAT
func SetupNat(host, user, key, externalInterface, internalInterface string) error {
	// Install iptables-persistent
	cmd := "sudo apt install -y iptables-persistent"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to install iptables-persistent: %v\nOutput: %s", err, out)
	}
	// Enable ip_forward
	cmd = "sudo sysctl -w net.ipv4.ip_forward=1"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to enable ip_forward: %v\nOutput: %s", err, out)
	}
	// check if masquerade rule exists
	cmd = fmt.Sprintf("sudo iptables -t nat -C POSTROUTING -o %s -j MASQUERADE 2>/dev/null || echo 'rule_not_found'", externalInterface)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to check masquerade rule: %v\nOutput: %s", err, out)
	} else if out == "rule_not_found" {
		log.Println("Masquerade rule not found, adding it")
	} else {
		log.Println("Masquerade rule already exists")
		return nil
	}
	// add masquerade rule
	cmd = fmt.Sprintf("sudo iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", externalInterface)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to add masquerade rule: %v\nOutput: %s", err, out)
	}
	// check if forward rule exists
	cmd = fmt.Sprintf("sudo iptables -C FORWARD -i %s -o %s -j ACCEPT 2>/dev/null || echo 'rule_not_found'", internalInterface, externalInterface)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to check forward rule: %v\nOutput: %s", err, out)
	} else if out == "rule_not_found" {
		log.Println("Forward rule not found, adding it")
	} else {
		log.Println("Forward rule already exists")
		return nil
	}
	// check if established rule exists
	cmd = "sudo iptables -C FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || echo 'rule_not_found'"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to check established rule: %v\nOutput: %s", err, out)
	} else if out == "rule_not_found" {
		log.Println("Established rule not found, adding it")
	} else {
		log.Println("Established rule already exists")
		return nil
	}
	// add forwarding rule
	cmd = fmt.Sprintf("sudo iptables -A FORWARD -i %s -o %s -j ACCEPT", internalInterface, externalInterface)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to add forwarding rule: %v\nOutput: %s", err, out)
	}
	// add established rule
	cmd = "sudo iptables -A FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to add established rule: %v\nOutput: %s", err, out)
	}
	// save iptables rules
	cmd = "sudo iptables-save | sudo tee /etc/iptables/rules.v4"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return fmt.Errorf("failed to save iptables rules: %v\nOutput: %s", err, out)
	}
	return nil
}
