package server

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// Deploy deploys the server.
func Deploy(host, port, user, password, key string) error {
	from, err := os.Open("bin/govnocloud2-linux-amd64")
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.Create("/usr/local/bin/govnocloud2")
	if err != nil {
		return err
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	err = os.Chmod("/usr/local/bin/govnocloud2", os.FileMode(0755))
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
	file, err := os.Create("/etc/systemd/system/govnocloud2.service")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(serviceBody)
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
	file, err = os.Create("/etc/systemd/system/govnocloud2-web.service")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(serviceWebBody)
	if err != nil {
		return err
	}
	err = exec.Command("systemctl", "daemon-reload").Run()
	if err != nil {
		return err
	}
	err = exec.Command("systemctl", "enable", "govnocloud2").Run()
	if err != nil {
		return err
	}
	err = exec.Command("systemctl", "enable", "govnocloud2-web").Run()
	if err != nil {
		return err
	}
	err = exec.Command("systemctl", "start", "govnocloud2").Run()
	if err != nil {
		return err
	}
	err = exec.Command("systemctl", "start", "govnocloud2-web").Run()
	if err != nil {
		return err
	}
	return nil
}

// Wol wakes on lan.
func Wol(ip string, macs []string) error {
	for _, mac := range macs {
		cmd := []string{"-i", ip, mac}
		err := exec.Command("wakeonlan", cmd...).Run()
		if err != nil {
			return err
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
