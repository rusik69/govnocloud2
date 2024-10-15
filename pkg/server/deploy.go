package server

import (
	"os"
	"os/exec"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// Deploy deploys the server.
func Deploy(host, port, user, key string) error {
	err := ssh.Copy("bin/govnocloud2-linux-amd64", "/usr/local/bin/govnocloud2", host, user, key)
	if err != nil {
		return err
	}
	cmd := "sudo chmod +x /usr/local/bin/govnocloud2"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	serviceBody := `[Unit]
Description=govnocloud2 server
After=network.target

[Service]
ExecStart=/usr/local/bin/govnocloud2 server --port ` + port + ` --host ` + host + `
Restart=on-failure
User=root
[Install]
WantedBy=multi-user.target
`
	tempFile, err := os.CreateTemp("", "govnocloud2.service")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	_, err = tempFile.WriteString(serviceBody)
	if err != nil {
		return err
	}
	err = ssh.Copy(tempFile.Name(), "/etc/systemd/system/govnocloud2.service", host, user, key)
	if err != nil {
		return err
	}
	serviceWebBody := `[Unit]
Description=govnocloud2 web
After=network.target

[Service]
ExecStart=/usr/local/bin/govnocloud2 web --port 8080 --host ` + host + `
Restart=on-failure
User=root
[Install]
WantedBy=multi-user.target
`
	tempFile2, err := os.CreateTemp("", "govnocloud2-web.service")
	if err != nil {
		return err
	}
	defer tempFile2.Close()
	_, err = tempFile2.WriteString(serviceWebBody)
	if err != nil {
		return err
	}
	err = ssh.Copy(tempFile2.Name(), "/etc/systemd/system/govnocloud2-web.service", host, user, key)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl daemon-reload"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl enable govnocloud2"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl enable govnocloud2-web"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl start govnocloud2"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl start govnocloud2-web"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	return nil
}

// Wol wakes on lan.
func Wol(ip string, macs []string) error {
	for _, mac := range macs {
		cmd := []string{"wakeonlan", "-i", ip, mac}
		err := exec.Command("wakeonlan", cmd...).Run()
		if err != nil {
			return err
		}
	}
	return nil
}
