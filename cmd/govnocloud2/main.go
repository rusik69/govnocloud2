package main

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cfg     types.Config
	rootCmd = &cobra.Command{
		Use:   "govnocloud2 [install | uninstall | server | client | web | tool]",
		Short: "govnocloud2 is a shitty cloud 2",
		Long:  `govnocloud2 is a shitty cloud 2`,
	}
)

func initConfig() error {
	// Initialize with default configuration
	cfg = types.DefaultConfig()

	// Validate configuration
	if err := types.ValidateConfig(cfg); err != nil {
		return err
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
	flags.StringVarP(&cfg.Server.SSHUser, "user", "", cfg.Server.SSHUser, "ssh user")
	flags.StringVarP(&cfg.Server.SSHPassword, "password", "", cfg.Server.SSHPassword, "ssh password")
	flags.StringVarP(&cfg.Server.Key, "key", "", cfg.Server.Key, "ssh key")
	flags.StringVarP(&cfg.Server.MasterHost, "master", "", cfg.Server.MasterHost, "master host")
	flags.StringVarP(&cfg.Server.RootPassword, "rootpassword", "", cfg.Server.RootPassword, "root password")
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
