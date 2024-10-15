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
		log.Println("master: ", masterFlag)
		log.Println("workersips: ", workersIPs)
		log.Println("user: ", userFlag)
		log.Println("key: ", keyFlag)
		if masterFlag == "" {
			panic("master is required")
		}
		workersSplit := strings.Split(workersIPs, ",")
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
