package server

import (
	"os"
	"os/exec"
)

// InstallPackages installs the packages.
func InstallPackages(packages []string) (string, error) {
	out, err := exec.Command("apt-get", "update").CombinedOutput()
	if err != nil {
		return string(out), err
	}
	command := []string{"install", "-y"}
	command = append(command, packages...)
	out, err = exec.Command("apt-get", command...).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return "", nil
}

// ConfigurePackages configures the packages.
func ConfigurePackages(macs, ips []string) (string, error) {
	os.Mkdir("/srv/tftp", 0755)
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
		out, err := exec.Command("sudo", "systemctl", "status", "dnsmasq").CombinedOutput()
		if err != nil {
			return string(out), err
		}
		return string(out), err
	}
	return "", nil
}
