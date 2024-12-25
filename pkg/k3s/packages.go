package k3s

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// PackageConfig holds configuration for package operations
type PackageConfig struct {
	Host        string
	User        string
	Key         string
	Interface   string
	Timeout     int
	DHCPRange   string
	DNSServer   string
	TFTPRoot    string
	NetworkInfo []NetworkEntry
}

// NetworkEntry represents a DHCP host entry
type NetworkEntry struct {
	MAC string
	IP  string
}

// NewPackageConfig creates a new package configuration with defaults
func NewPackageConfig(host, user, key, interfaceName string, macs, ips []string) *PackageConfig {
	entries := make([]NetworkEntry, len(macs))
	for i := range macs {
		entries[i] = NetworkEntry{
			MAC: macs[i],
			IP:  ips[i],
		}
	}

	return &PackageConfig{
		Host:        host,
		User:        user,
		Key:         key,
		Interface:   interfaceName,
		Timeout:     600,
		DHCPRange:   "10.0.0.10,10.0.0.200,255.255.255.0",
		DNSServer:   "8.8.8.8",
		TFTPRoot:    "/srv/tftp",
		NetworkInfo: entries,
	}
}

// InstallPackages installs the required packages.
func InstallPackages(master, user, key string, packages string) (string, error) {
	cfg := &PackageConfig{
		Host:    master,
		User:    user,
		Key:     key,
		Timeout: 600,
	}
	return cfg.InstallPackages(packages)
}

// InstallPackages installs specified packages
func (p *PackageConfig) InstallPackages(packages string) (string, error) {
	commands := []struct {
		cmd  string
		desc string
	}{
		{
			cmd:  "sudo apt-get update",
			desc: "update package list",
		},
		{
			cmd:  fmt.Sprintf("sudo apt-get install -y %s", packages),
			desc: "install packages",
		},
	}

	for _, command := range commands {
		log.Printf("Running: %s", command.cmd)
		out, err := ssh.Run(command.cmd, p.Host, p.Key, p.User, "", false, p.Timeout)
		if err != nil {
			return string(out), fmt.Errorf("failed to %s: %s: %w", command.desc, out, err)
		}
	}

	return "", nil
}

// ConfigurePackages configures the installed packages.
func ConfigurePackages(master, user, key string, interfaceName string, macs, ips []string) (string, error) {
	cfg := NewPackageConfig(master, user, key, interfaceName, macs, ips)
	return cfg.Configure()
}

// Configure sets up DHCP and TFTP services
func (p *PackageConfig) Configure() (string, error) {
	if err := p.createTFTPDirectory(); err != nil {
		return "", err
	}

	if err := p.configureDNSMasq(); err != nil {
		return "", err
	}

	if err := p.enableAndRestartDNSMasq(); err != nil {
		return "", err
	}

	return "", nil
}

// createTFTPDirectory creates the TFTP root directory
func (p *PackageConfig) createTFTPDirectory() error {
	cmd := fmt.Sprintf("sudo mkdir %s || true", p.TFTPRoot)
	log.Println(cmd)

	out, err := ssh.Run(cmd, p.Host, p.Key, p.User, "", false, 10)
	if err != nil {
		return fmt.Errorf("failed to create TFTP directory: %s: %w", out, err)
	}

	return nil
}

// configureDNSMasq creates and applies DNSMasq configuration
func (p *PackageConfig) configureDNSMasq() error {
	config := p.generateDNSMasqConfig()
	log.Printf("DNSMasq config:\n%s", config)

	tempFile, err := os.CreateTemp("", "dnsmasq")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.WriteString(config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := ssh.Copy(tempFile.Name(), "/tmp/dnsmasq.conf", p.Host, p.User, p.Key); err != nil {
		return fmt.Errorf("failed to copy config file: %w", err)
	}

	cmd := "sudo mv /tmp/dnsmasq.conf /etc/dnsmasq.conf"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, p.Host, p.Key, p.User, "", false, 10); err != nil {
		return fmt.Errorf("failed to move config file: %s: %w", out, err)
	}

	return nil
}

// generateDNSMasqConfig generates DNSMasq configuration
func (p *PackageConfig) generateDNSMasqConfig() string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("interface=%s\n", p.Interface))
	config.WriteString("bind-interfaces\n")
	config.WriteString(fmt.Sprintf("dhcp-range=enp0s31f6,%s\n", p.DHCPRange))

	for _, entry := range p.NetworkInfo {
		config.WriteString(fmt.Sprintf("dhcp-host=%s,%s\n", entry.MAC, entry.IP))
	}

	config.WriteString(`dhcp-match=set:efi-x86_64,option:client-arch,7
dhcp-boot=tag:efi-x86_64,grubnetx64.efi.signed
dhcp-boot=pxelinux.0
enable-tftp
`)
	config.WriteString(fmt.Sprintf("tftp-root=%s\n", p.TFTPRoot))
	config.WriteString(fmt.Sprintf("server=%s\n", p.DNSServer))

	return config.String()
}

// enableAndRestartDNSMasq enables and restarts the DNSMasq service
func (p *PackageConfig) enableAndRestartDNSMasq() error {
	commands := []struct {
		cmd  string
		desc string
	}{
		{
			cmd:  "sudo systemctl enable dnsmasq",
			desc: "enable dnsmasq",
		},
		{
			cmd:  "sudo systemctl restart dnsmasq",
			desc: "restart dnsmasq",
		},
	}

	for _, command := range commands {
		log.Printf("Running: %s", command.cmd)
		out, err := ssh.Run(command.cmd, p.Host, p.Key, p.User, "", false, 10)
		if err != nil {
			// Try to get service status for better error reporting
			statusOut, _ := ssh.Run("sudo systemctl status dnsmasq", p.Host, p.Key, p.User, "", false, 10)
			return fmt.Errorf("failed to %s: %s\nService status: %s: %w", command.desc, out, statusOut, err)
		}
	}

	return nil
}
