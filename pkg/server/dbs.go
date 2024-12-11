package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBManager handles database operations
type DBManager struct {
	kubectl KubectlRunner
}

// NewDBManager creates a new database manager instance
func NewDBManager() *DBManager {
	return &DBManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListDBsHandler handles requests to list databases
func ListDBsHandler(c *gin.Context) {
	manager := NewDBManager()
	dbs, err := manager.ListDBs()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list databases: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"databases": dbs})
}

// CreateDBHandler handles requests to create a new database
func CreateDBHandler(c *gin.Context) {
	var db types.DB
	if err := c.BindJSON(&db); err != nil {
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	manager := NewDBManager()
	if err := manager.CreateDB(&db); err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create database: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "Database created successfully", "database": db})
}

// GetDBHandler handles requests to get database details
func GetDBHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "database name is required")
		return
	}

	manager := NewDBManager()
	db, err := manager.GetDB(name)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get database: %v", err))
		return
	}

	if db == nil {
		respondWithError(c, http.StatusNotFound, "database not found")
		return
	}

	respondWithSuccess(c, gin.H{"database": db})
}

// DeleteDBHandler handles requests to delete a database
func DeleteDBHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "database name is required")
		return
	}

	manager := NewDBManager()
	if err := manager.DeleteDB(name); err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete database: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "Database deleted successfully"})
}

// generatePodManifest generates a Pod manifest for the database
func (m *DBManager) generatePodManifest(db *types.DB) (*corev1.Pod, error) {
	dbType, ok := types.DBTypes[db.Type]
	if !ok {
		return nil, fmt.Errorf("failed to get db image: %s", db.Type)
	}
	dbSize, ok := types.DBSizes[db.Size]
	if !ok {
		return nil, fmt.Errorf("failed to get db size: %s", db.Size)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: db.Name,
			Labels: map[string]string{
				"app":  db.Name,
				"type": "database",
				"dbtype": db.Type,
				"dbsize": db.Size,
				"db":   db.Type,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  db.Name,
					Image: dbType.Image,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: int32(dbType.Port),
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", dbSize.CPU)),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", dbSize.RAM)),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", dbSize.CPU)),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", dbSize.RAM)),
						},
					},
				},
			},
		},
	}

	return pod, nil
}

// ListDBs returns a list of databases
func (m *DBManager) ListDBs() ([]types.DB, error) {
	out, err := m.kubectl.Run("get", "pods", "-l", "type=database", "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	var podList corev1.PodList
	if err := json.Unmarshal([]byte(out), &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}

	dbs := make([]types.DB, 0, len(podList.Items))
	for _, pod := range podList.Items {
		db := types.DB{
			Name: pod.Name,
			Type: pod.Labels["dbtype"],
			Size: pod.Labels["dbsize"],
			Namespace: pod.Namespace,
		}
		dbs = append(dbs, db)
	}

	return dbs, nil
}

// CreateDB creates a new database
func (m *DBManager) CreateDB(db *types.DB) error {
	// Validate DB type exists
	if _, ok := types.DBTypes[db.Type]; !ok {
		return fmt.Errorf("invalid database type: %s", db.Type)
	}
	
	// Validate DB size exists
	if _, ok := types.DBSizes[db.Size]; !ok {
		return fmt.Errorf("invalid database size: %s", db.Size)
	}

	pod, err := m.generatePodManifest(db)
	if err != nil {
		return fmt.Errorf("failed to generate pod manifest: %w", err)
	}

	podJSON, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to marshal pod manifest: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "db-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), podJSON, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if _, err := m.kubectl.Run("apply", "-f", tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to create database pod: %w", err)
	}

	return nil
}

// GetDB retrieves database details
func (m *DBManager) GetDB(name string) (*types.DB, error) {
	out, err := m.kubectl.Run("get", "pod", name, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get database pod: %w", err)
	}

	var pod corev1.Pod
	if err := json.Unmarshal([]byte(out), &pod); err != nil {
		return nil, fmt.Errorf("failed to parse pod details: %w", err)
	}

	db := &types.DB{
		Name: pod.Name,
		Type: pod.Labels["dbtype"],
		Size: pod.Labels["dbsize"],
		Namespace: pod.Namespace,
	}

	return db, nil
}

// DeleteDB removes a database
func (m *DBManager) DeleteDB(name string) error {
	if _, err := m.kubectl.Run("delete", "pod", name); err != nil {
		return fmt.Errorf("failed to delete database pod: %w", err)
	}
	return nil
}
