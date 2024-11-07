package server

import (
	"errors"
	"log"
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// Deploy deploys the server.
func Deploy(host, port, user, password, key string) error {
	err := ssh.Copy("bin/govnocloud2-linux-amd64", "/usr/local/bin/govnocloud2", host, "root", key)
	if err != nil {
		return err
	}
	out, err := ssh.Run("chmod +x /usr/local/bin/govnocloud2", host, key, "root", password, false)
	if err != nil {
		return errors.New(string(out))
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
	file, err := os.CreateTemp("", "govnocloud2.service")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())
	_, err = file.WriteString(serviceBody)
	if err != nil {
		return err
	}
	err = ssh.Copy(file.Name(), "/etc/systemd/system/govnocloud2.service", host, "root", key)
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
	file, err = os.CreateTemp("", "govnocloud2-web.service")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())
	_, err = file.WriteString(serviceWebBody)
	if err != nil {
		return err
	}
	err = ssh.Copy(file.Name(), "/etc/systemd/system/govnocloud2-web.service", host, "root", key)
	if err != nil {
		return err
	}
	out, err = ssh.Run("sudo systemctl daemon-reload", host, key, user, password, false)
	if err != nil {
		return errors.New(string(out))
	}
	out, err = ssh.Run("sudo systemctl enable --now govnocloud2", host, key, user, password, false)
	if err != nil {
		return errors.New(string(out))
	}
	out, err = ssh.Run("sudo systemctl enable --now govnocloud2-web", host, key, user, password, false)
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

// Wol wakes on lan.
func Wol(master, user, key, ip string, macs []string) error {
	for _, mac := range macs {
		cmd := "wakeonlan -i " + ip + " " + mac
		out, err := ssh.Run(cmd, master, key, user, "", false)
		if err != nil {
			return errors.New(string(out))
		}
	}
	return nil
}

// Suspend suspends the servers
func Suspend(ips []string, user, password, key string) {
	for _, ip := range ips {
		log.Println("Suspending server: ", ip)
		cmd := "sudo systemctl suspend"
		out, err := ssh.Run(cmd, ip, key, user, password, false)
		log.Println(out)
		if err != nil {
			continue
		}
	}
}
