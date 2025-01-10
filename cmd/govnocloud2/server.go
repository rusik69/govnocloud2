package main

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/server"
	"github.com/spf13/cobra"
)

// server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start govnocloud2 server",
	Long:  `start govnocloud2 server`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("listenHost: ", cfg.Server.Host)
		log.Println("listenPort: ", cfg.Server.Port)
		log.Println("masterHost: ", cfg.Server.MasterHost)
		log.Println("user: ", cfg.Server.User)
		log.Println("password: ", cfg.Server.Password)
		log.Println("key: ", cfg.Server.Key)
		server.Serve(cfg.Server)
	},
}
