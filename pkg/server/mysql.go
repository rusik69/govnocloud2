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

// MysqlManager handles mysql operations
type MysqlManager struct {
	kubectl KubectlRunner
}

// NewMysqlManager creates a new mysql manager
func NewMysqlManager() *MysqlManager {
	return &MysqlManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListMysqlHandler handles requests to list mysql
func ListMysqlHandler(c *gin.Context) {
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
	mysql, err := mysqlManager.ListClusters(namespace)
	if err != nil {
		log.Printf("failed to list mysql: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list mysql: %v", err))
		return
	}
	c.JSON(http.StatusOK, mysql)
}

// CreateMysqlHandler handles requests to create a new mysql
func CreateMysqlHandler(c *gin.Context) {
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

	var mysql types.Mysql
	if err := c.BindJSON(&mysql); err != nil {
		log.Printf("failed to bind JSON: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("failed to bind JSON: %v", err))
		return
	}

	mysql.Namespace = namespace
	if err := mysqlManager.CreateCluster(namespace, mysql); err != nil {
		log.Printf("failed to create mysql: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create mysql: %v", err))
		return
	}
	c.JSON(http.StatusCreated, mysql)
}

// GetMysqlHandler handles requests to get a mysql
func GetMysqlHandler(c *gin.Context) {
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
	mysql, err := mysqlManager.GetCluster(namespace, c.Param("name"))
	if err != nil {
		log.Printf("failed to get mysql: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get mysql: %v", err))
		return
	}
	c.JSON(http.StatusOK, mysql)
}

// DeleteMysqlHandler handles requests to delete a mysql
func DeleteMysqlHandler(c *gin.Context) {
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
	if err := mysqlManager.DeleteCluster(namespace, c.Param("name")); err != nil {
		log.Printf("failed to delete mysql: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete mysql: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateCluster creates a new mysql cluster
func (m *MysqlManager) CreateCluster(namespace string, mysql types.Mysql) error {
	manifestBody := fmt.Sprintf(`
	apiVersion: mysql.oracle.com/v1alpha1
	kind: MysqlCluster
	metadata:
	  name: %s
	spec:
	  instances: %d
	  routerInstances: %d
	`, mysql.Name, mysql.Instances, mysql.RouterInstances)
	tempFile, err := os.CreateTemp("", "mysql-manifest.yaml")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.WriteString(manifestBody); err != nil {
		return err
	}
	if _, err := m.kubectl.Run("apply", "-f", tempFile.Name(), "-n", namespace, "--wait=true", "--timeout=300s"); err != nil {
		return err
	}
	return nil
}

// GetCluster retrieves mysql cluster details
func (m *MysqlManager) GetCluster(namespace, name string) (*types.Mysql, error) {
	out, err := m.kubectl.Run("get", "InnoDBCluster", name, "-o", "json", "-n", namespace)
	if err != nil {
		return nil, err
	}
	var mysqlCluster types.MysqlCluster
	if err := json.Unmarshal([]byte(out), &mysqlCluster); err != nil {
		return nil, err
	}
	mysql := &types.Mysql{
		Name:            mysqlCluster.Metadata.Name,
		Namespace:       namespace,
		Instances:       mysqlCluster.Spec.Instances,
		RouterInstances: mysqlCluster.Spec.RouterInstances,
	}
	return mysql, nil
}

// DeleteCluster removes a mysql cluster
func (m *MysqlManager) DeleteCluster(namespace, name string) error {
	if out, err := m.kubectl.Run("delete", "InnoDBCluster", name, "-n", namespace, "--force", "--grace-period=0"); err != nil {
		return fmt.Errorf("failed to delete mysql cluster: %w\nOutput: %s", err, out)
	}
	return nil
}

// ListClusters lists all mysql clusters in a namespace
func (m *MysqlManager) ListClusters(namespace string) ([]types.Mysql, error) {
	out, err := m.kubectl.Run("get", "InnoDBCluster", "-o", "json", "-n", namespace)
	if err != nil {
		return nil, err
	}
	var mysqlClusterList types.MysqlClusterList
	if err := json.Unmarshal([]byte(out), &mysqlClusterList); err != nil {
		return nil, err
	}
	res := make([]types.Mysql, 0, len(mysqlClusterList.Items))
	for _, mysqlCluster := range mysqlClusterList.Items {
		res = append(res, types.Mysql{
			Name:            mysqlCluster.Metadata.Name,
			Namespace:       namespace,
			Instances:       mysqlCluster.Spec.Instances,
			RouterInstances: mysqlCluster.Spec.RouterInstances,
		})
	}
	return res, nil
}
