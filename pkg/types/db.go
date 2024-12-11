package types

// DB is a database.
type DB struct {
	// Name is the name of the database.
	Name string `json:"name"`
	// Namespace is the namespace of the database.
	Namespace string `json:"namespace"`
	// Type is the type of the database.
	Type string `json:"type"`
	// Size is the size of the database.
	Size string `json:"size"`
	// Volume is the volume of the database.
	Volume string `json:"volume"`
}

// DBType is a database type.
type DBType struct {
	// Name is the name of the database type.
	Name string `json:"name"`
	// Image is the image of the database type.
	Image string `json:"image"`
	// MountPath is the mount path of the database volume.
	MountPath string `json:"mountPath"`
	// Port is the port of the database.
	Port int `json:"port"`
}

// DBSize is a database size.
type DBSize struct {
	// Name is the name of the database size.
	Name string `json:"name"`
	// RAM is the RAM of the database size.
	RAM int `json:"ram"`
	// CPU is the CPU of the database size.
	CPU int `json:"cpu"`
	// Disk is the disk of the database size.
	Disk int `json:"disk"`
}

// DBVolume is a database volume.
type DBVolume struct {
	// Name is the name of the database volume.
	Name string `json:"name"`
	// Size is the size of the database volume.
	Size int `json:"size"`
}

// DBTypes is a map of database types.
var DBTypes = map[string]DBType{
	"mysql": DBType{
		Name:      "mysql",
		Image:     "mysql:8.0",
		MountPath: "/var/lib/mysql",
		Port:      3306,
	},
	"postgres": DBType{
		Name:      "postgres",
		Image:     "postgres:15",
		MountPath: "/var/lib/postgresql/data",
		Port:      5432,
	},
}

// DBSizes is a map of database sizes.
var DBSizes = map[string]DBSize{
	"small": DBSize{
		Name: "small",
		RAM:  1024,
		CPU:  1,
		Disk: 10,
	},
	"medium": DBSize{
		Name: "medium",
		RAM:  2048,
		CPU:  2,
		Disk: 20,
	},
	"large": DBSize{
		Name: "large",
		RAM:  4096,
		CPU:  4,
		Disk: 40,
	},
}
