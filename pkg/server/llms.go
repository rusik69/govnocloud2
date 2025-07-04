package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// LLMManager handles LLM operations
type LLMManager struct {
	kubectl KubectlRunner
}

// NewLLMManager creates a new LLM manager instance
func NewLLMManager() *LLMManager {
	return &LLMManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// CreateLLMHandler handles LLM creation requests
func CreateLLMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" {
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}
	var llm types.LLM
	if err := c.BindJSON(&llm); err != nil {
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if llm.Type == "" {
		respondWithError(c, http.StatusBadRequest, "type is required")
		return
	}

	if _, ok := types.LLMTypes[llm.Type]; !ok {
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
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" {
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	if name == "" {
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}

	llm, err := llmManager.GetLLM(namespace, name)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get LLM: %v", err))
		return
	}

	respondWithSuccess(c, llm)
}

// DeleteLLMHandler handles LLM deletion requests
func DeleteLLMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" {
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	if name == "" {
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}

	if err := llmManager.DeleteLLM(namespace, name); err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete LLM: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "LLM deleted successfully"})
}

// CreateLLM creates a new LLM deployment
func (m *LLMManager) CreateLLM(llm types.LLM) error {
	llmType := types.LLMTypes[llm.Type]
	llmConfig := fmt.Sprintf(`apiVersion: ollama.ayaka.io/v1
kind: Model
metadata:
  name: %s
  namespace: %s
spec:
  image: %s`,
		llm.Name, llm.Namespace, llmType.Type)
	log.Printf("llmConfig: %s", llmConfig)
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
	cmd := []string{"apply", "-f", tmpfile.Name()}
	out, err := m.kubectl.Run(cmd...)
	if err != nil {
		return fmt.Errorf("failed to create LLM %s: %s: %w", llm.Name, out, err)
	}

	return nil
}

// GetLLM retrieves an LLM deployment
func (m *LLMManager) GetLLM(namespace, name string) (*types.LLM, error) {
	cmd := []string{"get", "Model", name, "-n", namespace, "-o", "json"}
	out, err := m.kubectl.Run(cmd...)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM %s: %s: %w", name, out, err)
	}

	// Parse the JSON output and create an LLM object
	model := struct {
		Spec struct {
			Image string `json:"image"`
		} `json:"spec"`
	}{}
	if err := json.Unmarshal(out, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LLM: %w", err)
	}

	llm := &types.LLM{
		Name:      name,
		Namespace: namespace,
		Type:      model.Spec.Image,
	}
	log.Printf("llm: %+v", llm)
	return llm, nil
}

// DeleteLLM deletes LLM
func (m *LLMManager) DeleteLLM(namespace, name string) error {
	cmd := []string{"delete", "Model", name, "-n", namespace}
	out, err := m.kubectl.Run(cmd...)
	if err != nil {
		return fmt.Errorf("failed to delete LLM %s: %s: %w", name, out, err)
	}

	return nil
}

// ListLLMsHandler handles list llms request
func ListLLMsHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}
	llms, err := llmManager.ListLLMs(namespace)
	if err != nil {
		log.Printf("failed to list LLMs: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list LLMs: %v", err))
		return
	}

	c.JSON(http.StatusOK, llms)
}

// ListLLMs lists all LLMs in a namespace
func (m *LLMManager) ListLLMs(namespace string) ([]types.LLM, error) {
	cmd := []string{"get", "Model", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name},{.items[*].spec.image}"}
	out, err := m.kubectl.Run(cmd...)
	if err != nil {
		return nil, fmt.Errorf("failed to list LLMs: %s: %w", out, err)
	}
	if len(out) == 0 {
		return []types.LLM{}, nil
	}

	parts := strings.Split(string(out), ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected output format from kubectl")
	}

	names := strings.Fields(parts[0])
	images := strings.Fields(parts[1])

	if len(names) != len(images) {
		return nil, fmt.Errorf("mismatched number of names and images")
	}

	models := make([]types.LLM, len(names))
	for i := range names {
		models[i] = types.LLM{
			Name:      names[i],
			Namespace: namespace,
			Type:      images[i],
		}
	}
	log.Printf("models: %+v", models)
	return models, nil
}
