package types

// Mysql is a mysql cluster
type Mysql struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	Instances       int    `json:"replicas"`
	RouterInstances int    `json:"router_replicas"`
}

// MysqlCluster is a mysql cluster
type MysqlCluster struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Instances       int `json:"instances"`
		RouterInstances int `json:"routerInstances"`
	} `json:"spec"`
}

// MysqlClusterList is a list of mysql clusters
type MysqlClusterList struct {
	Items []MysqlCluster `json:"items"`
}
