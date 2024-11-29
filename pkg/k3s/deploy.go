package k3s

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// Deploy deploys the server.
func Deploy(host, port, user, password, key string) error {
	log.Println("Copying govnocloud2 to ", host)
	err := ssh.Copy("bin/govnocloud2-linux-amd64", "/usr/local/bin/govnocloud2", host, "root", key)
	if err != nil {
		return err
	}
	cmd := "sudo chmod +x /usr/local/bin/govnocloud2"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, key, user, password, false, 5)
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
	log.Println(serviceBody)
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
	log.Println("Copying govnocloud2 service to ", host)
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
	log.Println(serviceWebBody)
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
	log.Println("Copying govnocloud2-web service to ", host)
	err = ssh.Copy(file.Name(), "/etc/systemd/system/govnocloud2-web.service", host, "root", key)
	if err != nil {
		return err
	}
	cmd = "sudo systemctl daemon-reload"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, password, false, 5)
	if err != nil {
		return errors.New(string(out))
	}
	cmd = "sudo systemctl enable --now govnocloud2"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, password, false, 5)
	if err != nil {
		return errors.New(string(out))
	}
	cmd = "sudo systemctl enable --now govnocloud2-web"
	log.Println(cmd)
	out, err = ssh.Run(cmd, host, key, user, password, false, 5)
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

// Wol wakes on lan.
func Wol(master, user, key, ip string, macs []string) error {
	for _, mac := range macs {
		cmd := "wakeonlan -i " + ip + " " + mac
		log.Println(cmd)
		out, err := ssh.Run(cmd, master, key, user, "", false, 5)
		if err != nil {
			log.Println(err)
		}
		log.Println(out)
	}
	return nil
}

// Suspend suspends the servers
func Suspend(ips []string, master, user, password, key string) {
	for _, ip := range ips {
		log.Println("Suspending server: ", ip)
		cmd := fmt.Sprintf("ssh %s@%s 'sudo systemctl suspend'", user, ip)
		log.Println(cmd)
		out, err := ssh.Run(cmd, master, key, user, password, false, 5)
		log.Println(out)
		if err != nil {
			continue
		}
	}
}
