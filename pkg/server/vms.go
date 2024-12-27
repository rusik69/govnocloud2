package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log"

	"strings"

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
	log.Printf("%+v", vm)
	if _, ok := types.VMSizes[vm.Size]; !ok {
		log.Printf("invalid VM size: %s", vm.Size)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid VM size: %s", vm.Size))
		return
	}
	if _, ok := types.VMImages[vm.Image]; !ok {
		log.Printf("invalid VM image: %s", vm.Image)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid VM image: %s", vm.Image))
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
	cmd := fmt.Sprintf("virtctl create vm --instancetype %s --name %s --namespace %s | kubectl create -f -", vm.Size, vm.Name, vm.Namespace)
	log.Println(cmd)
	if _, err := m.kubectl.Run(cmd); err != nil {
		return fmt.Errorf("failed to create VM %s: %w", vm.Name, err)
	}
	log.Printf("VM %s created successfully", vm.Name)
	return nil
}

// ListVMsHandler handles VM listing requests
func ListVMsHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	manager := NewVMManager()
	vms, err := manager.ListVMs(namespace)
	if err != nil {
		log.Printf("failed to list VMs: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list VMs: %v", err))
		return
	}
	log.Printf("vms: %+v", vms)
	respondWithSuccess(c, vms)
}

// ListVMs returns a list of virtual machines
func (m *VMManager) ListVMs(namespace string) ([]string, error) {
	out, err := m.kubectl.Run("get", "VirtualMachines", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		log.Printf("failed to list VMs: %v", err)
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	// If output is empty, return empty slice
	if len(out) == 0 {
		return []string{}, nil
	}

	// Split the space-separated output into slice
	names := strings.Fields(string(out))
	return names, nil
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

	manager := NewVMManager()
	vm, err := manager.GetVM(name, namespace)
	if err != nil {
		log.Printf("failed to get VM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get VM: %v", err))
		return
	}
	log.Printf("%+v", vm)
	c.JSON(http.StatusOK, vm)
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
	name := c.Param("name")
	if name == "" {
		log.Printf("VM name is required")
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}

	manager := NewVMManager()
	if err := manager.DeleteVM(name, namespace); err != nil {
		log.Printf("failed to delete VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete VM: %v", err))
		return
	}

	log.Printf("VM %s deleted successfully", name)
	respondWithSuccess(c, gin.H{"message": "VM deleted successfully"})
}

// DeleteVM removes a virtual machine
func (m *VMManager) DeleteVM(name, namespace string) error {
	out, err := m.kubectl.Run("delete", "VirtualMachine", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to delete VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}
