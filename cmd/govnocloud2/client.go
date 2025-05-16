package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/rusik69/govnocloud2/pkg/client"
	"github.com/rusik69/govnocloud2/pkg/types"
	"github.com/spf13/cobra"
)

// CommandHandler defines the interface for resource command handlers
type CommandHandler interface {
	Handle(c *client.Client, args []string)
}

// BaseCommandHandler provides common functionality for all command handlers
type BaseCommandHandler struct {
	ResourceName string
	Commands     map[string]CommandFunc
}

// CommandFunc defines the function signature for command handlers
type CommandFunc func(c *client.Client, args []string) error

// NewBaseCommandHandler creates a new base command handler
func NewBaseCommandHandler(resourceName string) *BaseCommandHandler {
	return &BaseCommandHandler{
		ResourceName: resourceName,
		Commands:     make(map[string]CommandFunc),
	}
}

// RegisterCommand registers a new command handler
func (h *BaseCommandHandler) RegisterCommand(name string, handler CommandFunc) {
	h.Commands[name] = handler
}

// Handle processes the command
func (h *BaseCommandHandler) Handle(c *client.Client, args []string) {
	if len(args) == 0 {
		handleError(fmt.Errorf("%s subcommand required", h.ResourceName))
	}

	cmd := args[0]
	handler, exists := h.Commands[cmd]
	if !exists {
		handleError(fmt.Errorf("unknown action: %s", cmd))
	}

	if err := handler(c, args[1:]); err != nil {
		handleError(err)
	}
}

var handlers = map[string]CommandHandler{
	"nodes":      initNodeHandler(),
	"vms":        initVMHandler(),
	"containers": initContainerHandler(),
	"clickhouse": initClickhouseHandler(),
	"postgres":   initPostgresHandler(),
	"mysql":      initMysqlHandler(),
	"volumes":    initVolumeHandler(),
	"namespaces": initNamespaceHandler(),
	"users":      initUserHandler(),
}

// client command
var clientCmd = &cobra.Command{
	Use:   "client [action] [args]",
	Short: "govnocloud2 client",
	Long: `govnocloud2 client is a command-line interface for managing GovnoCloud resources.
	
	Examples:
	  govnocloud2 client nodes list
	  govnocloud2 client vms create myvm ubuntu 2 4 20
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			handleError(fmt.Errorf("action is required"))
		}

		c := client.NewClient(cfg.Client.Host, cfg.Client.Port, cfg.Client.User, cfg.Client.Password)

		switch args[0] {
		case "version":
			serverVer, err := client.GetServerVersion(cfg.Client.Host, cfg.Client.Port)
			if err != nil {
				handleError(err)
			}
			fmt.Println("server version:", serverVer)

		default:
			handler, exists := handlers[args[0]]
			if !exists {
				handleError(fmt.Errorf("unknown action: %s", args[0]))
			}
			handler.Handle(c, args[1:])
		}
	},
}

func initNodeHandler() CommandHandler {
	handler := NewBaseCommandHandler("nodes")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		nodes, err := c.ListNodes()
		if err != nil {
			return err
		}
		for _, node := range nodes {
			fmt.Println(node)
		}
		return nil
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		node, err := c.GetNode(args[0])
		if err != nil {
			return err
		}
		return printJSON(node)
	})

	handler.RegisterCommand("add", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 5); err != nil {
			return err
		}
		return c.AddNode(args[0], args[1], args[2], args[3], args[4])
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		return c.DeleteNode(args[0])
	})

	handler.RegisterCommand("restart", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		return c.RestartNode(args[0])
	})

	return handler
}

func initVMHandler() CommandHandler {
	handler := NewBaseCommandHandler("vms")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		vms, err := c.ListVMs(args[0])
		if err != nil {
			return err
		}
		for _, vm := range vms {
			fmt.Println(vm)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 4); err != nil {
			return err
		}
		return c.CreateVM(args[0], args[1], args[2], args[3])
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteVM(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		vm, err := c.GetVM(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(vm)
	})

	handler.RegisterCommand("wait", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.WaitVM(args[0], args[1])
	})

	handler.RegisterCommand("stop", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.StopVM(args[0], args[1])
	})

	handler.RegisterCommand("start", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.StartVM(args[0], args[1])
	})

	handler.RegisterCommand("restart", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.RestartVM(args[0], args[1])
	})

	return handler
}

func initContainerHandler() CommandHandler {
	handler := NewBaseCommandHandler("containers")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		containers, err := c.ListContainers(args[0])
		if err != nil {
			return err
		}
		for _, container := range containers {
			fmt.Printf("%+v\n", container)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 7); err != nil {
			return err
		}
		cpu, err := parseInt(args[3])
		if err != nil {
			return err
		}
		ram, err := parseInt(args[4])
		if err != nil {
			return err
		}
		disk, err := parseInt(args[5])
		if err != nil {
			return err
		}
		port, err := parseInt(args[6])
		if err != nil {
			return err
		}
		return c.CreateContainer(args[0], args[1], args[2], cpu, ram, disk, port)
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteContainer(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		container, err := c.GetContainer(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(container)
	})

	return handler
}

func initVolumeHandler() CommandHandler {
	handler := NewBaseCommandHandler("volumes")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		volumes, err := c.ListVolumes(args[0])
		if err != nil {
			return err
		}
		for _, volume := range volumes {
			fmt.Println(volume)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 3); err != nil {
			return err
		}
		return c.CreateVolume(args[0], args[1], args[2])
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteVolume(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		volume, err := c.GetVolume(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(volume)
	})

	return handler
}

func initNamespaceHandler() CommandHandler {
	handler := NewBaseCommandHandler("namespaces")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		namespaces, err := c.ListNamespaces()
		if err != nil {
			return err
		}
		for _, ns := range namespaces {
			fmt.Println(ns)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		return c.CreateNamespace(args[0])
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		return c.DeleteNamespace(args[0])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		ns, err := c.GetNamespace(args[0])
		if err != nil {
			return err
		}
		return printJSON(ns)
	})

	return handler
}

func initClickhouseHandler() CommandHandler {
	handler := NewBaseCommandHandler("clickhouse")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		clickhouses, err := c.ListClickhouse(args[0])
		if err != nil {
			return err
		}
		for _, ch := range clickhouses {
			fmt.Printf("%+v\n", ch)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 3); err != nil {
			return err
		}
		replicas, err := parseInt(args[2])
		if err != nil {
			return err
		}
		return c.CreateClickhouse(args[0], args[1], replicas)
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteClickhouse(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		ch, err := c.GetClickhouse(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(ch)
	})

	return handler
}

func initPostgresHandler() CommandHandler {
	handler := NewBaseCommandHandler("postgres")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		dbs, err := c.ListPostgres(args[0])
		if err != nil {
			return err
		}
		for _, db := range dbs {
			fmt.Printf("%+v\n", db)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 5); err != nil {
			return err
		}
		instances, err := parseInt(args[3])
		if err != nil {
			return err
		}
		routerInstances, err := parseInt(args[4])
		if err != nil {
			return err
		}
		return c.CreatePostgres(args[0], args[1], args[2], instances, routerInstances)
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeletePostgres(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		db, err := c.GetPostgres(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(db)
	})

	return handler
}

func initMysqlHandler() CommandHandler {
	handler := NewBaseCommandHandler("mysql")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		dbs, err := c.ListMysql(args[0])
		if err != nil {
			return err
		}
		for _, db := range dbs {
			fmt.Printf("%+v\n", db)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 4); err != nil {
			return err
		}
		instances, err := parseInt(args[2])
		if err != nil {
			return err
		}
		routerInstances, err := parseInt(args[3])
		if err != nil {
			return err
		}
		return c.CreateMysql(args[0], args[1], instances, routerInstances)
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteMysql(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		db, err := c.GetMysql(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(db)
	})

	return handler
}

func initUserHandler() CommandHandler {
	handler := NewBaseCommandHandler("users")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		users, err := c.ListUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			fmt.Printf("%+v\n", user)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		user := types.User{
			Name:     args[0],
			Password: args[1],
		}
		// Add optional namespaces if provided
		if len(args) > 2 {
			user.Namespaces = args[2:]
		}
		return c.CreateUser(args[0], user)
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		return c.DeleteUser(args[0])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		user, err := c.GetUser(args[0])
		if err != nil {
			return err
		}
		return printJSON(user)
	})

	handler.RegisterCommand("setpassword", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.SetUserPassword(args[0], args[1])
	})

	handler.RegisterCommand("addnamespace", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.AddNamespaceToUser(args[0], args[1])
	})

	handler.RegisterCommand("removenamespace", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.RemoveNamespaceFromUser(args[0], args[1])
	})

	return handler
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func validateArgs(args []string, required int) error {
	if len(args) < required {
		return fmt.Errorf("insufficient arguments: expected %d, got %d", required, len(args))
	}
	return nil
}

func parseInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", s)
	}
	return val, nil
}

func printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
