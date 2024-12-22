package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// NamespaceManager handles namespace operations
type NamespaceManager struct {
	kubectl KubectlRunner
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
	return strings.Split(string(namespaces), " "), nil
}

// GetNamespace gets details of a specific namespace
func (m *NamespaceManager) GetNamespace(name string) (string, error) {
	namespace, err := m.kubectl.Run("get", "namespace", name, "-o", "yaml")
	if err != nil {
		return "", err
	}
	return string(namespace), nil
}

// NewNamespaceManager creates a new namespace manager
func NewNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// CreateNamespaceHandler creates a new namespace
func CreateNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	manager := NewNamespaceManager()
	err := manager.CreateNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "namespace created successfully"})
}

// DeleteNamespaceHandler deletes a namespace
func DeleteNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	manager := NewNamespaceManager()
	err := manager.DeleteNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "namespace deleted successfully"})
}

// ListNamespacesHandler lists all namespaces
func ListNamespacesHandler(c *gin.Context) {
	manager := NewNamespaceManager()
	namespaces, err := manager.ListNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
}

// GetNamespaceHandler gets details of a specific namespace
func GetNamespaceHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	manager := NewNamespaceManager()
	namespace, err := manager.GetNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"namespace": namespace})
}
