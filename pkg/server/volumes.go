package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// VolumeManager handles volume operations
type VolumeManager struct {
	kubectl KubectlRunner
}

// NewVolumeManager creates a new volume manager
func NewVolumeManager() *VolumeManager {
	return &VolumeManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// CreateVolumeHandler creates a new volume
func (m *VolumeManager) CreateVolume(volume types.Volume, namespace string) (string, error) {
	// Create longhorn volume
	pvc := fmt.Sprintf(`apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: %s
  namespace: %s
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: %s
  storageClassName: longhorn
`, volume.Name, namespace, volume.Size)
	log.Println(pvc)
	tempFile, err := os.CreateTemp("", "longhorn-pvc.yaml")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	_, err = tempFile.WriteString(pvc)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	out, err := m.kubectl.Run("apply", "-f", tempFile.Name())
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// DeleteVolume deletes a volume
func (m *VolumeManager) DeleteVolume(volume, namespace string) (string, error) {
	out, err := m.kubectl.Run("delete", "pvc", volume, "-n", namespace)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ListVolumes lists all volumes
func (m *VolumeManager) ListVolumes(namespace string) ([]string, error) {
	out, err := m.kubectl.Run("get", "pvc", "-n", namespace, "-o", "jsonpath={range .items[*]}{.metadata.name} {.spec.resources.requests.storage}{'\n'}{end}")
	if err != nil {
		return nil, err
	}
	volumes := strings.Split(string(out), "\n")
	res := []string{}
	for _, volume := range volumes {
		parts := strings.Split(volume, " ")
		if len(parts) != 2 {
			continue
		}
		res = append(res, parts[0])
	}
	return res, nil
}

// GetVolume gets details of a specific volume
func (m *VolumeManager) GetVolume(name, namespace string) (types.Volume, error) {
	out, err := m.kubectl.Run("get", "pvc", name, "-n", namespace, "-o", "jsonpath={.metadata.name} {.spec.resources.requests.storage}")
	if err != nil {
		return types.Volume{}, err
	}
	parts := strings.Split(string(out), " ")
	if len(parts) != 2 {
		return types.Volume{}, fmt.Errorf("invalid output format")
	}
	size := parts[1]
	return types.Volume{Name: name, Size: size}, nil
}

// CreateVolumeHandler creates a new volume
func CreateVolumeHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}
	volume := types.Volume{}
	m := NewVolumeManager()
	if err := c.ShouldBindJSON(&volume); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := m.CreateVolume(volume, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println(out)
	c.JSON(http.StatusOK, gin.H{"message": "Volume created", "output": out})
}

// DeleteVolumeHandler deletes a volume
func DeleteVolumeHandler(c *gin.Context) {
	m := NewVolumeManager()
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}
	out, err := m.DeleteVolume(name, namespace)
	if err != nil {
		log.Printf("failed to delete volume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println(out)
	c.JSON(http.StatusOK, gin.H{"message": "Volume deleted", "output": out})
}

// ListVolumesHandler lists all volumes
func ListVolumesHandler(c *gin.Context) {
	m := NewVolumeManager()
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}
	volumes, err := m.ListVolumes(namespace)
	if err != nil {
		log.Printf("failed to list volumes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, volumes)
}

// GetVolumeHandler gets details of a specific volume
func GetVolumeHandler(c *gin.Context) {
	m := NewVolumeManager()
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}
	volume, err := m.GetVolume(name, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println(volume)
	c.JSON(http.StatusOK, volume)
}
