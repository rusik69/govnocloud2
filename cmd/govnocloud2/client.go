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

		case "clickhouse":
			if len(args) < 2 {
				panic("clickhouse subcommand required [list|get|create|delete]")
			}
			handleClickhouse(c, args[1:])

		case "postgres":
			if len(args) < 2 {
				panic("postgres subcommand required [list|get|create|delete]")
			}
			handlePostgres(c, args[1:])

		case "mysql":
			if len(args) < 2 {
				panic("mysql subcommand required [list|get|create|delete]")
			}
			handleMysql(c, args[1:])

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

		case "users":
			if len(args) < 2 {
				panic("users subcommand required [list|get|create|delete|setpassword|addnamespace|removenamespace]")
			}
			handleUsers(c, args[1:])

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
		err := c.AddNode(args[1], args[2], args[3], args[4], args[5])
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

func handleClickhouse(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		clickhouses, err := c.ListClickhouse(args[1])
		if err != nil {
			panic(err)
		}
		for _, ch := range clickhouses {
			fmt.Printf("%+v\n", ch)
		}

	case "create":
		if len(args) < 4 {
			panic("required: name namespace replicas")
		}
		replicas, _ := strconv.Atoi(args[3])
		err := c.CreateClickhouse(args[1], args[2], replicas)
		if err != nil {
			panic(err)
		}
		fmt.Println("clickhouse created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteClickhouse(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("clickhouse deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		ch, err := c.GetClickhouse(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", ch)

	default:
		panic("unknown action: " + args[0])
	}
}

func handlePostgres(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		dbs, err := c.ListPostgres(args[1])
		if err != nil {
			panic(err)
		}
		for _, db := range dbs {
			fmt.Printf("%+v\n", db)
		}

	case "create":
		if len(args) < 6 {
			panic("required: name namespace size instances routerInstances")
		}
		instances, _ := strconv.Atoi(args[4])
		routerInstances, _ := strconv.Atoi(args[5])
		err := c.CreatePostgres(args[1], args[2], args[3], instances, routerInstances)
		if err != nil {
			panic(err)
		}
		fmt.Println("postgres created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeletePostgres(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("postgres deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		db, err := c.GetPostgres(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", db)

	default:
		panic("unknown action: " + args[0])
	}
}

func handleMysql(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		if len(args) < 2 {
			panic("namespace required")
		}
		dbs, err := c.ListMysql(args[1])
		if err != nil {
			panic(err)
		}
		for _, db := range dbs {
			fmt.Printf("%+v\n", db)
		}

	case "create":
		if len(args) < 5 {
			panic("required: name namespace instances routerInstances")
		}
		instances, _ := strconv.Atoi(args[3])
		routerInstances, _ := strconv.Atoi(args[4])
		err := c.CreateMysql(args[1], args[2], instances, routerInstances)
		if err != nil {
			panic(err)
		}
		fmt.Println("mysql created successfully")

	case "delete":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		err := c.DeleteMysql(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("mysql deleted successfully")

	case "get":
		if len(args) < 3 {
			panic("required: name namespace")
		}
		db, err := c.GetMysql(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", db)

	default:
		panic("unknown action: " + args[0])
	}
}

func handleUsers(c *client.Client, args []string) {
	switch args[0] {
	case "list":
		users, err := c.ListUsers()
		if err != nil {
			panic(err)
		}
		for _, user := range users {
			fmt.Printf("%+v\n", user)
		}

	case "create":
		if len(args) < 3 {
			panic("required: name password")
		}
		user := types.User{
			Name:     args[1],
			Password: args[2],
		}
		// Add optional namespaces if provided
		if len(args) > 3 {
			user.Namespaces = args[3:]
		}
		err := c.CreateUser(args[1], user)
		if err != nil {
			panic(err)
		}
		fmt.Println("user created successfully")

	case "delete":
		if len(args) < 2 {
			panic("user name required")
		}
		err := c.DeleteUser(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Println("user deleted successfully")

	case "get":
		if len(args) < 2 {
			panic("user name required")
		}
		user, err := c.GetUser(args[1])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", user)

	case "setpassword":
		if len(args) < 3 {
			panic("required: name password")
		}
		err := c.SetUserPassword(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("user password set successfully")

	case "addnamespace":
		if len(args) < 3 {
			panic("required: username namespace")
		}
		err := c.AddNamespaceToUser(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("namespace added to user successfully")

	case "removenamespace":
		if len(args) < 3 {
			panic("required: username namespace")
		}
		err := c.RemoveNamespaceFromUser(args[1], args[2])
		if err != nil {
			panic(err)
		}
		fmt.Println("namespace removed from user successfully")

	default:
		panic("unknown action: " + args[0])
	}
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
