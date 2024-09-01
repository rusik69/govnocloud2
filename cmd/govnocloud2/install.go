package main

import (
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/k3s"
	"github.com/rusik69/govnocloud2/pkg/server"
	"github.com/spf13/cobra"
)

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
		log.Println("Deploying server on " + masterFlag)
		err := server.Deploy(masterFlag, listenPort, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		log.Println("Deploying k3s master on " + masterFlag)
		err = k3s.DeployMaster(masterFlag, userFlag, keyFlag)
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
		log.Println("Installing Helm")
		err = k3s.InstallHelm()
		if err != nil {
			panic(err)
		}
		log.Println("Installing KubeVirt")
		err = k3s.InstallKubeVirt()
		if err != nil {
			panic(err)
		}
		log.Println("Installing rook")
		err = k3s.InstallRook()
		if err != nil {
			panic(err)
		}
	},
}
