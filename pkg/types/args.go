package types

// Args represents the command line arguments.
type Args struct {
	// Command is the command to execute.
	Command string `json:"command"`
	// Master is the k8s master host.
	Master string `json:"master"`
	// Workers is the slice of workers.
	Workers []string `json:"workers"`
	// User is the user to connect to the remote host.
	User string `json:"user"`
	// Key is the path to the private key file.
	Key string `json:"key"`
}
