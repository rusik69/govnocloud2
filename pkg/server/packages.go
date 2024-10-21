package server

import (
	"os"
	"os/exec"
)

// InstallPackages installs the packages.
func InstallPackages(packages []string) (string, error) {
	cmd := exec.Command("apt-get", "update")
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	command := []string{"install", "-y"}
	command = append(command, packages...)
	cmd = exec.Command("apt-get", command...)
	err = cmd.Run()
	if err != nil {
		out, _ := cmd.CombinedOutput()
		return string(out), err
	}
	return "", nil
}

// ConfigurePackages configures the packages.
func ConfigurePackages(macs, ips []string) (string, error) {
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
	err := os.WriteFile("/etc/dnsmasq.conf", []byte(dnsmasqConfig), 0644)
	if err != nil {
		return "", err
	}
	err = exec.Command("sudo", "systemctl", "enable", "dnsmasq").Run()
	if err != nil {
		return "", err
	}
	err = exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run()
	if err != nil {
		return "", err
	}
	return "", nil
}
