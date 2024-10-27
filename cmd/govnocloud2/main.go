package main

import (
	"os"

	"github.com/spf13/cobra"
)

var masterFlag, workersMacs, workersIPs, userFlag, passwordFlag, keyFlag, kubeConfigPath, listenHost, listenPort string
var clientHost, clientPort, webHost, webPort, ipRange string

// root command
var rootCmd = &cobra.Command{
	Use:   "govnocloud2 [install | uninstall | server | client | web]",
	Short: "govnocloud2 is a shitty cloud 2",
	Long:  `govnocloud2 is a shitty cloud 2`,
}

func init() {
	usr, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultKeyPath := usr + "/.ssh/id_rsa"
	defaultKubeConfigPath := usr + "/.kube/config"
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(toolCmd)
	toolCmd.AddCommand(wolCmd)
	toolCmd.AddCommand(suspendCmd)
	uninstallCmd.Flags().StringVarP(&masterFlag, "master", "", "", "master host")
	uninstallCmd.Flags().StringVarP(&workersIPs, "ips", "", "", "workers ips")
	uninstallCmd.Flags().StringVarP(&userFlag, "user", "", "ubuntu", "ssh user")
	uninstallCmd.Flags().StringVarP(&keyFlag, "key", "", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&masterFlag, "master", "", "", "master host")
	installCmd.Flags().StringVarP(&workersMacs, "macs", "", "", "workers mac addresses")
	installCmd.Flags().StringVarP(&workersIPs, "ips", "", "", "workers ip addresses")
	installCmd.Flags().StringVarP(&ipRange, "iprange", "", "10.0.0.0/24", "workers ip range")
	installCmd.Flags().StringVarP(&userFlag, "user", "", "ubuntu", "ssh user")
	installCmd.Flags().StringVarP(&passwordFlag, "password", "", "ubuntu", "ssh password")
	installCmd.Flags().StringVarP(&keyFlag, "key", "", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", defaultKubeConfigPath, "kubeconfig path")
	serverCmd.Flags().StringVarP(&listenHost, "host", "", "0.0.0.0", "listen host")
	serverCmd.Flags().StringVarP(&listenPort, "port", "", "6969", "listen port")
	clientCmd.Flags().StringVarP(&clientHost, "host", "", "127.0.0.1", "server host")
	clientCmd.Flags().StringVarP(&clientPort, "port", "", "6969", "server port")
	webCmd.Flags().StringVarP(&webHost, "host", "", "0.0.0.0", "listen host")
	webCmd.Flags().StringVarP(&webPort, "port", "", "8080", "listen port")
	wolCmd.Flags().StringVarP(&workersMacs, "macs", "", "", "comma separated mac addresses")
	wolCmd.Flags().StringVarP(&ipRange, "iprange", "", "", "ip range")
	suspendCmd.Flags().StringVarP(&workersIPs, "ips", "", "", "comma separated ips")
	suspendCmd.Flags().StringVarP(&userFlag, "user", "", "ubuntu", "ssh user")
	suspendCmd.Flags().StringVarP(&keyFlag, "key", "", defaultKeyPath, "ssh key")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
