package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
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
	ns := strings.Split(string(namespaces), " ")
	res := []string{}
	// check if namespace is reserved
	for _, n := range ns {
		if !types.ReservedNamespaces[n] {
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
		kubectl: &DefaultKubectlRunner{}}
}

// CreateNamespaceHandler creates a new namespace
func CreateNamespaceHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	// check if namespace is reserved
	if types.ReservedNamespaces[name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user is not an admin")
		return
	}
	err = namespaceManager.CreateNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "namespace created successfully"})
}

// DeleteNamespaceHandler deletes a namespace
func DeleteNamespaceHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "namespace name is required")
		return
	}
	// check if namespace is reserved
	if types.ReservedNamespaces[name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user is not an admin")
		return
	}
	err = namespaceManager.DeleteNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "namespace deleted successfully"})
}

// ListNamespacesHandler lists all namespaces
func ListNamespacesHandler(c *gin.Context) {
	auth, _, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	namespaces, err := namespaceManager.ListNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespaces listed successfully")
	c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
}

// GetNamespaceHandler gets details of a specific namespace
func GetNamespaceHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace name is required"})
		return
	}
	// check if namespace is reserved
	if types.ReservedNamespaces[name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is reserved"})
		return
	}
	if !CheckNamespaceAccess(username, name) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}
	namespace, err := namespaceManager.GetNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("namespace retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"namespace": namespace})
}
