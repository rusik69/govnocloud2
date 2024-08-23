package args

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// Parse parses the command line arguments.
func Parse() (types.Args, error) {
	arg := types.Args{}
	if len(os.Args) < 2 {
		return arg, errors.New("command is required")
	}
	switch os.Args[1] {
	case "install":
		arg.Command = "install"
	case "uninstall":
		arg.Command = "uninstall"
	default:
		return arg, errors.New("unknown command")
	}
	masterString := flag.String("master", "", "master host")
	workersString := flag.String("workers", "", "comma separated list of workers")
	user := flag.String("user", "ubuntu", "user to connect to the remote host")
	key := flag.String("key", "~/.ssh/id_rsa", "path to the private key file")
	flag.Parse()
	if *masterString != "" {
		arg.Master = *masterString
	} else {
		return arg, errors.New("master is required")
	}
	if *workersString != "" {
		split := strings.Split(*workersString, ",")
		if len(split) > 0 {
			arg.Workers = split
		} else {
			return arg, errors.New("failed to parse workers")
		}
	}
	arg.User = *user
	arg.Key = *key
	return arg, nil
}
