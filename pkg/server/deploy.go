package server

import (
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// Deploy starts the server.
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
	cmd = "sudo systemctl start govnocloud2"
	_, err = ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	return nil
}
