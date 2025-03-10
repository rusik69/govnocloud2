package types

// Clickhouse represents a Clickhouse cluster
type Clickhouse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int    `json:"replicas"`
	Shards    int    `json:"shards"`
}

// ClickhouseInstallation is a Clickhouse installation
type ClickhouseInstallation struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Clickhouse struct {
			Shards   int `json:"shards"`
			Replicas int `json:"replicas"`
		} `json:"clickhouse"`
	} `json:"spec"`
}

// ClickhouseInstallationList is a list of Clickhouse installations
type ClickhouseInstallationList struct {
	Items []ClickhouseInstallation `json:"items"`
}
