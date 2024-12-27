package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	corev1 "k8s.io/api/core/v1"
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
		log.Printf("failed to list databases: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list databases: %v", err)})
		return
	}
	c.JSON(http.StatusOK, dbs)
}

// CreateDBHandler handles requests to create a new database
func CreateDBHandler(c *gin.Context) {
	var db types.DB
	if err := c.BindJSON(&db); err != nil {
		log.Printf("invalid request: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	manager := NewDBManager()
	if err := manager.CreateDB(&db); err != nil {
		log.Printf("failed to create database: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create database: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database created successfully", "database": db})
}

// GetDBHandler handles requests to get database details
func GetDBHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Printf("database name is required")
		respondWithError(c, http.StatusBadRequest, "database name is required")
		return
	}

	manager := NewDBManager()
	db, err := manager.GetDB(name)
	if err != nil {
		log.Printf("failed to get database: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get database: %v", err))
		return
	}

	if db == nil {
		log.Printf("database not found")
		respondWithError(c, http.StatusNotFound, "database not found")
		return
	}

	c.JSON(http.StatusOK, db)
}

// DeleteDBHandler handles requests to delete a database
func DeleteDBHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Printf("database name is required")
		respondWithError(c, http.StatusBadRequest, "database name is required")
		return
	}

	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	manager := NewDBManager()
	if err := manager.DeleteDB(name, namespace); err != nil {
		log.Printf("failed to delete database: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete database: %v", err))
		return
	}

	log.Printf("database %s deleted successfully", name)
	c.JSON(http.StatusOK, gin.H{"message": "Database deleted successfully"})
}

// generatePodManifest generates a Pod manifest for the database
func (m *DBManager) generatePodManifest(db *types.DB) (string, error) {
	dbType, ok := types.DBTypes[db.Type]
	if !ok {
		return "", fmt.Errorf("failed to get db image: %s", db.Type)
	}
	dbSize, ok := types.DBSizes[db.Size]
	if !ok {
		return "", fmt.Errorf("failed to get db size: %s", db.Size)
	}
	dbImage := dbType.Image
	dbPort := dbType.Port
	pod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
  labels:
    app: %s
    type: database
    dbtype: %s
    dbsize: %s
spec:
  containers:
    - name: %s
      image: %s
      ports:
        - containerPort: %d
      resources:
        requests:
          cpu: %dm
          memory: %dMi
        limits:
          cpu: %dm
          memory: %dMi
`, db.Name, db.Name, db.Type, db.Size, db.Name, dbImage, dbPort, dbSize.CPU, dbSize.RAM, dbSize.CPU, dbSize.RAM)

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
			Name:      pod.Name,
			Type:      pod.Labels["dbtype"],
			Size:      pod.Labels["dbsize"],
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

	log.Println(pod)

	tmpFile, err := os.CreateTemp("", "db-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(pod), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if _, err := m.kubectl.Run("apply", "-f", tmpFile.Name(), "-n", db.Namespace, "--wait=true", "--timeout=300s"); err != nil {
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
		Name:      pod.Name,
		Type:      pod.Labels["dbtype"],
		Size:      pod.Labels["dbsize"],
		Namespace: pod.Namespace,
	}

	return db, nil
}

// DeleteDB removes a database
func (m *DBManager) DeleteDB(name, namespace string) error {
	if out, err := m.kubectl.Run("delete", "pod", name, "-n", namespace, "--force", "--grace-period=0"); err != nil {
		return fmt.Errorf("failed to delete database pod: %w\nOutput: %s", err, out)
	}
	return nil
}
