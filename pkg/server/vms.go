package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log"

	"strings"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// VMManager handles VM operations
type VMManager struct {
	kubectl KubectlRunner
	virtctl VirtctlRunner
}

// NewVMManager creates a new VM manager instance
func NewVMManager() *VMManager {
	return &VMManager{
		kubectl: &DefaultKubectlRunner{},
		virtctl: &DefaultVirtctlRunner{},
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
	if err := vmManager.CreateVM(vm); err != nil {
		log.Printf("failed to create VM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create VM: %v", err))
		return
	}
	log.Printf("vm created successfully: %v+", vm)
	respondWithSuccess(c, gin.H{"message": "VM created successfully"})
}

// DeleteVolumeForVM deletes the volume for a VM
func DeleteVolumeForVM(vm types.VM) error {
	out, err := volumeManager.DeleteVolume(vm.Name, vm.Namespace)
	if err != nil {
		return fmt.Errorf("failed to delete volume for VM %s: %w %s", vm.Name, err, out)
	}
	log.Printf("volume deleted successfully: %s", out)
	return nil
}

// CreateVM creates a new virtual machine
func (m *VMManager) CreateVM(vm types.VM) error {
	vmSize := types.VMSizes[vm.Size]
	vmImage := types.VMImages[vm.Image]
	vmConfig := fmt.Sprintf(`apiVersion: kubevirt.io/v1
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
      domain:
        devices:
          disks:
          - name: rootdisk
            disk:
              bus: virtio
          - name: cloudinitdisk
            disk:
              bus: virtio
        resources:
          requests:
            memory: %dMi
            cpu: %d
      volumes:
      - name: rootdisk
        containerDisk:
          image: %s
      - name: cloudinitdisk
        cloudInitNoCloud:
          userData: |
            #cloud-config
            password: ubuntu
            chpasswd:
              expire: false
            ssh_pwauth: true`,
		vm.Name, vm.Namespace, vm.Size, vm.Image, vmSize.RAM, vmSize.CPU, vmImage.Image)
	log.Println(vmConfig)
	// Write config to temporary file
	tmpfile, err := os.CreateTemp("", "vm-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(vmConfig); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Apply the configuration
	out, err := m.kubectl.Run("apply", "-f", tmpfile.Name(), "--wait=true", "--timeout=600s")
	if err != nil {
		return fmt.Errorf("failed to create VM %s: %s: %w", vm.Name, out, err)
	}

	return nil
}

// ListVMsHandler handles VM listing requests
func ListVMsHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	vms, err := vmManager.ListVMs(namespace)
	if err != nil {
		log.Printf("failed to list VMs: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list VMs: %v", err))
		return
	}
	log.Printf("vms: %+v", vms)
	c.JSON(http.StatusOK, vms)
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
	// Filter out reserved namespaces
	res := []string{}
	for _, name := range names {
		if !types.ReservedNamespaces[name] {
			res = append(res, name)
		}
	}
	return res, nil
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

	vm, err := vmManager.GetVM(name, namespace)
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

	// Check if namespace is reserved
	if types.ReservedNamespaces[namespace] {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}

	if err := vmManager.DeleteVM(name, namespace); err != nil {
		log.Printf("failed to delete VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete VM: %v", err))
		return
	}

	vm, err := vmManager.GetVM(name, namespace)
	if err != nil {
		log.Printf("failed to get VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get VM: %v", err))
		return
	}

	if err := DeleteVolumeForVM(vm); err != nil {
		log.Printf("failed to delete volume for VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete volume: %v", err))
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

// StartVMHandler handles VM start requests
func StartVMHandler(c *gin.Context) {
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
	// check if namespace is reserved
	if types.ReservedNamespaces[namespace] {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	if err := vmManager.StartVM(name, namespace); err != nil {
		log.Printf("failed to start VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to start VM: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"message": "VM started successfully"})
}

// StartVM starts a virtual machine
func (m *VMManager) StartVM(name, namespace string) error {
	out, err := m.virtctl.Run("start", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to start VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}

// StopVMHandler handles VM stop requests
func StopVMHandler(c *gin.Context) {
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
	// check if namespace is reserved
	if types.ReservedNamespaces[namespace] {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	if err := vmManager.StopVM(name, namespace); err != nil {
		log.Printf("failed to stop VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to stop VM: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"message": "VM stopped successfully"})
}

// StopVM stops a virtual machine
func (m *VMManager) StopVM(name, namespace string) error {
	out, err := m.virtctl.Run("stop", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to stop VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	// wait for VM to stop
	out, err = m.kubectl.Run("wait", "--for=condition=Status=Stopped", fmt.Sprintf("virtualmachine.kubevirt.io/%s", name), "-n", namespace, "--timeout=1m")
	if err != nil {
		return fmt.Errorf("failed waiting for VM %s to stop in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}

// RestartVMHandler handles VM restart requests
func RestartVMHandler(c *gin.Context) {
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
	// check if namespace is reserved
	if types.ReservedNamespaces[namespace] {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	if err := vmManager.RestartVM(name, namespace); err != nil {
		log.Printf("failed to restart VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to restart VM: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"message": "VM restarted successfully"})
}

// RestartVM restarts a virtual machine
func (m *VMManager) RestartVM(name, namespace string) error {
	out, err := m.virtctl.Run("restart", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to restart VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}

// WaitVMHandler handles VM wait requests
func WaitVMHandler(c *gin.Context) {
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
	// check if namespace is reserved
	if types.ReservedNamespaces[namespace] {
		log.Printf("namespace %s is reserved", namespace)
		respondWithError(c, http.StatusForbidden, fmt.Sprintf("namespace %s is reserved", namespace))
		return
	}
	if err := vmManager.WaitVM(name, namespace); err != nil {
		log.Printf("failed to wait for VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to wait for VM: %v", err))
		return
	}
	respondWithSuccess(c, gin.H{"message": "VM waited successfully"})
}

// WaitVM waits for a virtual machine to be ready
func (m *VMManager) WaitVM(name, namespace string) error {
	out, err := m.kubectl.Run("wait", "VirtualMachine", name, "-n", namespace, "--for=condition=ready", "--timeout=10m")
	if err != nil {
		return fmt.Errorf("failed to wait for VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}
