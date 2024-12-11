package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	manager := NewContainerManager()
	containers, err := manager.ListContainers()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list containers: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"containers": containers})
}

// CreateContainerHandler handles requests to create a new container
func CreateContainerHandler(c *gin.Context) {
	var container types.Container
	if err := c.BindJSON(&container); err != nil {
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	manager := NewContainerManager()
	if err := manager.CreateContainer(&container); err != nil {
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

	manager := NewContainerManager()
	container, err := manager.GetContainer(name)
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
		respondWithError(c, http.StatusBadRequest, "container name is required")
		return
	}

	manager := NewContainerManager()
	if err := manager.DeleteContainer(name); err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete container: %v", err))
		return
	}

	respondWithSuccess(c, gin.H{"message": "Container deleted successfully"})
}

// Manager methods
func (m *ContainerManager) generatePodManifest(container *types.Container) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: container.Name,
			Labels: map[string]string{
				"app":  container.Name,
				"type": "container",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  container.Name,
					Image: container.Image,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: int32(container.Port),
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", container.CPU)),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", container.RAM)),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", container.CPU)),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", container.RAM)),
						},
					},
					Env: generateEnvVars(container.Env),
				},
			},
		},
	}

	return pod, nil
}

func generateEnvVars(envVars []string) []corev1.EnvVar {
	var result []corev1.EnvVar
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			result = append(result, corev1.EnvVar{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}
	return result
}

func (m *ContainerManager) ListContainers() ([]types.Container, error) {
	out, err := m.kubectl.Run("get", "pods", "-l", "type=container", "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var podList corev1.PodList
	if err := json.Unmarshal([]byte(out), &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}

	containers := make([]types.Container, 0, len(podList.Items))
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

	podJSON, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to marshal pod manifest: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "container-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), podJSON, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if _, err := m.kubectl.Run("apply", "-f", tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to create container pod: %w", err)
	}

	return nil
}

func (m *ContainerManager) GetContainer(name string) (*types.Container, error) {
	out, err := m.kubectl.Run("get", "pod", name, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get container pod: %w", err)
	}

	var pod corev1.Pod
	if err := json.Unmarshal([]byte(out), &pod); err != nil {
		return nil, fmt.Errorf("failed to parse pod details: %w", err)
	}

	return podToContainer(&pod), nil
}

func (m *ContainerManager) DeleteContainer(name string) error {
	if _, err := m.kubectl.Run("delete", "pod", name); err != nil {
		return fmt.Errorf("failed to delete container pod: %w", err)
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

