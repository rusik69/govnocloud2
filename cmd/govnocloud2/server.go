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
		log.Println("listenHost: ", listenHost)
		log.Println("listenPort: ", listenPort)
		server.Serve(listenHost, listenPort)
	},
}
