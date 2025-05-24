package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"encoding/base64"

	"github.com/rusik69/govnocloud2/pkg/types"
	"github.com/spf13/cobra"
)

// Config holds all command line flags
type Config struct {
	Master  MasterConfig
	Worker  WorkerConfig
	SSH     SSHConfig
	Kube    KubeConfig
	Server  types.ServerConfig
	Web     WebConfig
	Client  ClientConfig
	Install InstallConfig
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

type ServerConfig struct {
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
	Server     ServerConfig
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

var (
	cfg     Config
	rootCmd = &cobra.Command{
		Use:   "govnocloud2 [install | uninstall | server | client | web | tool]",
		Short: "govnocloud2 is a shitty cloud 2",
		Long:  `govnocloud2 is a shitty cloud 2`,
	}
)

func initConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Generate a random password if not set
	defaultPassword := os.Getenv("GOVNOCLOUD_DEFAULT_PASSWORD")
	if defaultPassword == "" {
		// Generate a secure random password
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to generate random password: %v", err)
		}
		defaultPassword = base64.URLEncoding.EncodeToString(b)
	}

	cfg = Config{
		SSH: SSHConfig{
			KeyPath:    filepath.Join(homeDir, ".ssh/id_rsa"),
			PubKeyPath: filepath.Join(homeDir, ".ssh/id_rsa.pub"),
			User:       os.Getenv("GOVNOCLOUD_SSH_USER"),
			Password:   defaultPassword,
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
		Server: types.ServerConfig{
			Host:       os.Getenv("GOVNOCLOUD_SERVER_HOST"),
			Port:       os.Getenv("GOVNOCLOUD_SERVER_PORT"),
			MasterHost: "10.0.0.1",
			User:       os.Getenv("GOVNOCLOUD_SERVER_USER"),
			Password:   defaultPassword,
			Key:        "/home/ubuntu/.ssh/id_rsa",
		},
		Web: WebConfig{
			Host:       os.Getenv("GOVNOCLOUD_WEB_HOST"),
			Port:       os.Getenv("GOVNOCLOUD_WEB_PORT"),
			Path:       "/var/www/govnocloud2",
			MasterHost: "master.govno2.cloud",
		},
		Client: ClientConfig{
			Host:     os.Getenv("GOVNOCLOUD_CLIENT_HOST"),
			Port:     os.Getenv("GOVNOCLOUD_CLIENT_PORT"),
			User:     os.Getenv("GOVNOCLOUD_CLIENT_USER"),
			Password: defaultPassword,
		},
		Install: InstallConfig{
			Master: MasterConfig{
				KeyPath:      "~/.ssh/id_rsa",
				PubKeyPath:   "~/.ssh/id_rsa.pub",
				Interface:    "enp0s25",
				RootPassword: defaultPassword,
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
		},
	}

	// Validate configuration
	if err := validateConfig(); err != nil {
		return err
	}

	return nil
}

// validateConfig performs security checks on the configuration
func validateConfig() error {
	// Check if server is bound to localhost in development
	if cfg.Server.Host == "0.0.0.0" && os.Getenv("GOVNOCLOUD_ENV") != "production" {
		log.Println("Warning: Server is bound to all interfaces. In production, consider binding to specific interfaces.")
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
	if err := checkFilePermissions(cfg.SSH.PubKeyPath); err != nil {
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

func setupCommands() {
	commands := []*cobra.Command{
		installCmd,
		uninstallCmd,
		serverCmd,
		clientCmd,
		webCmd,
		toolCmd,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	toolCmd.AddCommand(wolCmd, suspendCmd)
}

func setupInstallFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&cfg.Install.Master.Host, "master", "", "", "master host")
	flags.StringVarP(&cfg.Install.Workers.MACs, "macs", "", "", "workers mac addresses")
	flags.StringVarP(&cfg.Install.Workers.IPs, "ips", "", "", "workers ip addresses")
	flags.StringVarP(&cfg.Install.Workers.IPRange, "iprange", "", cfg.Install.Workers.IPRange, "workers ip range")
	flags.StringVarP(&cfg.Install.SSH.User, "user", "", cfg.Install.SSH.User, "ssh user")
	flags.StringVarP(&cfg.Install.SSH.Password, "password", "", cfg.Install.SSH.Password, "ssh password")
	flags.StringVarP(&cfg.Install.SSH.KeyPath, "key", "", cfg.Install.SSH.KeyPath, "ssh key")
	flags.StringVarP(&cfg.Install.SSH.PubKeyPath, "pubkey", "", cfg.Install.SSH.PubKeyPath, "ssh public key")
	flags.StringVarP(&cfg.Install.Master.PubKeyPath, "masterpubkey", "", cfg.Install.Master.PubKeyPath, "master public key path")
	flags.StringVarP(&cfg.Install.Master.KeyPath, "masterkey", "", cfg.Install.Master.KeyPath, "master key path")
	flags.StringVarP(&cfg.Install.Master.RootPassword, "masterrootpassword", "", cfg.Install.Master.RootPassword, "master root password")
	flags.StringVarP(&cfg.Install.Master.Interface, "masterinterface", "", cfg.Install.Master.Interface, "master interface name")
	flags.StringVarP(&cfg.Install.Workers.Interface, "interface", "", cfg.Install.Workers.Interface, "interface name")
	flags.StringVarP(&cfg.Install.ImagesDir, "imagesdir", "", cfg.Install.ImagesDir, "images directory")
	flags.BoolVarP(&cfg.Install.Monitoring.Enabled, "monitoring", "", cfg.Install.Monitoring.Enabled, "enable monitoring")
	flags.StringVarP(&cfg.Install.Monitoring.GrafanaHost, "grafanahost", "", cfg.Install.Monitoring.GrafanaHost, "grafana host")
	flags.StringVarP(&cfg.Install.Monitoring.PrometheusHost, "prometheushost", "", cfg.Install.Monitoring.PrometheusHost, "prometheus host")
	flags.StringVarP(&cfg.Install.Monitoring.AlertmanagerHost, "alertmanagerhost", "", cfg.Install.Monitoring.AlertmanagerHost, "alertmanager host")
	flags.StringVarP(&cfg.Install.Monitoring.KubevirtManagerHost, "kubevirtmanagerhost", "", cfg.Install.Monitoring.KubevirtManagerHost, "kubevirt manager host")
	flags.StringVarP(&cfg.Install.Dashboard.Host, "dashboardhost", "", cfg.Install.Dashboard.Host, "dashboard host")
	flags.StringVarP(&cfg.Install.Longhorn.Host, "longhornhost", "", cfg.Install.Longhorn.Host, "longhorn host")
	flags.StringVarP(&cfg.Install.Longhorn.Disk, "longhorndisk", "", cfg.Install.Longhorn.Disk, "longhorn disk")
	flags.BoolVarP(&cfg.Install.Longhorn.FormatDisk, "longhornformatdisk", "", cfg.Install.Longhorn.FormatDisk, "format longhorn disk")
	flags.BoolVarP(&cfg.Install.Nat.Enabled, "nat", "", cfg.Install.Nat.Enabled, "enable nat")
	flags.StringVarP(&cfg.Install.Nat.ExternalInterface, "nat-external-interface", "", cfg.Install.Nat.ExternalInterface, "external interface")
	flags.StringVarP(&cfg.Install.Nat.InternalInterface, "nat-internal-interface", "", cfg.Install.Nat.InternalInterface, "internal interface")
	flags.StringVarP(&cfg.Install.Web.Host, "web-host", "", cfg.Install.Web.Host, "web host")
	flags.StringVarP(&cfg.Install.Web.Port, "web-port", "", cfg.Install.Web.Port, "web port")
	flags.StringVarP(&cfg.Install.Web.Path, "web-path", "", cfg.Install.Web.Path, "web path")
	flags.StringVarP(&cfg.Install.Web.MasterHost, "web-master-host", "", cfg.Install.Web.MasterHost, "web master host")
}

func setupUninstallFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&cfg.Master.Host, "master", "", "", "master host")
	flags.StringVarP(&cfg.Worker.IPs, "ips", "", "", "workers ips")
	flags.StringVarP(&cfg.SSH.User, "user", "", cfg.SSH.User, "ssh user")
	flags.StringVarP(&cfg.SSH.KeyPath, "key", "", cfg.SSH.KeyPath, "ssh key")
}

func setupServerFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&cfg.Server.Host, "host", "", cfg.Server.Host, "listen host")
	flags.StringVarP(&cfg.Server.Port, "port", "", cfg.Server.Port, "listen port")
	flags.StringVarP(&cfg.Server.User, "user", "", cfg.Server.User, "ssh user")
	flags.StringVarP(&cfg.Server.Password, "password", "", cfg.Server.Password, "ssh password")
	flags.StringVarP(&cfg.Server.Key, "key", "", cfg.Server.Key, "ssh key")
	flags.StringVarP(&cfg.Server.MasterHost, "master", "", cfg.Server.MasterHost, "master host")
}

func setupClientFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&cfg.Client.Host, "host", "", cfg.Client.Host, "server host")
	flags.StringVarP(&cfg.Client.Port, "port", "", cfg.Client.Port, "server port")
	flags.StringVarP(&cfg.Client.User, "user", "", cfg.Client.User, "server username")
	flags.StringVarP(&cfg.Client.Password, "password", "", cfg.Client.Password, "server password")
}

func setupWebFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&cfg.Web.Host, "host", "", cfg.Web.Host, "listen host")
	flags.StringVarP(&cfg.Web.Port, "port", "", cfg.Web.Port, "listen port")
	flags.StringVarP(&cfg.Web.Path, "webpath", "", cfg.Web.Path, "web path")
}

func setupToolFlags(wolCmd, suspendCmd *cobra.Command) {
	wolFlags := wolCmd.Flags()
	wolFlags.StringVarP(&cfg.Worker.MACs, "macs", "", "", "comma separated mac addresses")
	wolFlags.StringVarP(&cfg.Worker.IPRange, "iprange", "", "", "ip range")
	wolFlags.StringVarP(&cfg.Master.Host, "master", "", cfg.Master.Host, "master host")

	suspendFlags := suspendCmd.Flags()
	suspendFlags.StringVarP(&cfg.Worker.IPs, "ips", "", "", "comma separated ips")
	suspendFlags.StringVarP(&cfg.SSH.User, "user", "", cfg.SSH.User, "ssh user")
	suspendFlags.StringVarP(&cfg.SSH.KeyPath, "key", "", cfg.SSH.KeyPath, "ssh key")
	suspendFlags.StringVarP(&cfg.Master.Host, "master", "", cfg.Master.Host, "master host")
}

func init() {
	log.SetFlags(0)
	if err := initConfig(); err != nil {
		panic(err)
	}

	setupCommands()
	setupInstallFlags(installCmd)
	setupUninstallFlags(uninstallCmd)
	setupServerFlags(serverCmd)
	setupClientFlags(clientCmd)
	setupWebFlags(webCmd)
	setupToolFlags(wolCmd, suspendCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
