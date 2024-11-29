package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/k3s"
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
		macsSplit := strings.Split(workersMacs, ",")
		if len(macsSplit) == 0 {
			panic("macs are required")
		}
		if masterFlag == "" {
			panic("master is required")
		}
		log.Println("running WOL on host ", masterFlag, " macs ", macsSplit)
		k3s.Wol(masterFlag, userFlag, keyFlag, ipRange, macsSplit)
	},
}

var suspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: "suspend",
	Long:  `suspend`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("suspend")
		ipsSplit := strings.Split(workersIPs, ",")
		if len(ipsSplit) == 0 {
			panic("ips are required")
		}
		if masterFlag == "" {
			panic("master is required")
		}
		log.Println("master: " + masterFlag)
		k3s.Suspend(ipsSplit, masterFlag, userFlag, passwordFlag, keyFlag)
	},
}
