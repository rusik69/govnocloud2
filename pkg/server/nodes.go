package server

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListNodesHandler lists nodes
func ListNodesHandler(c *gin.Context) {
	nodes, err := ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// ListNodes lists nodes
func ListNodes() ([]string, error) {
	out, err := exec.Command("kubectl", "get", "nodes", "-o", "jsonpath={.items[*].metadata.name}").Output()
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

// AddNodeHandler adds a node
func AddNodeHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// GetNodeHandler gets a node
func GetNodeHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// DeleteNodeHandler deletes a node
func DeleteNodeHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
