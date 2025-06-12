package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	"llms":       initLLMHandler(),
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
	  govnocloud2 client help
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

		case "help":
			printHelp()

		default:
			handler, exists := handlers[args[0]]
			if !exists {
				handleError(fmt.Errorf("unknown action: %s", args[0]))
			}
			handler.Handle(c, args[1:])
		}
	},
}

// validateArgs performs comprehensive validation of command arguments
func validateArgs(args []string, required int) error {
	if len(args) < required {
		return fmt.Errorf("insufficient arguments: expected %d, got %d", required, len(args))
	}

	// Validate each argument
	for i, arg := range args {
		if arg == "" {
			return fmt.Errorf("argument %d cannot be empty", i+1)
		}
		// Check for potentially dangerous characters
		if strings.ContainsAny(arg, ";&|`$") {
			return fmt.Errorf("argument %d contains potentially dangerous characters", i+1)
		}
	}

	return nil
}

// validateResourceName checks if a resource name is valid
func validateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	// Only allow alphanumeric characters, hyphens, and underscores
	if !regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(name) {
		return fmt.Errorf("resource name can only contain alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// validateIP checks if an IP address is valid
func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// validateMAC checks if a MAC address is valid
func validateMAC(mac string) error {
	_, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}
	return nil
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
		// Validate node name
		if err := validateResourceName(args[0]); err != nil {
			return err
		}
		// Validate IP address
		if err := validateIP(args[1]); err != nil {
			return err
		}
		// Validate MAC address
		if err := validateMAC(args[2]); err != nil {
			return err
		}
		// Validate CPU and memory values
		cpu, err := parseInt(args[3])
		if err != nil {
			return err
		}
		if cpu < 1 {
			return fmt.Errorf("CPU count must be at least 1")
		}
		mem, err := parseInt(args[4])
		if err != nil {
			return err
		}
		if mem < 512 {
			return fmt.Errorf("memory must be at least 512MB")
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
		// Validate name
		if err := validateResourceName(args[0]); err != nil {
			return err
		}
		// Validate image
		if err := validateResourceName(args[1]); err != nil {
			return err
		}
		// Validate VM size
		if err := validateResourceName(args[2]); err != nil {
			return err
		}
		// Validate VM namespace
		if err := validateResourceName(args[3]); err != nil {
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

func initLLMHandler() CommandHandler {
	handler := NewBaseCommandHandler("llms")

	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		llms, err := c.ListLLMs(args[0])
		if err != nil {
			return err
		}
		for _, llm := range llms {
			fmt.Printf("%+v\n", llm)
		}
		return nil
	})

	handler.RegisterCommand("create", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 3); err != nil {
			return err
		}
		return c.CreateLLM(args[0], args[1], args[2])
	})

	handler.RegisterCommand("delete", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		return c.DeleteLLM(args[0], args[1])
	})

	handler.RegisterCommand("get", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 2); err != nil {
			return err
		}
		llm, err := c.GetLLM(args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(llm)
	})
	handler.RegisterCommand("list", func(c *client.Client, args []string) error {
		if err := validateArgs(args, 1); err != nil {
			return err
		}
		llms, err := c.ListLLMs(args[0])
		if err != nil {
			return err
		}
		for _, llm := range llms {
			fmt.Printf("%+v\n", llm)
		}
		return nil
	})

	return handler
}

// printHelp displays the help information with all available actions
func printHelp() {
	fmt.Println("govnocloud2 client - Command-line interface for managing GovnoCloud resources")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  govnocloud2 client [resource] [action] [args...]")
	fmt.Println()
	fmt.Println("Available Resources and Actions:")
	fmt.Println()

	fmt.Println("  nodes:")
	fmt.Println("    list                           - List all nodes")
	fmt.Println("    get <name>                     - Get node details")
	fmt.Println("    add <name> <ip> <mac> <cpu> <mem> - Add a new node")
	fmt.Println("    delete <name>                  - Delete a node")
	fmt.Println("    restart <name>                 - Restart a node")
	fmt.Println()

	fmt.Println("  vms:")
	fmt.Println("    list <namespace>               - List VMs in namespace")
	fmt.Println("    create <namespace> <name> <cpu> <mem> - Create a new VM")
	fmt.Println("    get <namespace> <name>         - Get VM details")
	fmt.Println("    delete <namespace> <name>      - Delete a VM")
	fmt.Println("    start <namespace> <name>       - Start a VM")
	fmt.Println("    stop <namespace> <name>        - Stop a VM")
	fmt.Println("    restart <namespace> <name>     - Restart a VM")
	fmt.Println("    wait <namespace> <name>        - Wait for VM to be ready")
	fmt.Println()

	fmt.Println("  containers:")
	fmt.Println("    list <namespace>               - List containers in namespace")
	fmt.Println("    create <namespace> <name> <image> <cpu> <ram> <disk> <port> - Create a container")
	fmt.Println("    get <namespace> <name>         - Get container details")
	fmt.Println("    delete <namespace> <name>      - Delete a container")
	fmt.Println()

	fmt.Println("  volumes:")
	fmt.Println("    list <namespace>               - List volumes in namespace")
	fmt.Println("    create <namespace> <name> <size> - Create a volume")
	fmt.Println("    get <namespace> <name>         - Get volume details")
	fmt.Println("    delete <namespace> <name>      - Delete a volume")
	fmt.Println()

	fmt.Println("  namespaces:")
	fmt.Println("    list                           - List all namespaces")
	fmt.Println("    create <name>                  - Create a namespace")
	fmt.Println("    get <name>                     - Get namespace details")
	fmt.Println("    delete <name>                  - Delete a namespace")
	fmt.Println()

	fmt.Println("  clickhouse:")
	fmt.Println("    list <namespace>               - List ClickHouse instances in namespace")
	fmt.Println("    create <namespace> <name> <replicas> - Create a ClickHouse instance")
	fmt.Println("    get <namespace> <name>         - Get ClickHouse instance details")
	fmt.Println("    delete <namespace> <name>      - Delete a ClickHouse instance")
	fmt.Println()

	fmt.Println("  postgres:")
	fmt.Println("    list <namespace>               - List PostgreSQL instances in namespace")
	fmt.Println("    create <namespace> <name> <version> <instances> <routers> - Create a PostgreSQL instance")
	fmt.Println("    get <namespace> <name>         - Get PostgreSQL instance details")
	fmt.Println("    delete <namespace> <name>      - Delete a PostgreSQL instance")
	fmt.Println()

	fmt.Println("  mysql:")
	fmt.Println("    list <namespace>               - List MySQL instances in namespace")
	fmt.Println("    create <namespace> <name> <instances> <routers> - Create a MySQL instance")
	fmt.Println("    get <namespace> <name>         - Get MySQL instance details")
	fmt.Println("    delete <namespace> <name>      - Delete a MySQL instance")
	fmt.Println()

	fmt.Println("  llms:")
	fmt.Println("    list <namespace>               - List LLM instances in namespace")
	fmt.Println("    create <name> <namespace> <type> - Create an LLM instance")
	fmt.Println("    get <name> <namespace>         - Get LLM instance details")
	fmt.Println("    delete <namespace> <name>      - Delete an LLM instance")
	fmt.Println()

	fmt.Println("  users:")
	fmt.Println("    list                           - List all users")
	fmt.Println("    create <name> <password> [namespaces...] - Create a user")
	fmt.Println("    get <name>                     - Get user details")
	fmt.Println("    delete <name>                  - Delete a user")
	fmt.Println("    setpassword <name> <password>  - Set user password")
	fmt.Println("    addnamespace <name> <namespace> - Add namespace to user")
	fmt.Println("    removenamespace <name> <namespace> - Remove namespace from user")
	fmt.Println()

	fmt.Println("  Other Commands:")
	fmt.Println("    version                        - Get server version")
	fmt.Println("    help                           - Show this help message")
	fmt.Println()

	fmt.Println("Examples:")
	fmt.Println("  govnocloud2 client nodes list")
	fmt.Println("  govnocloud2 client vms create mynamespace myvm 2 4096")
	fmt.Println("  govnocloud2 client namespaces create mynamespace")
	fmt.Println("  govnocloud2 client users create myuser mypassword mynamespace")
	fmt.Println("  govnocloud2 client llms create myllm mynamespace deepseek-r1-1.5b")
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
