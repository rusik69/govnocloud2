package server

import (
	"fmt"
	"log"
	"os"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallPackages installs the packages.
func InstallPackages(master, user, key string, packages string) (string, error) {
	command := "sudo apt-get update"
	log.Println(command)
	out, err := ssh.Run(command, master, key, user, "", false)
	if err != nil {
		return string(out), nil
	}
	out, err = ssh.Run(fmt.Sprintf("sudo apt-get install -y %s", packages), master, key, user, "", false)
	if err != nil {
		return string(out), err
	}
	return "", nil
}

// ConfigurePackages configures the packages.
func ConfigurePackages(master, user, key string, macs, ips []string) (string, error) {
	out, err := ssh.Run("mkdir /srv/tftp", master, key, user, "", false)
	if err != nil {
		return string(out), err
	}
	dnsmasqConfig := `interface=enp7s0
bind-interfaces
dhcp-range=enp0s31f6,10.0.0.10,10.0.0.200,255.255.255.0
`
	for i, mac := range macs {
		dnsmasqConfig += "dhcp-host=" + mac + "," + ips[i] + "\n"
	}
	dnsmasqConfig += `dhcp-match=set:efi-x86_64,option:client-arch,7
dhcp-boot=tag:efi-x86_64,grubnetx64.efi.signed
dhcp-boot=pxelinux.0
enable-tftp
tftp-root=/srv/tftp
server=8.8.8.8
`
	tempFile, err := os.CreateTemp("", "dnsmasq")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	_, err = tempFile.WriteString(dnsmasqConfig)
	if err != nil {
		return "", err
	}
	err = ssh.Copy(tempFile.Name(), "/etc/dnsmasq.conf", master, user, key)
	if err != nil {
		return "", err
	}
	out, err = ssh.Run("sudo systemctl enable dnsmasq", master, key, user, "", false)
	if err != nil {
		return string(out), err
	}
	_, err = ssh.Run("sudo systemctl restart dnsmasq", master, key, user, "", false)
	if err != nil {
		statusOut, statusErr := ssh.Run("sudo systemctl status dnsmasq", master, key, user, "", false)
		if statusErr != nil {
			return string(statusOut), statusErr
		}
		return string(statusOut), err
	}
	return "", nil
}
