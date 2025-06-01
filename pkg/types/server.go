package types

type ServerConfig struct {
	Host         string
	Port         string
	SSHUser      string
	SSHPassword  string
	Key          string
	MasterHost   string
	RootPassword string
}
