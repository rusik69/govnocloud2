package main

import (
	"fmt"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/server"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "tool commands",
	Long:  `tool commands`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tool commands: wol, suspend")
	},
}

var wolCmd = &cobra.Command{
	Use:   "wol",
	Short: "wake on lan",
	Long:  `wake on lan`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wol")
		macsSplit := strings.Split(workersMacs, ",")
		if len(macsSplit) == 0 {
			panic("macs are required")
		}
		server.Wol(ipRange, macsSplit)
	},
}

var suspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: "suspend",
	Long:  `suspend`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("suspend")
		ipsSplit := strings.Split(workersIPs, ",")
		if len(ipsSplit) == 0 {
			panic("ips are required")
		}
		server.Suspend(ipsSplit, userFlag, keyFlag)
	},
}
