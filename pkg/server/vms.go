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
func (m *VMManager) generateVMConfig(vm types.VM) string {
	vmSize := types.VMSizes[vm.Size]
	vmImage := types.VMImages[vm.Image]
	vmConfig := fmt.Sprintf(`
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: %s
  namespace: %s
spec:
  running: true
  template:
    metadata:
      labels:
        kubevirt.io/size: %s
        kubevirt.io/image: %s
    spec:
	  source:
	  	http:
	  		url: %s
      domain:
        resources:
          requests:
            memory: %dMi
            cpu: %d
`, vm.Name, vm.Namespace, vmSize.Name, vmImage.Name, vmImage.URL, vmSize.RAM, vmSize.CPU)

	return vmConfig
}

// writeVMConfig writes the VM configuration to a temporary file
func (m *VMManager) writeVMConfig(config string) (string, error) {
	tempFile, err := os.CreateTemp("", "vm-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	if err := os.WriteFile(tempFile.Name(), []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write VM config: %w", err)
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

// Add this struct before the GetVM function
type VMTemplate struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Template struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"template"`
	} `json:"spec"`
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
