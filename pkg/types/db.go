package types

import (
	"fmt"
)

// DBFlavor represents a specific database flavor configuration
type DBFlavor struct {
	Version string `json:"version"`
	Port    int    `json:"port"`
}

// DBType represents a database type configuration
type DBType struct {
	Name        string               `json:"name"`
	BaseImage   string              `json:"baseImage"`
	DefaultPort int                 `json:"defaultPort"`
	Flavors     map[string]DBFlavor `json:"flavors"`
}

// DB represents a database instance configuration
type DB struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Flavor    string            `json:"flavor"`
	Port      int              `json:"port,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Resources ResourceConfig    `json:"resources,omitempty"`
}

// ResourceConfig represents resource requirements
type ResourceConfig struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Disk   string `json:"disk,omitempty"`
}

// DatabaseRegistry holds all supported database configurations
type DatabaseRegistry struct {
	types map[string]DBType
}

// NewDatabaseRegistry creates a new database registry with default configurations
func NewDatabaseRegistry() *DatabaseRegistry {
	return &DatabaseRegistry{
		types: map[string]DBType{
			"mysql": {
				Name:        "MySQL",
				BaseImage:   "mysql",
				DefaultPort: 3306,
				Flavors: map[string]DBFlavor{
					"5.7": {
						Version: "5.7",
						Port:    3306,
					},
					"8.0": {
						Version: "8.0",
						Port:    3306,
					},
				},
			},
			"postgres": {
				Name:        "PostgreSQL",
				BaseImage:   "postgres",
				DefaultPort: 5432,
				Flavors: map[string]DBFlavor{
					"9.6": {
						Version: "9.6",
						Port:    5432,
					},
					"13": {
						Version: "13",
						Port:    5432,
					},
					"14": {
						Version: "14",
						Port:    5432,
					},
				},
			},
		},
	}
}

// GetDBType returns the configuration for a specific database type
func (r *DatabaseRegistry) GetDBType(dbType string) (DBType, bool) {
	t, exists := r.types[dbType]
	return t, exists
}

// GetDBFlavor returns the configuration for a specific database flavor
func (r *DatabaseRegistry) GetDBFlavor(dbType, flavor string) (DBFlavor, bool) {
	t, exists := r.types[dbType]
	if !exists {
		return DBFlavor{}, false
	}
	f, exists := t.Flavors[flavor]
	return f, exists
}

// ValidateDB validates a database configuration
func (r *DatabaseRegistry) ValidateDB(db *DB) error {
	dbType, exists := r.GetDBType(db.Type)
	if !exists {
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}

	if db.Flavor != "" {
		_, exists := r.GetDBFlavor(db.Type, db.Flavor)
		if !exists {
			return fmt.Errorf("unsupported flavor %s for database type %s", db.Flavor, db.Type)
		}
	}

	if db.Port == 0 {
		db.Port = dbType.DefaultPort
	}

	return nil
}

// GetImageName returns the full image name for a database configuration
func (r *DatabaseRegistry) GetImageName(db *DB) (string, error) {
	dbType, exists := r.GetDBType(db.Type)
	if !exists {
		return "", fmt.Errorf("unsupported database type: %s", db.Type)
	}

	if db.Flavor == "" {
		return fmt.Sprintf("%s:latest", dbType.BaseImage), nil
	}

	flavor, exists := r.GetDBFlavor(db.Type, db.Flavor)
	if !exists {
		return "", fmt.Errorf("unsupported flavor %s for database type %s", db.Flavor, db.Type)
	}

	return fmt.Sprintf("%s:%s", dbType.BaseImage, flavor.Version), nil
}

// For backward compatibility
var DefaultRegistry = NewDatabaseRegistry()

var DBTypes = map[string]DBType{
	"mysql": {
		BaseImage:   "mysql",
		DefaultPort: 3306,
		Flavors: map[string]DBFlavor{
			"5.7": {
				Version: "5.7",
				Port:    3306,
			},
		},
	},
	"postgres": {
		BaseImage:   "postgres",
		DefaultPort: 5432,
		Flavors: map[string]DBFlavor{
			"9.6": {
				Version: "9.6",
				Port:    5432,
			},
		},
	},
}
