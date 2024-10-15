package main

import (
	"os"

	"github.com/spf13/cobra"
)

var masterFlag, workersMacs, workersIPs, userFlag, passwordFlag, keyFlag, kubeConfigPath, listenHost, listenPort string
var clientHost, clientPort, webHost, webPort string

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
	uninstallCmd.Flags().StringVarP(&masterFlag, "master", "", "", "master host")
	uninstallCmd.Flags().StringVarP(&workersIPs, "workersips", "", "", "workers ips")
	uninstallCmd.Flags().StringVarP(&userFlag, "user", "", "root", "ssh user")
	uninstallCmd.Flags().StringVarP(&keyFlag, "key", "", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&masterFlag, "master", "", "", "master host")
	installCmd.Flags().StringVarP(&workersMacs, "workersmacs", "", "", "workers mac addresses")
	installCmd.Flags().StringVarP(&workersIPs, "workersips", "", "", "workers ip addresses")
	installCmd.Flags().StringVarP(&userFlag, "user", "", "root", "ssh user")
	installCmd.Flags().StringVarP(&passwordFlag, "password", "", "", "ssh password")
	installCmd.Flags().StringVarP(&keyFlag, "key", "", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", defaultKubeConfigPath, "kubeconfig path")
	serverCmd.Flags().StringVarP(&listenHost, "host", "", "0.0.0.0", "listen host")
	serverCmd.Flags().StringVarP(&listenPort, "port", "", "6969", "listen port")
	clientCmd.Flags().StringVarP(&clientHost, "host", "", "127.0.0.1", "server host")
	clientCmd.Flags().StringVarP(&clientPort, "port", "", "6969", "server port")
	webCmd.Flags().StringVarP(&webHost, "host", "", "0.0.0.0", "listen host")
	webCmd.Flags().StringVarP(&webPort, "port", "", "8080", "listen port")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
