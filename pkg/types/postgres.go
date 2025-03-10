package types

// Postgres is a postgres.
type Postgres struct {
	// Name is the name of the postgres.
	Name string `json:"name"`
	// Namespace is the namespace of the postgres.
	Namespace string `json:"namespace"`
	// Size is the size of the postgres.
	Size string `json:"size"`
	// Replicas is the number of replicas of the postgres.
	Replicas int `json:"replicas"`
	// Storage is the storage of the postgres.
	Storage int `json:"storage"`
}

// PostgresSize is a postgres size.
type PostgresSize struct {
	// RAM is the RAM of the postgres size.
	RAM int `json:"ram"`
	// CPU is the CPU of the postgres size.
	CPU int `json:"cpu"`
}

// PostgresSizes is a map of postgres sizes.
var PostgresSizes = map[string]PostgresSize{
	"small": PostgresSize{
		RAM: 1024,
		CPU: 1,
	},
	"medium": PostgresSize{
		RAM: 2048,
		CPU: 2,
	},
	"large": PostgresSize{
		RAM: 4096,
		CPU: 4,
	},
}

// PostgresCluster is a postgres cluster
type PostgresCluster struct {
	Metadata struct {
		Name   string `json:"name"`
		Labels struct {
			Size string `json:"size"`
		} `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		Instances int `json:"instances"`
		Storage   struct {
			Size string `json:"size"`
		} `json:"storage"`
	} `json:"spec"`
}

// PostgresClusterList is a list of postgres clusters	.
type PostgresClusterList struct {
	Items []PostgresCluster `json:"items"`
}
