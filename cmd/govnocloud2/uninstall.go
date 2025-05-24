package main

import (
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/k3s"
	"github.com/spf13/cobra"
)

// uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [master] [workers]",
	Short: "uninstall govnocloud2 cluster",
	Long:  `uninstall govnocloud2 cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("master: ", cfg.Master.Host)
		log.Println("workersips: ", cfg.Worker.IPs)
		log.Println("user: ", cfg.SSH.User)
		log.Println("key: ", cfg.SSH.KeyPath)

		if cfg.Master.Host == "" {
			panic("master is required")
		}

		workersSplit := strings.Split(cfg.Worker.IPs, ",")
		if len(workersSplit) == 0 {
			panic("workers are required")
		}

		log.Println("Uninstalling k3s master on " + cfg.Master.Host)
		err := k3s.UninstallMaster(
			cfg.Master.Host,
			cfg.SSH.User,
			cfg.SSH.KeyPath,
			cfg.SSH.Password,
		)
		if err != nil {
			panic(err)
		}

		for _, worker := range workersSplit {
			log.Println("Uninstalling k3s worker on " + worker)
			err := k3s.UninstallNode(
				cfg.Master.Host,
				worker,
				cfg.SSH.User,
				cfg.SSH.KeyPath,
				cfg.SSH.Password,
			)
			if err != nil {
				log.Println(err)
			}
		}
	},
}
