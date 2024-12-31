package types

// Node represents a node in the cluster
type Node struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	User       string `json:"user"`
	Key        string `json:"key"`
	Password   string `json:"password"`
	MasterHost string `json:"master_host"`
	Status     string `json:"status"`
	MacAddress string `json:"mac_address"`
}
