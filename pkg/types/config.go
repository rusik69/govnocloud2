package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all command line flags
type Config struct {
	Master       MasterConfig
	Worker       WorkerConfig
	SSH          SSHConfig
	Kube         KubeConfig
	Server       ServerConfig
	RootPassword string
	Web          WebConfig
	Client       ClientConfig
	Install      InstallConfig
}

type MasterConfig struct {
	Host         string
	KeyPath      string
	PubKeyPath   string
	Interface    string
	RootPassword string
}

type WorkerConfig struct {
	MACs      string
	IPs       string
	IPRange   string
	Interface string
}

type SSHConfig struct {
	User       string
	Password   string
	KeyPath    string
	PubKeyPath string
}

type KubeConfig struct {
	ConfigPath string
}

type WebConfig struct {
	Host       string
	Port       string
	Path       string
	MasterHost string
}

type ClientConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// InstallServerConfig is used for install-specific server configuration
type InstallServerConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	Key        string
	MasterHost string
	ImageDir   string
}

type MonitoringConfig struct {
	Enabled             bool
	GrafanaHost         string
	PrometheusHost      string
	AlertmanagerHost    string
	KubevirtManagerHost string
}

type DashboardConfig struct {
	Host string
}

type InstallConfig struct {
	Master     MasterConfig
	Workers    WorkerConfig
	SSH        SSHConfig
	Server     InstallServerConfig
	ImagesDir  string
	Monitoring MonitoringConfig
	Dashboard  DashboardConfig
	Longhorn   LonghornConfig
	Nat        NatConfig
	Web        WebConfig
}

type NatConfig struct {
	Enabled           bool
	ExternalInterface string
	InternalInterface string
}

type LonghornConfig struct {
	Host       string
	Disk       string
	FormatDisk bool
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()

	return Config{
		SSH: SSHConfig{
			KeyPath:    filepath.Join(homeDir, ".ssh/id_rsa"),
			PubKeyPath: filepath.Join(homeDir, ".ssh/id_rsa.pub"),
			User:       "ubuntu",
			Password:   "ubuntu",
		},
		Master: MasterConfig{
			KeyPath:    "~/.ssh/id_rsa",
			PubKeyPath: "~/.ssh/id_rsa.pub",
			Host:       "localhost",
			Interface:  "enp0s25",
		},
		Kube: KubeConfig{
			ConfigPath: filepath.Join(homeDir, ".kube/config"),
		},
		Worker: WorkerConfig{
			IPRange:   "10.0.0.0/24",
			Interface: "enp0s25",
		},
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         "6969",
			MasterHost:   "10.0.0.1",
			SSHUser:      "ubuntu",
			SSHPassword:  "ubuntu",
			Key:          filepath.Join(homeDir, ".ssh/id_rsa"),
			RootPassword: "password",
		},
		Web: WebConfig{
			Host:       "0.0.0.0",
			Port:       "8080",
			Path:       "/var/www/govnocloud2",
			MasterHost: "master.govno2.cloud",
		},
		Client: ClientConfig{
			Host:     "localhost",
			Port:     "6969",
			User:     "root",
			Password: "password",
		},
		Install: InstallConfig{
			Master: MasterConfig{
				KeyPath:      "~/.ssh/id_rsa",
				PubKeyPath:   "~/.ssh/id_rsa.pub",
				Interface:    "enp0s25",
				RootPassword: "password",
			},
			Workers: WorkerConfig{
				IPRange:   "10.0.0.0/24",
				Interface: "enp0s25",
			},
			SSH: SSHConfig{
				User:       "ubuntu",
				Password:   "ubuntu",
				KeyPath:    filepath.Join(homeDir, ".ssh/id_rsa"),
				PubKeyPath: filepath.Join(homeDir, ".ssh/id_rsa.pub"),
			},
			Server: InstallServerConfig{
				Host:       "0.0.0.0",
				Port:       "6969",
				User:       "ubuntu",
				Password:   "ubuntu",
				Key:        filepath.Join(homeDir, ".ssh/id_rsa"),
				MasterHost: "10.0.0.1",
				ImageDir:   "/var/lib/govnocloud2/images",
			},
			ImagesDir: "/var/lib/govnocloud2/images",
			Monitoring: MonitoringConfig{
				Enabled:             true,
				GrafanaHost:         "grafana.govno2.cloud",
				PrometheusHost:      "prometheus.govno2.cloud",
				AlertmanagerHost:    "alertmanager.govno2.cloud",
				KubevirtManagerHost: "kubevirt.govno2.cloud",
			},
			Dashboard: DashboardConfig{
				Host: "dashboard.govno2.cloud",
			},
			Nat: NatConfig{
				Enabled:           true,
				ExternalInterface: "wlp2s0",
				InternalInterface: "enp0s25",
			},
			Longhorn: LonghornConfig{
				Host:       "longhorn.govno2.cloud",
				Disk:       "sda",
				FormatDisk: false,
			},
			Web: WebConfig{
				Host:       "0.0.0.0",
				Port:       "8080",
				Path:       "/var/www/govnocloud2",
				MasterHost: "master.govno2.cloud",
			},
		},
	}
}

// ValidateConfig performs security checks on the configuration
func ValidateConfig(cfg Config) error {
	// Check if server is bound to localhost in development
	if cfg.Server.Host == "0.0.0.0" && os.Getenv("GOVNOCLOUD_ENV") != "production" {
		fmt.Println("Warning: Server is bound to all interfaces. In production, consider binding to specific interfaces.")
	}

	// Validate port numbers
	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port < 1024 || port > 65535 {
		return fmt.Errorf("invalid server port: %s", cfg.Server.Port)
	}
	if port, err := strconv.Atoi(cfg.Web.Port); err != nil || port < 1024 || port > 65535 {
		return fmt.Errorf("invalid web port: %s", cfg.Web.Port)
	}

	// Check if sensitive files have proper permissions
	if err := checkFilePermissions(cfg.SSH.KeyPath); err != nil {
		return err
	}

	return nil
}

// checkFilePermissions verifies that sensitive files have proper permissions
func checkFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return nil // File doesn't exist yet, that's okay
	}

	mode := info.Mode()
	if mode&0077 != 0 {
		return fmt.Errorf("file %s has too permissive permissions: %v", path, mode)
	}

	return nil
}
