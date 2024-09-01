package main

import (
	"os"

	"github.com/spf13/cobra"
)

var masterFlag, workersFlag, userFlag, keyFlag, kubeConfigPath, listenHost, listenPort string
var clientHost, clientPort string

// root command
var rootCmd = &cobra.Command{
	Use:   "govnocloud2 [install | uninstall]",
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
	uninstallCmd.Flags().StringVarP(&masterFlag, "master", "m", "", "master host")
	uninstallCmd.Flags().StringVarP(&workersFlag, "workers", "w", "", "workers hosts")
	uninstallCmd.Flags().StringVarP(&userFlag, "user", "u", "ubuntu", "ssh user")
	uninstallCmd.Flags().StringVarP(&keyFlag, "key", "k", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&masterFlag, "master", "m", "", "master host")
	installCmd.Flags().StringVarP(&workersFlag, "workers", "w", "", "workers hosts")
	installCmd.Flags().StringVarP(&userFlag, "user", "u", "ubuntu", "ssh user")
	installCmd.Flags().StringVarP(&keyFlag, "key", "k", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "c", defaultKubeConfigPath, "kubeconfig path")
	serverCmd.Flags().StringVarP(&listenHost, "host", "h", "0.0.0.0", "listen host")
	serverCmd.Flags().StringVarP(&listenPort, "port", "p", "8080", "listen port")
	clientCmd.Flags().StringVarP(&clientHost, "host", "h", "127.0.0.1", "server host")
	clientCmd.Flags().StringVarP(&clientPort, "port", "p", "8080", "server port")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
