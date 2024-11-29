package main

import (
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/k3s"
	"github.com/rusik69/govnocloud2/pkg/ssh"
	"github.com/spf13/cobra"
)

// install command
var installCmd = &cobra.Command{
	Use:   "install [master] [workers]",
	Short: "install govnocloud2 cluster",
	Long:  `install govnocloud2 cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("master: ", masterFlag)
		log.Println("workers ips: ", workersIPs)
		log.Println("workers macs: ", workersMacs)
		log.Println("user: ", userFlag)
		log.Println("key: ", keyFlag)
		if masterFlag == "" {
			panic("master is required")
		}
		workersIPsSplit := strings.Split(workersIPs, ",")
		if len(workersIPsSplit) == 0 {
			panic("workers ips are required")
		}
		workersMacsSplit := strings.Split(workersMacs, ",")
		if len(workersMacsSplit) == 0 {
			panic("workers macs are required")
		}
		if len(workersIPsSplit) != len(workersMacsSplit) {
			panic("workers ips and macs should be the same length")
		}
		log.Println("Creating ssh key")
		_, err := ssh.CreateKey(masterFlag, masterKeyPath, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		if passwordFlag != "" {
			log.Println("Installing key on " + masterFlag)
			err := ssh.CopySSHKey(masterFlag, userFlag, passwordFlag, pubKeyPath, "")
			if err != nil {
				panic(err)
			}
		}
		log.Println("Installing packages on " + masterFlag)
		out, err := k3s.InstallPackages(masterFlag, userFlag, keyFlag, "sshpass wakeonlan dnsmasq")
		if err != nil {
			log.Println(out)
			panic(err)
		}
		log.Println("Configuring packages on " + masterFlag)
		out, err = k3s.ConfigurePackages(masterFlag, userFlag, keyFlag, interfaceName, workersMacsSplit, workersIPsSplit)
		if err != nil {
			log.Println(out)
			panic(err)
		}
		for _, worker := range workersIPsSplit {
			log.Println("Installing key on " + worker)
			err := ssh.CopySSHKey(worker, userFlag, passwordFlag, pubKeyPath, masterFlag)
			if err != nil {
				panic(err)
			}
		}
		log.Println("Deploying server on " + masterFlag)
		err = k3s.Deploy(masterFlag, listenPort, userFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		log.Println("Deploying k3s master on " + masterFlag)
		out, err = k3s.DeployMaster(masterFlag, userFlag, keyFlag)
		if err != nil {
			log.Println(out)
			panic(err)
		}
		log.Println(out)
		log.Println("Getting k3s token")
		token, err := k3s.GetToken(masterFlag, userFlag, keyFlag)
		if err != nil {
			panic(err)
		}
		for _, worker := range workersIPsSplit {
			log.Println("Deploying k3s worker on " + worker)
			out, err := k3s.DeployNode(worker, userFlag, keyFlag, passwordFlag, masterFlag, token)
			if err != nil {
				log.Println(out)
				panic(err)
			}
			log.Println(out)
		}
		log.Println("Installing Helm")
		err = k3s.InstallHelm()
		if err != nil {
			panic(err)
		}
		log.Println("Installing monitoring stack")
		err = k3s.DeployPrometheus()
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

// InstallPackages
