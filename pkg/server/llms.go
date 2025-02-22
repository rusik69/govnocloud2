package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// LLMManager handles LLM operations
type LLMManager struct {
	kubectl KubectlRunner
}

var llmManager *LLMManager

// NewLLMManager creates a new LLM manager instance
func NewLLMManager() *LLMManager {
	return &LLMManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// CreateLLMHandler handles LLM creation requests
func CreateLLMHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}

	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	var llm types.LLM
	if err := c.BindJSON(&llm); err != nil {
		log.Printf("invalid request body: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if llm.Type == "" {
		log.Printf("type is required")
		respondWithError(c, http.StatusBadRequest, "type is required")
		return
	}

	if _, ok := types.LLMTypes[llm.Type]; !ok {
		log.Printf("invalid type: %s", llm.Type)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid type: %s", llm.Type))
		return
	}

	llm.Name = name
	llm.Namespace = namespace

	if err := llmManager.CreateLLM(llm); err != nil {
		log.Printf("failed to create LLM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create LLM: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "LLM created successfully"})
}

// GetLLMHandler handles LLM retrieval requests
func GetLLMHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}

	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	llm, err := llmManager.GetLLM(namespace, name)
	if err != nil {
		log.Printf("failed to get LLM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get LLM: %v", err))
		return
	}

	respondWithSuccess(c, llm)
}

// DeleteLLMHandler handles LLM deletion requests
func DeleteLLMHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}

	if _, ok := types.ReservedNamespaces[namespace]; ok {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	if err := llmManager.DeleteLLM(namespace, name); err != nil {
		log.Printf("failed to delete LLM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete LLM: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "LLM deleted successfully"})
}

// CreateLLM creates a new LLM deployment
func (m *LLMManager) CreateLLM(llm types.LLM) error {
	llmType := types.LLMTypes[llm.Type]
	llmConfig := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: llm
        image: %s
        resources:
          requests:
            memory: "%dGi"
            cpu: "%d"
          limits:
            memory: "%dGi"
            cpu: "%d"`,
		llm.Name, llm.Namespace, llm.Name, llm.Name,
		llmType.Image, llmType.Memory, llmType.CPU, llmType.Memory, llmType.CPU)

	tmpfile, err := os.CreateTemp("", "llm-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(llmConfig); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	out, err := m.kubectl.Run("apply", "-f", tmpfile.Name())
	if err != nil {
		return fmt.Errorf("failed to create LLM %s: %s: %w", llm.Name, out, err)
	}

	return nil
}

// GetLLM retrieves an LLM deployment
func (m *LLMManager) GetLLM(namespace, name string) (*types.LLM, error) {
	out, err := m.kubectl.Run("get", "deployment", name, "-n", namespace, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM %s: %s: %w", name, out, err)
	}

	// Parse the JSON output and create an LLM object
	llm := &types.LLM{}
	// Add parsing logic here based on your needs

	return llm, nil
}

// DeleteLLM deletes an LLM deployment
func (m *LLMManager) DeleteLLM(namespace, name string) error {
	out, err := m.kubectl.Run("delete", "deployment", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to delete LLM %s: %s: %w", name, out, err)
	}

	return nil
}
