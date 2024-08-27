package main

import (
	"log"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/k3s"
	"github.com/spf13/cobra"
)

var masterFlag, workersFlag, userFlag, keyFlag, kubeConfigPath string

// root command
var rootCmd = &cobra.Command{
	Use:   "govnocloud2 [install | uninstall]",
	Short: "govnocloud2 is a shitty cloud 2",
	Long:  `govnocloud2 is a shitty cloud 2`,
}

// install command
var installCmd = &cobra.Command{
	Use:   "install [master] [workers]",
	Short: "install govnocloud2 cluster",
	Long:  `install govnocloud2 cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("master: ", masterFlag)
		log.Println("workers: ", workersFlag)
		log.Println("user: ", userFlag)
		log.Println("key: ", keyFlag)
		if masterFlag == "" {
			panic("master is required")
		}
		workersSplit := strings.Split(workersFlag, ",")
		if len(workersSplit) == 0 {
			panic("workers are required")
		}
		log.Println("Deploying k3s master on " + masterFlag)
		err := k3s.DeployMaster(masterFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		token, err := k3s.GetToken(masterFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		for _, worker := range workersSplit {
			log.Println("Deploying k3s worker on " + worker)
			err := k3s.DeployNode(worker, userFlag, keyFlag, masterFlag, token)
			if err != nil {
				panic(err)
			}
		}
		log.Println("Getting kubeconfig")
		kubeConfigBody, err := k3s.GetKubeconfig(masterFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		err = k3s.WriteKubeConfig(kubeConfigBody, kubeConfigPath)
		if err != nil {
			panic(err)
		}
		log.Println("Kubeconfig is written to " + kubeConfigPath)
		err = k3s.InstallKubeVirt()
		if err != nil {
			panic(err)
		}
	},
}

// uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [master] [workers]",
	Short: "uninstall govnocloud2 cluster",
	Long:  `uninstall govnocloud2 cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("master: ", masterFlag)
		log.Println("workers: ", workersFlag)
		log.Println("user: ", userFlag)
		log.Println("key: ", keyFlag)
		if masterFlag == "" {
			panic("master is required")
		}
		workersSplit := strings.Split(workersFlag, ",")
		if len(workersSplit) == 0 {
			panic("workers are required")
		}
		log.Println("Uninstalling k3s master on " + masterFlag)
		err := k3s.UninstallMaster(masterFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		for _, worker := range workersSplit {
			log.Println("Uninstalling k3s worker on " + worker)
			err := k3s.UninstallNode(worker, userFlag, keyFlag)
			if err != nil {
				panic(err)
			}
		}
	},
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
	uninstallCmd.Flags().StringVarP(&masterFlag, "master", "m", "", "master host")
	uninstallCmd.Flags().StringVarP(&workersFlag, "workers", "w", "", "workers hosts")
	uninstallCmd.Flags().StringVarP(&userFlag, "user", "u", "ubuntu", "ssh user")
	uninstallCmd.Flags().StringVarP(&keyFlag, "key", "k", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&masterFlag, "master", "m", "", "master host")
	installCmd.Flags().StringVarP(&workersFlag, "workers", "w", "", "workers hosts")
	installCmd.Flags().StringVarP(&userFlag, "user", "u", "ubuntu", "ssh user")
	installCmd.Flags().StringVarP(&keyFlag, "key", "k", defaultKeyPath, "ssh key")
	installCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "c", defaultKubeConfigPath, "kubeconfig path")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
