package types

// DB is a database.
type DB struct {
	// Name is the name of the database.
	Name string `json:"name"`
	// Type is the type of the database.
	Type string `json:"type"`
	// Flavor is the flavor of the database.
	Flavor string `json:"flavor"`
}

// DBType is a type of database.
type DBType struct {
	// Image is the image of the database type.
	Image string `json:"image"`
	// Port is the port of the database type.
	Port int `json:"port"`
}

// DBTypes is a map of database types.
var DBTypes = map[string]DBType{
	"mysql": DBType{
		Image: "mysql:5.7",
		Port:  3306,
	},
	"postgres": DBType{
		Image: "postgres:9.6",
		Port:  5432,
	},
}
