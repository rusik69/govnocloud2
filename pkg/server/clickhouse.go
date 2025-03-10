package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// ClickhouseManager handles clickhouse operations
type ClickhouseManager struct {
	kubectl KubectlRunner
}

// NewClickhouseManager creates a new clickhouse manager instance
func NewClickhouseManager() *ClickhouseManager {
	return &ClickhouseManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListClickhouseHandler handles requests to list clickhouse
func ListClickhouseHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	clickhouse, err := clickhouseManager.ListClusters(namespace)
	if err != nil {
		log.Printf("failed to list clickhouse: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list clickhouse: %v", err))
		return
	}
	c.JSON(http.StatusOK, clickhouse)
}

// CreateClickhouseHandler handles requests to create a new clickhouse
func CreateClickhouseHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	cluster := types.Clickhouse{}
	if err := c.ShouldBindJSON(&cluster); err != nil {
		log.Printf("failed to bind clickhouse: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("failed to bind clickhouse: %v", err))
		return
	}
	err := clickhouseManager.CreateCluster(namespace, cluster)
	if err != nil {
		log.Printf("failed to create clickhouse: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create clickhouse: %v", err))
		return
	}
	c.Status(http.StatusOK)
}

// DeleteClickhouseHandler handles requests to delete a clickhouse
func DeleteClickhouseHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	err := clickhouseManager.DeleteCluster(namespace, c.Param("name"))
	if err != nil {
		log.Printf("failed to delete clickhouse: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete clickhouse: %v", err))
		return
	}
	c.Status(http.StatusOK)
}

// GetClickhouseHandler handles requests to get a clickhouse
func GetClickhouseHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	clickhouse, err := clickhouseManager.GetCluster(namespace, c.Param("name"))
	if err != nil {
		log.Printf("failed to get clickhouse: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get clickhouse: %v", err))
		return
	}
	c.JSON(http.StatusOK, clickhouse)
}

// CreateCluster creates a new clickhouse cluster
func (m *ClickhouseManager) CreateCluster(namespace string, cluster types.Clickhouse) error {
	manifest := fmt.Sprintf(`apiVersion: clickhouse.altinity.com/v1
kind: ClickHouseInstallation
metadata:
  name: %s
  namespace: %s
spec:
  clickhouse:
    shards: %d
    replicas: %d
`, cluster.Name, namespace, cluster.Shards, cluster.Replicas)
	log.Println(manifest)
	tmpFile, err := os.CreateTemp("", "clickhouse-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if out, err := m.kubectl.Run("apply", "-f", tmpFile.Name(), "-n", namespace, "--wait=true", "--timeout=300s"); err != nil {
		return fmt.Errorf("failed to create clickhouse cluster: %w %s", err, out)
	}
	return nil
}

// GetCluster retrieves clickhouse cluster details
func (m *ClickhouseManager) GetCluster(namespace, name string) (types.Clickhouse, error) {
	out, err := m.kubectl.Run("get", "ClickHouseInstallation", name, "-o", "json", "-n", namespace)
	if err != nil {
		return types.Clickhouse{}, fmt.Errorf("failed to get clickhouse cluster: %w %s", err, out)
	}
	cluster := types.ClickhouseInstallation{}
	if err := json.Unmarshal([]byte(out), &cluster); err != nil {
		return types.Clickhouse{}, fmt.Errorf("failed to unmarshal clickhouse cluster: %w", err)
	}
	return types.Clickhouse{
		Name:      cluster.Metadata.Name,
		Namespace: namespace,
		Shards:    cluster.Spec.Clickhouse.Shards,
		Replicas:  cluster.Spec.Clickhouse.Replicas,
	}, nil
}

// ListClusters lists all clickhouse clusters
func (m *ClickhouseManager) ListClusters(namespace string) ([]types.Clickhouse, error) {
	out, err := m.kubectl.Run("get", "ClickHouseInstallation", "-o", "json", "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get clickhouse clusters: %w", err)
	}
	clusterList := types.ClickhouseInstallationList{}
	if err := json.Unmarshal([]byte(out), &clusterList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clickhouse cluster list: %w", err)
	}
	res := []types.Clickhouse{}
	for _, cluster := range clusterList.Items {
		res = append(res, types.Clickhouse{
			Name:      cluster.Metadata.Name,
			Namespace: namespace,
			Shards:    cluster.Spec.Clickhouse.Shards,
			Replicas:  cluster.Spec.Clickhouse.Replicas,
		})
	}
	return res, nil
}

// DeleteCluster deletes a clickhouse cluster
func (m *ClickhouseManager) DeleteCluster(namespace, name string) error {
	if out, err := m.kubectl.Run("delete", "ClickHouseInstallation", name, "-n", namespace, "--wait=true", "--timeout=300s"); err != nil {
		return fmt.Errorf("failed to delete clickhouse cluster: %w %s", err, out)
	}
	return nil
}
