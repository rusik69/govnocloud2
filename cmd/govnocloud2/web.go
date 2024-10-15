package main

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/web"
	"github.com/spf13/cobra"
)

// web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "start govnocloud2 web",
	Long:  `start govnocloud2 web`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("starting web server")
		web.Listen(webHost, webPort)
	},
}
