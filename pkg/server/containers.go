package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

// ContainerManager handles container operations
type ContainerManager struct {
	kubectl KubectlRunner
}

// NewContainerManager creates a new container manager
func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListContainersHandler handles requests to list containers
func ListContainersHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	manager := NewContainerManager()
	containers, err := manager.ListContainers(namespace)
	if err != nil {
		log.Printf("failed to list containers: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list containers: %v", err))
		return
	}
	respondWithSuccess(c, containers)
}

// CreateContainerHandler handles requests to create a new container
func CreateContainerHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	var container types.Container
	if err := c.BindJSON(&container); err != nil {
		log.Printf("invalid request: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}
	container.Namespace = namespace
	container.Name = name
	manager := NewContainerManager()
	if err := manager.CreateContainer(&container); err != nil {
		log.Printf("container create error: %s", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create container: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "Container created successfully", "container": container})
}

// GetContainerHandler handles requests to get container details
func GetContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "container name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	manager := NewContainerManager()
	container, err := manager.GetContainer(name, namespace)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get container: %v", err))
		return
	}

	if container == nil {
		respondWithError(c, http.StatusNotFound, "container not found")
		return
	}

	respondWithSuccess(c, gin.H{"container": container})
}

// DeleteContainerHandler handles requests to delete a container
func DeleteContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		log.Printf("container name is required")
		respondWithError(c, http.StatusBadRequest, "container name is required")
		return
	}

	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	manager := NewContainerManager()
	if err := manager.DeleteContainer(name, namespace); err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete container: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "Container deleted successfully"})
}

// Manager methods
func (m *ContainerManager) generatePodManifest(container *types.Container) (string, error) {
	pod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
  labels:
    app: %s
    type: container
spec:
  containers:
    - name: %s
      image: %s
      ports:
        - containerPort: %d
  resources:
    requests:
      cpu: %d
      memory: %dMi	
`, container.Name, container.Name, container.Name, container.Image, container.Port, container.CPU, container.RAM)

	return pod, nil
}

func (m *ContainerManager) ListContainers(namespace string) ([]types.Container, error) {
	out, err := m.kubectl.Run("get", "pods", "-l", "type=container", "-o", "json", "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %s %w", out, err)
	}

	var podList corev1.PodList
	if err := json.Unmarshal([]byte(out), &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}

	containers := []types.Container{}
	for _, pod := range podList.Items {
		container := podToContainer(&pod)
		containers = append(containers, *container)
	}

	return containers, nil
}

func (m *ContainerManager) CreateContainer(container *types.Container) error {
	pod, err := m.generatePodManifest(container)
	if err != nil {
		return fmt.Errorf("failed to generate pod manifest: %w", err)
	}

	log.Printf("pod manifest: %v", string(pod))
	tmpFile, err := os.CreateTemp("", "container-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(pod), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if out, err := m.kubectl.Run("apply", "-f", tmpFile.Name(), "-n", container.Namespace); err != nil {
		return fmt.Errorf("failed to create container pod: %s, %w", out, err)
	}

	return nil
}

func (m *ContainerManager) GetContainer(name, namespace string) (*types.Container, error) {
	out, err := m.kubectl.Run("get", "pod", name, "-o", "json", "-n", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get container pod: %s, %w", out, err)
	}

	var pod corev1.Pod
	if err := json.Unmarshal([]byte(out), &pod); err != nil {
		return nil, fmt.Errorf("failed to parse pod details: %w", err)
	}

	return podToContainer(&pod), nil
}

func (m *ContainerManager) DeleteContainer(name, namespace string) error {
	if out, err := m.kubectl.Run("delete", "pod", name, "-n", namespace); err != nil {
		return fmt.Errorf("failed to delete container pod: %s, %w", out, err)
	}
	return nil
}

func podToContainer(pod *corev1.Pod) *types.Container {
	container := pod.Spec.Containers[0]
	return &types.Container{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Image:     container.Image,
		Port:      int(container.Ports[0].ContainerPort),
		CPU:       int(container.Resources.Requests.Cpu().MilliValue()),
		RAM:       int(container.Resources.Requests.Memory().Value() / 1024 / 1024), // Convert to Mi
		Env:       envVarsToStrings(container.Env),
	}
}

func envVarsToStrings(envVars []corev1.EnvVar) []string {
	var result []string
	for _, env := range envVars {
		result = append(result, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	return result
}
