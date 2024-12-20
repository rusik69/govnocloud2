package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// VMManager handles VM operations
type VMManager struct {
	kubectl KubectlRunner
}

// NewVMManager creates a new VM manager instance
func NewVMManager() *VMManager {
	return &VMManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// VMTemplate represents the KubeVirt VM template
type VMTemplate struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Running  bool `json:"running"`
		Template struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				Domain struct {
					Devices   VMDevices   `json:"devices"`
					Resources VMResources `json:"resources"`
				} `json:"domain"`
				Networks []VMNetwork `json:"networks"`
				Volumes  []VMVolume  `json:"volumes"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

type VMDevices struct {
	Disks      []VMDisk      `json:"disks"`
	Interfaces []VMInterface `json:"interfaces"`
}

type VMDisk struct {
	Name string    `json:"name"`
	Disk VMDiskBus `json:"disk"`
}

type VMDiskBus struct {
	Bus string `json:"bus"`
}

type VMInterface struct {
	Name       string      `json:"name"`
	Masquerade interface{} `json:"masquerade"`
}

type VMResources struct {
	Requests VMResourceRequests `json:"requests"`
}

type VMResourceRequests struct {
	Memory string `json:"memory"`
	CPU    int    `json:"cpu"`
}

type VMNetwork struct {
	Name string      `json:"name"`
	Pod  interface{} `json:"pod"`
}

type VMVolume struct {
	Name          string           `json:"name"`
	ContainerDisk *ContainerDisk   `json:"containerDisk,omitempty"`
	CloudInit     *CloudInitConfig `json:"cloudInitNoCloud,omitempty"`
}

type ContainerDisk struct {
	Image string `json:"image"`
}

type CloudInitConfig struct {
	UserDataBase64 string `json:"userDataBase64"`
}

// CreateVMHandler handles VM creation requests
func CreateVMHandler(c *gin.Context) {
	var vm types.VM
	if err := c.BindJSON(&vm); err != nil {
		log.Printf("invalid request: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	manager := NewVMManager()
	if err := manager.CreateVM(vm); err != nil {
		log.Printf("failed to create VM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create VM: %v", err))
		return
	}
	log.Printf("vm created successfully: %v+", vm)
	respondWithSuccess(c, gin.H{"message": "VM created successfully"})
}

// CreateVM creates a new virtual machine
func (m *VMManager) CreateVM(vm types.VM) error {
	vmConfig := m.generateVMConfig(vm)
	log.Printf("VM config: %v", vmConfig)
	tempFile, err := m.writeVMConfig(vmConfig)
	if err != nil {
		log.Printf("failed to write VM config: %v", err)
		return fmt.Errorf("failed to write VM config: %w", err)
	}
	defer os.Remove(tempFile)

	if err := m.applyVMConfig(tempFile, vm.Namespace); err != nil {
		log.Printf("failed to apply VM config: %v", err)
		return fmt.Errorf("failed to apply VM config: %w", err)
	}

	log.Printf("VM %s created successfully", vm.Name)
	return nil
}

// generateVMConfig creates the VM configuration
func (m *VMManager) generateVMConfig(vm types.VM) VMTemplate {
	vmSize := types.VMSizes[vm.Size]

	return VMTemplate{
		APIVersion: "kubevirt.io/v1",
		Kind:       "VirtualMachine",
		Metadata: struct {
			Name string `json:"name"`
		}{
			Name: vm.Name,
		},
		Spec: struct {
			Running  bool `json:"running"`
			Template struct {
				Metadata struct {
					Labels map[string]string `json:"labels"`
				} `json:"metadata"`
				Spec struct {
					Domain struct {
						Devices   VMDevices   `json:"devices"`
						Resources VMResources `json:"resources"`
					} `json:"domain"`
					Networks []VMNetwork `json:"networks"`
					Volumes  []VMVolume  `json:"volumes"`
				} `json:"spec"`
			} `json:"template"`
		}{
			Running: false,
			Template: struct {
				Metadata struct {
					Labels map[string]string `json:"labels"`
				} `json:"metadata"`
				Spec struct {
					Domain struct {
						Devices   VMDevices   `json:"devices"`
						Resources VMResources `json:"resources"`
					} `json:"domain"`
					Networks []VMNetwork `json:"networks"`
					Volumes  []VMVolume  `json:"volumes"`
				} `json:"spec"`
			}{
				Metadata: struct {
					Labels map[string]string `json:"labels"`
				}{
					Labels: map[string]string{
						"kubevirt.io/size":   vm.Size,
						"kubevirt.io/domain": vm.Name,
						"kubevirt.io/image":  vm.Image,
					},
				},
				Spec: struct {
					Domain struct {
						Devices   VMDevices   `json:"devices"`
						Resources VMResources `json:"resources"`
					} `json:"domain"`
					Networks []VMNetwork `json:"networks"`
					Volumes  []VMVolume  `json:"volumes"`
				}{
					Domain: struct {
						Devices   VMDevices   `json:"devices"`
						Resources VMResources `json:"resources"`
					}{
						Devices: VMDevices{
							Disks: []VMDisk{{
								Name: "containerdisk",
								Disk: VMDiskBus{Bus: "virtio"},
							}},
							Interfaces: []VMInterface{{
								Name:       "default",
								Masquerade: struct{}{},
							}},
						},
						Resources: VMResources{
							Requests: VMResourceRequests{
								Memory: fmt.Sprintf("%dM", vmSize.RAM),
								CPU:    vmSize.CPU,
							},
						},
					},
					Networks: []VMNetwork{{
						Name: "default",
						Pod:  struct{}{},
					}},
					Volumes: []VMVolume{
						{
							Name: "containerdisk",
							ContainerDisk: &ContainerDisk{
								Image: vm.Image,
							},
						},
						{
							Name: "cloudinitdisk",
							CloudInit: &CloudInitConfig{
								UserDataBase64: "SGkuXG4=",
							},
						},
					},
				},
			},
		},
	}
}

// writeVMConfig writes the VM configuration to a temporary file
func (m *VMManager) writeVMConfig(config VMTemplate) (string, error) {
	tempFile, err := os.CreateTemp("", "vm-*.yaml")
	if err != nil {
		log.Printf("failed to create temp file: %v", err)
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		log.Printf("failed to encode VM config: %v", err)
		return "", fmt.Errorf("failed to encode VM config: %w", err)
	}

	return tempFile.Name(), nil
}

// applyVMConfig applies the VM configuration using kubectl
func (m *VMManager) applyVMConfig(configPath, namespace string) error {
	out, err := m.kubectl.Run("apply", "-f", configPath, "-n", namespace)
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %s: %w", out, err)
	}
	return nil
}

// ListVMsHandler handles VM listing requests
func ListVMsHandler(c *gin.Context) {
	manager := NewVMManager()
	vms, err := manager.ListVMs()
	if err != nil {
		log.Printf("failed to list VMs: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list VMs: %v", err))
		return
	}

	log.Printf("vms: %v+", vms)
	respondWithSuccess(c, vms)
}

// ListVMs returns a list of virtual machines
func (m *VMManager) ListVMs() ([]types.VM, error) {
	out, err := m.kubectl.Run("get", "VirtualMachines", "-o", "json")
	if err != nil {
		log.Printf("failed to list VMs: %v", err)
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Template struct {
					Metadata struct {
						Labels map[string]string `json:"labels"`
					} `json:"metadata"`
					Spec struct {
						Domain struct {
							Resources struct {
								Requests struct {
									Memory string `json:"memory"`
									CPU    string `json:"cpu"`
								} `json:"requests"`
							} `json:"resources"`
						} `json:"domain"`
					} `json:"spec"`
				} `json:"template"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		log.Printf("failed to parse VM list: %v", err)
		return nil, fmt.Errorf("failed to parse VM list: %w", err)
	}

	vms := make([]types.VM, len(result.Items))
	for i, item := range result.Items {
		vms[i] = types.VM{
			Name:  item.Metadata.Name,
			Size:  item.Spec.Template.Metadata.Labels["kubevirt.io/size"],
			Image: item.Spec.Template.Metadata.Labels["kubevirt.io/image"],
		}
	}

	return vms, nil
}

// GetVMHandler handles VM retrieval requests
func GetVMHandler(c *gin.Context) {
	var vm types.VM
	if err := c.BindJSON(&vm); err != nil {
		log.Printf("invalid request: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	manager := NewVMManager()
	vm, err := manager.GetVM(vm.Name, vm.Namespace)
	if err != nil {
		log.Printf("failed to get VM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get VM: %v", err))
		return
	}
	log.Printf("vm: %v", vm)
	respondWithSuccess(c, gin.H{"vm": vm})
}

// GetVM retrieves a specific virtual machine
func (m *VMManager) GetVM(name, namespace string) (types.VM, error) {
	out, err := m.kubectl.Run("get", "VirtualMachine", name, "-n", namespace, "-o", "json")
	if err != nil {
		return types.VM{}, fmt.Errorf("failed to get VM %s: %w", out, err)
	}
	var VMTemplate VMTemplate
	if err := json.Unmarshal(out, &VMTemplate); err != nil {
		return types.VM{}, fmt.Errorf("failed to parse VM %s: %w", out, err)
	}
	vm := types.VM{
		Name:  VMTemplate.Metadata.Name,
		Size:  VMTemplate.Spec.Template.Metadata.Labels["kubevirt.io/size"],
		Image: VMTemplate.Spec.Template.Metadata.Labels["kubevirt.io/image"],
	}
	return vm, nil
}

// DeleteVMHandler handles VM deletion requests
func DeleteVMHandler(c *gin.Context) {
	vmName := c.Param("name")
	if vmName == "" {
		log.Printf("VM name is required")
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}

	manager := NewVMManager()
	if err := manager.DeleteVM(vmName); err != nil {
		log.Printf("failed to delete VM %s: %v", vmName, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete VM: %v", err))
		return
	}

	log.Printf("VM %s deleted successfully", vmName)
	respondWithSuccess(c, gin.H{"message": "VM deleted successfully"})
}

// DeleteVM removes a virtual machine
func (m *VMManager) DeleteVM(name string) error {
	out, err := m.kubectl.Run("delete", "VirtualMachine", name)
	if err != nil {
		log.Printf("failed to delete VM %s: %s: %v", name, out, err)
		return fmt.Errorf("failed to delete VM %s: %s: %w", name, out, err)
	}
	return nil
}
