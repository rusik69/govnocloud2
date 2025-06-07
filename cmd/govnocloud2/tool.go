package main

import (
	"fmt"
	"log"
	"strings"

	k8s "github.com/rusik69/govnocloud2/pkg/k8s"
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
		macsSplit := strings.Split(cfg.Worker.MACs, ",")
		if len(macsSplit) == 0 {
			panic("macs are required")
		}
		if cfg.Master.Host == "" {
			panic("master is required")
		}
		log.Println("running WOL on host ", cfg.Master.Host, " macs ", macsSplit)
		k8s.Wol(
			cfg.Master.Host,
			cfg.SSH.User,
			cfg.SSH.KeyPath,
			cfg.Worker.IPRange,
			macsSplit,
		)
	},
}

var suspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: "suspend",
	Long:  `suspend`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("suspend")
		ipsSplit := strings.Split(cfg.Worker.IPs, ",")
		if len(ipsSplit) == 0 {
			panic("ips are required")
		}
		if cfg.Master.Host == "" {
			panic("master is required")
		}
		log.Println("master: " + cfg.Master.Host)
		k8s.Suspend(
			ipsSplit,
			cfg.Master.Host,
			cfg.SSH.User,
			cfg.SSH.Password,
			cfg.SSH.KeyPath,
		)
	},
}
