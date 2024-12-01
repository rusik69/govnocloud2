package main

import (
	"fmt"

	"github.com/rusik69/govnocloud2/pkg/client"
	"github.com/spf13/cobra"
)

// client command
var clientCmd = &cobra.Command{
	Use:   "client [action] [args]",
	Short: "govnocloud2 client",
	Long:  `govnocloud2 client`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			panic("action is required")
		}
		if args[0] == "version" {
			fmt.Println("govnocloud2 client v0.0.1")
			serverVer, err := client.GetServerVersion(cfg.Client.Host, cfg.Client.Port)
			if err != nil {
				panic(err)
			}
			fmt.Println("server version: " + serverVer)
		}
	},
}
