package server

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// NamespaceManager handles namespace operations
type NamespaceManager struct {
	kubectl            KubectlRunner
	reservedNamespaces []string
}

// CreateNamespace creates a new namespace
func (m *NamespaceManager) CreateNamespace(name string) error {
	_, err := m.kubectl.Run("create", "namespace", name)
	return err
}

// DeleteNamespace deletes a namespace
func (m *NamespaceManager) DeleteNamespace(name string) error {
	_, err := m.kubectl.Run("delete", "namespace", name)
	return err
}

// ListNamespaces lists all namespaces
func (m *NamespaceManager) ListNamespaces() ([]string, error) {
	namespaces, err := m.kubectl.Run("get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}
	ns := strings.Split(string(namespaces), " ")
	res := []string{}
	// check if namespace is reserved
	for _, n := range ns {
		if !slices.Contains(m.reservedNamespaces, n) {
			res = append(res, n)
		}
	}
	return res, nil
}

// GetNamespace gets details of a specific namespace
func (m *NamespaceManager) GetNamespace(name string) (types.Namespace, error) {
	namespace, err := m.kubectl.Run("get", "namespace", name, "-o", "jsonpath={.metadata.name}")
	if err != nil {
		return types.Namespace{}, err
	}
	ns := types.Namespace{Name: string(namespace)}
	return ns, nil
}

// NewNamespaceManager creates a new namespace manager
func NewNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		kubectl:            &DefaultKubectlRunner{},
		reservedNamespaces: []string{"default", "longhorn-system", "kube-system", "kube-public", "kube-node-lease", "cnpg-system", "clickhouse-system", "kubevirt-manager", "kubevirt", "kubernetes-dashboard", "monitoring", "mysql-operator"},
	}
}

// CreateNamespaceHandler creates a new namespace
func CreateNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Println("namespace name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	// check if namespace is reserved
	if slices.Contains(namespaceManager.reservedNamespaces, name) {
		log.Println("namespace is reserved")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	err := namespaceManager.CreateNamespace(name)
	if err != nil {
		log.Println("failed to create namespace", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "namespace created successfully"})
}

// DeleteNamespaceHandler deletes a namespace
func DeleteNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	// check if namespace is reserved
	if slices.Contains(namespaceManager.reservedNamespaces, name) {
		log.Println("namespace is reserved")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	err := namespaceManager.DeleteNamespace(name)
	if err != nil {
		log.Println("failed to delete namespace", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "namespace deleted successfully"})
}

// ListNamespacesHandler lists all namespaces
func ListNamespacesHandler(c *gin.Context) {
	namespaces, err := namespaceManager.ListNamespaces()
	if err != nil {
		log.Println("failed to list namespaces", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespaces listed successfully")
	c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
}

// GetNamespaceHandler gets details of a specific namespace
func GetNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Println("namespace name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	// check if namespace is reserved
	if slices.Contains(namespaceManager.reservedNamespaces, name) {
		log.Println("namespace is reserved")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	namespace, err := namespaceManager.GetNamespace(name)
	if err != nil {
		log.Println("failed to get namespace", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"namespace": namespace})
}
