package main

import (
	"fmt"
	"strconv"

	"github.com/rusik69/govnocloud2/pkg/client"
	"github.com/rusik69/govnocloud2/pkg/types"
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

		c := client.NewClient(cfg.Client.Host, cfg.Client.Port)

		switch args[0] {
		case "version":
			serverVer, err := client.GetServerVersion(cfg.Client.Host, cfg.Client.Port)
			if err != nil {
				panic(err)
			}
			fmt.Println("server version:", serverVer)

		case "nodes":
			if len(args) < 2 {
				panic("nodes subcommand required [list|get|add|delete|restart]")
			}
			handleNodes(c, args[1:])

		case "vms":
			if len(args) < 2 {
				panic("vms subcommand required [list|get|create|delete]")
			}
			handleVMs(c, args[1:])

		case "containers":
			if len(args) < 2 {
				panic("containers subcommand required [list|get|create|delete]")
			}
			handleContainers(c, args[1:])

		case "dbs":
			if len(args) < 2 {
				panic("dbs subcommand required [list|get|create|delete]")
			}
			handleDBs(c, args[1:])

		case "volumes":
			if len(args) < 2 {
				panic("volumes subcommand required [list|get|create|delete]")
			}
			handleVolumes(c, args[1:])

		case "namespaces":
			if len(args) < 2 {
				panic("namespaces subcommand required [list|get|create|delete]")
			}
			handleNamespaces(c, args[1:])

		default:
			panic("unknown action: " + args[0])
		}
	},
}

func handleNodes(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		nodes, err := c.ListNodes()
		if err != nil {
			panic(err)
		}
		for _, node := range nodes {
			fmt.Println(node)
		}

	case "get":
		if len(args) < 2 {
			panic("node name required")
		}
		node, err := c.GetNode(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", node)

	case "add":
		if len(args) < 6 {
			panic("required: name host masterhost user key")
		}
		node := types.Node{
			Name:       args[1],
			Host:       args[2],
			MasterHost: args[3],
			User:       args[4],
			Key:        args[5],
		}
		err := c.AddNode(node)
		if err != nil {
			panic(err)
		}
		fmt.Println("node added successfully")

	case "delete":
		if len(args) < 2 {
			panic("node name required")
		}
		err := c.DeleteNode(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Println("node deleted successfully")

	case "restart":
		if len(args) < 2 {
			panic("node name required")
		}
		err := c.RestartNode(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Println("node restarted successfully")

	default:
		panic("unknown action: " + args[0])
	}
}

func handleVMs(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		vms, err := c.ListVMs(args[1])
		if err != nil {
			panic(err)
		}
		for _, vm := range vms {
			fmt.Println(vm)
		}

	case "create":
		if len(args) < 5 {
			panic("required: name image size namespace")
		}
		err := c.CreateVM(args[1], args[2], args[3], args[4])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		vm, err := c.GetVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", vm)

	case "wait":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.WaitVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM waited successfully")

	case "stop":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.StopVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM stopped successfully")

	case "start":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.StartVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM started successfully")

	case "restart":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.RestartVM(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("VM restarted successfully")
	default:
		panic("unknown action: " + args[0])
	}
}

func handleContainers(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		containers, err := c.ListContainers(args[1])
		if err != nil {
			panic(err)
		}
		for _, container := range containers {
			fmt.Printf("%+v\n", container)
		}

	case "create":
		if len(args) < 8 {
			panic("required: name image namespace cpu ram disk port")
		}
		cpu, _ := strconv.Atoi(args[4])
		ram, _ := strconv.Atoi(args[5])
		disk, _ := strconv.Atoi(args[6])
		port, _ := strconv.Atoi(args[7])
		err := c.CreateContainer(args[1], args[2], args[3], cpu, ram, disk, port)
		if err != nil {
			panic(err)
		}
		fmt.Println("container created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteContainer(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("container deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		container, err := c.GetContainer(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", container)

	default:
		panic("unknown action: " + args[0])
	}
}

func handleVolumes(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		volumes, err := c.ListVolumes(args[1])
		if err != nil {
			panic(err)
		}
		for _, volume := range volumes {
			fmt.Println(volume)
		}

	case "create":
		if len(args) < 4 {
			panic("required: name namespace size")
		}
		err := c.CreateVolume(args[1], args[2], args[3])
		if err != nil {
			panic(err)
		}
		fmt.Println("volume created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteVolume(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("volume deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		volume, err := c.GetVolume(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", volume)

	default:
		panic("unknown action: " + args[0])
	}
}

func handleNamespaces(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		namespaces, err := c.ListNamespaces()
		if err != nil {
			panic(err)
		}
		for _, ns := range namespaces {
			fmt.Println(ns)
		}

	case "create":
		if len(args) < 2 {
			panic("namespace name required")
		}
		err := c.CreateNamespace(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Println("namespace created successfully")

	case "delete":
		if len(args) < 2 {
			panic("namespace name required")
		}
		err := c.DeleteNamespace(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Println("namespace deleted successfully")

	case "get":
		if len(args) < 2 {
			panic("namespace name required")
		}
		ns, err := c.GetNamespace(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", ns)

	default:
		panic("unknown action: " + args[0])
	}
}

func handleDBs(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		dbs, err := c.ListDBs(args[1])
		if err != nil {
			panic(err)
		}
		for _, db := range dbs {
			fmt.Printf("%+v\n", db)
		}

	case "create":
		if len(args) < 5 {
			panic("required: name namespace type size")
		}
		err := c.CreateDB(args[1], args[2], args[3], args[4])
		if err != nil {
			panic(err)
		}
		fmt.Println("database created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteDB(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("database deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		db, err := c.GetDB(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", db)

	default:
		panic("unknown action: " + args[0])
	}
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
