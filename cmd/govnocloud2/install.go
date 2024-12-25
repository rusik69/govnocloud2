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
		log.Println("master: ", cfg.Install.Master.Host)
		log.Println("workers ips: ", cfg.Install.Workers.IPs)
		log.Println("workers macs: ", cfg.Install.Workers.MACs)
		log.Println("user: ", cfg.Install.SSH.User)
		log.Println("key: ", cfg.Install.SSH.KeyPath)

		if cfg.Install.Master.Host == "" {
			panic("master is required")
		}

		workersIPsSplit := strings.Split(cfg.Install.Workers.IPs, ",")
		if len(workersIPsSplit) == 0 {
			panic("workers ips are required")
		}

		workersMacsSplit := strings.Split(cfg.Install.Workers.MACs, ",")
		if len(workersMacsSplit) == 0 {
			panic("workers macs are required")
		}

		if len(workersIPsSplit) != len(workersMacsSplit) {
			panic("workers ips and macs should be the same length")
		}

		log.Println("Installing packages on " + cfg.Install.Master.Host)
		out, err := k3s.InstallPackages(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			"sshpass wakeonlan dnsmasq",
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}

		log.Println("Configuring packages on " + cfg.Install.Master.Host)
		out, err = k3s.ConfigurePackages(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.Workers.Interface,
			workersMacsSplit,
			workersIPsSplit,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}

		log.Println("Deploying server on " + cfg.Install.Master.Host)
		err = k3s.Deploy(
			cfg.Install.Master.Host,
			cfg.Server.Host,
			cfg.Web.Host,
			cfg.Server.Port,
			cfg.Web.Port,
			cfg.Install.SSH.User,
			cfg.Install.SSH.Password,
			cfg.Install.SSH.KeyPath,
			cfg.Web.Path,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Downloading VM images")
		err = k3s.DownloadVMImages(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.ImagesDir,
		)
		if err != nil {
			panic(err)
		}

		// Install k3sup tool
		log.Println("Installing k3sup tool on " + cfg.Install.Master.Host)
		err = k3s.InstallK3sUp(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Deploying k3s master on " + cfg.Install.Master.Host)
		err = k3s.DeployMaster(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}

		for _, worker := range workersIPsSplit {
			log.Println("Deploying k3s worker on " + worker)
			err = k3s.DeployNode(
				worker,
				cfg.Install.SSH.User,
				cfg.Install.SSH.KeyPath,
				cfg.Install.SSH.Password,
				cfg.Install.Master.Host,
			)
			if err != nil {
				panic(err)
			}
		}

		command := "sudo k3s kubectl get nodes"
		out, err = ssh.Run(
			command,
			cfg.Install.Master.Host,
			cfg.Install.SSH.KeyPath,
			cfg.Install.SSH.User,
			"",
			false,
			60,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}
		log.Println("Nodes:")
		log.Println(out)

		log.Println("Installing Helm")
		err = k3s.InstallHelm(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing monitoring stack")
		err = k3s.DeployPrometheus(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing KubeVirt")
		err = k3s.InstallKubeVirt(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Longhorn")
		err = k3s.InstallLonghorn(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}
	},
}
