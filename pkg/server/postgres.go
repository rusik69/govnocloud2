package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// PostgresManager handles postgres operations
type PostgresManager struct {
	kubectl KubectlRunner
}

// NewPostgresManager creates a new postgres manager instance
func NewPostgresManager() *PostgresManager {
	return &PostgresManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListPostgresHandler handles requests to list postgres
func ListPostgresHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	// Check if namespace is reserved
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	postgres, err := postgresManager.ListClusters(namespace)
	if err != nil {
		log.Printf("failed to list databases: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list databases: %v", err)})
		return
	}
	c.JSON(http.StatusOK, postgres)
}

// CreatePostgresHandler handles requests to create a new postgres
func CreatePostgresHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	// Check if namespace is reserved
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	var postgres types.Postgres
	if err := c.BindJSON(&postgres); err != nil {
		log.Printf("invalid request: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if err := postgresManager.CreateCluster(&postgres); err != nil {
		log.Printf("failed to create postgres: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create postgres: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Postgres created successfully", "postgres": postgres})
}

// GetPostgresHandler handles requests to get postgres details
func GetPostgresHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Printf("postgres name is required")
		respondWithError(c, http.StatusBadRequest, "postgres name is required")
		return
	}

	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	// Check if namespace is reserved
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	postgres, err := postgresManager.GetCluster(name, namespace)
	if err != nil {
		log.Printf("failed to get postgres: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get postgres: %v", err))
		return
	}

	if postgres == nil {
		log.Printf("postgres not found")
		respondWithError(c, http.StatusNotFound, "postgres not found")
		return
	}

	c.JSON(http.StatusOK, postgres)
}

// DeletePostgresHandler handles requests to delete a postgres
func DeletePostgresHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Printf("postgres name is required")
		respondWithError(c, http.StatusBadRequest, "postgres name is required")
		return
	}

	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	// Check if namespace is reserved
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	if err := postgresManager.DeleteCluster(name, namespace); err != nil {
		log.Printf("failed to delete postgres: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete postgres: %v", err))
		return
	}

	log.Printf("postgres %s deleted successfully", name)
	c.JSON(http.StatusOK, gin.H{"message": "Postgres deleted successfully"})
}

// generateManifest generates a Pod manifest for the postgres
func (m *PostgresManager) generateManifest(postgres *types.Postgres) (string, error) {
	size, ok := types.PostgresSizes[postgres.Size]
	if !ok {
		return "", fmt.Errorf("failed to get postgres size: %s", postgres.Size)
	}
	pod := fmt.Sprintf(`apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: %s
  labels:
    size: %s
spec:
  instances: %d
  resources:
    requests:
      memory: %dMi
      cpu: %d
    limits:
      memory: %dMi
      cpu: %d
  storage:
    size: %dGi
`, postgres.Name, postgres.Size, postgres.Replicas, size.RAM, size.CPU, size.RAM, size.CPU, postgres.Storage)

	return pod, nil
}

// ListClusters returns a list of postgres clusters
func (m *PostgresManager) ListClusters(namespace string) ([]types.Postgres, error) {
	out, err := m.kubectl.Run("get", "clusters", "-o", "json", "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list postgres: %w", err)
	}

	var clusterList types.PostgresClusterList
	if err := json.Unmarshal([]byte(out), &clusterList); err != nil {
		return nil, fmt.Errorf("failed to parse postgres cluster list: %w", err)
	}
	postgresClusters := make([]types.Postgres, 0, len(clusterList.Items))
	for _, cluster := range clusterList.Items {
		// cut Gi from the end of the storage size and convert to int
		storageSize, err := strconv.Atoi(strings.TrimSuffix(cluster.Spec.Storage.Size, "Gi"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse storage size: %w", err)
		}
		postgresClusters = append(postgresClusters, types.Postgres{
			Name:      cluster.Metadata.Name,
			Size:      cluster.Metadata.Labels.Size,
			Namespace: namespace,
			Replicas:  cluster.Spec.Instances,
			Storage:   storageSize,
		})
	}

	return postgresClusters, nil
}

// CreateCluster creates a new postgres cluster
func (m *PostgresManager) CreateCluster(postgres *types.Postgres) error {
	// Validate DB size exists
	if _, ok := types.PostgresSizes[postgres.Size]; !ok {
		return fmt.Errorf("invalid database size: %s", postgres.Size)
	}

	cluster, err := m.generateManifest(postgres)
	if err != nil {
		return fmt.Errorf("failed to generate pod manifest: %w", err)
	}
	log.Println(cluster)
	tmpFile, err := os.CreateTemp("", "db-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(cluster), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if _, err := m.kubectl.Run("apply", "-f", tmpFile.Name(), "-n", postgres.Namespace, "--wait=true", "--timeout=300s"); err != nil {
		return fmt.Errorf("failed to create database pod: %w", err)
	}

	return nil
}

// GetCluster retrieves postgres cluster details
func (m *PostgresManager) GetCluster(name, namespace string) (*types.Postgres, error) {
	out, err := m.kubectl.Run("get", "cluster", name, "-o", "json", "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres cluster: %w", err)
	}

	var cluster types.PostgresCluster
	if err := json.Unmarshal([]byte(out), &cluster); err != nil {
		return nil, fmt.Errorf("failed to parse cluster details: %w", err)
	}

	storageSize, err := strconv.Atoi(strings.TrimSuffix(cluster.Spec.Storage.Size, "Gi"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse storage size: %w", err)
	}

	postgres := &types.Postgres{
		Name:      cluster.Metadata.Name,
		Size:      cluster.Metadata.Labels.Size,
		Namespace: namespace,
		Replicas:  cluster.Spec.Instances,
		Storage:   storageSize,
	}

	return postgres, nil
}

// DeleteCluster removes a postgres cluster
func (m *PostgresManager) DeleteCluster(name, namespace string) error {
	if out, err := m.kubectl.Run("delete", "Cluster", name, "-n", namespace, "--force", "--grace-period=0"); err != nil {
		return fmt.Errorf("failed to delete postgres cluster: %w\nOutput: %s", err, out)
	}
	return nil
}
