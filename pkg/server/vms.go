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
	var vm types.VM
	if err := c.BindJSON(&vm); err != nil {
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
	if err := vmManager.CreateVM(namespace, vm); err != nil {
		log.Printf("failed to create VM: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create VM: %v", err))
		return
	}
	log.Printf("vm created successfully: %v+", vm)
	respondWithSuccess(c, gin.H{"message": "VM created successfully"})
}

// CreateVM creates a new virtual machine
func (m *VMManager) CreateVM(namespace string, vm types.VM) error {
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
		vm.Name, namespace, vm.Size, vm.Image, vmSize.RAM, vmSize.CPU, vmImage.Image)
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

	// wait for VM to start
	out, err = m.kubectl.Run("wait", "--for=condition=Ready=true", fmt.Sprintf("virtualmachine.kubevirt.io/%s", vm.Name), "-n", vm.Namespace, "--timeout=5m")
	if err != nil {
		return fmt.Errorf("failed waiting for VM %s to start in namespace %s: %s %w", vm.Name, vm.Namespace, out, err)
	}

	return nil
}

// ListVMsHandler handles VM listing requests
func ListVMsHandler(c *gin.Context) {
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
	vms, err := vmManager.ListVMs(namespace)
	if err != nil {
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
	PrintableStatus string `json:"printableStatus"`
}

// GetVMHandler handles VM retrieval requests
func GetVMHandler(c *gin.Context) {
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
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}

	vm, err := vmManager.GetVM(name, namespace)
	if err != nil {
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
		Name:      VMTemplate.Metadata.Name,
		Namespace: namespace,
		Size:      VMTemplate.Spec.Template.Metadata.Labels["kubevirt.io/size"],
		Image:     VMTemplate.Spec.Template.Metadata.Labels["kubevirt.io/image"],
		Status:    VMTemplate.PrintableStatus,
	}
	return vm, nil
}

// DeleteVMHandler handles VM deletion requests
func DeleteVMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}

	if err := vmManager.DeleteVM(name, namespace); err != nil {
		log.Printf("failed to delete VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete VM: %v", err))
		return
	}

	log.Printf("VM %s deleted successfully", name)
	respondWithSuccess(c, gin.H{"message": "VM deleted successfully"})
}

// DeleteVM removes a virtual machine
func (m *VMManager) DeleteVM(name, namespace string) error {
	out, err := m.kubectl.Run("delete", "VirtualMachineInstance", name, "-n", namespace)
	if err != nil {
		return fmt.Errorf("failed to delete VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	return nil
}

// StartVMHandler handles VM start requests
func StartVMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}
	// check if VM is already running
	vm, err := vmManager.GetVM(name, namespace)
	if err != nil {
		log.Printf("failed to get VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get VM: %v", err))
		return
	}
	if vm.Status == "Running" {
		log.Printf("VM %s is already running in namespace %s", name, namespace)
		respondWithSuccess(c, gin.H{"message": "VM is already running"})
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
	log.Printf("starting VM %s in namespace %s", name, namespace)
	out, err := m.virtctl.Run("start", name, "-n", namespace)
	if err != nil {
		// Check if the error is because VM is already running
		if strings.Contains(string(out), "VM is already running") || strings.Contains(string(out), "already running") {
			log.Printf("VM %s is already running in namespace %s", name, namespace)
			return nil // Not an error, VM is already in desired state
		}
		return fmt.Errorf("failed to start VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	// wait for VM to start
	log.Printf("waiting for VM %s to start in namespace %s", name, namespace)
	out, err = m.kubectl.Run("wait", "--for=condition=Ready=true", fmt.Sprintf("virtualmachine.kubevirt.io/%s", name), "-n", namespace, "--timeout=5m")
	if err != nil {
		return fmt.Errorf("failed waiting for VM %s to start in namespace %s: %s %w", name, namespace, out, err)
	}
	log.Printf("VM %s started successfully in namespace %s", name, namespace)
	return nil
}

// StopVMHandler handles VM stop requests
func StopVMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
		return
	}
	// check if VM is already stopped
	vm, err := vmManager.GetVM(name, namespace)
	if err != nil {
		log.Printf("failed to get VM %s in namespace %s: %v", name, namespace, err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get VM: %v", err))
		return
	}
	if vm.Status == "Stopped" {
		log.Printf("VM %s is already stopped in namespace %s", name, namespace)
		respondWithSuccess(c, gin.H{"message": "VM is already stopped"})
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
	log.Printf("stopping VM %s in namespace %s", name, namespace)
	out, err := m.virtctl.Run("stop", name, "-n", namespace)
	if err != nil {
		// Check if the error is because VM is already stopped
		if strings.Contains(string(out), "VM is not running") || strings.Contains(string(out), "already stopped") || strings.Contains(string(out), "not running") {
			log.Printf("VM %s is already stopped in namespace %s", name, namespace)
			return nil // Not an error, VM is already in desired state
		}
		return fmt.Errorf("failed to stop VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	// wait for VM to stop
	log.Printf("waiting for VM %s to stop in namespace %s", name, namespace)
	out, err = m.kubectl.Run("wait", "--for=condition=Ready=false", fmt.Sprintf("virtualmachine.kubevirt.io/%s", name), "-n", namespace, "--timeout=5m")
	if err != nil {
		return fmt.Errorf("failed waiting for VM %s to stop in namespace %s: %s %w", name, namespace, out, err)
	}
	log.Printf("VM %s stopped successfully in namespace %s", name, namespace)
	return nil
}

// RestartVMHandler handles VM restart requests
func RestartVMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
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
	log.Printf("restarting VM %s in namespace %s", name, namespace)

	// First try to stop the VM (with force to ensure it stops)
	log.Printf("stopping VM %s in namespace %s", name, namespace)
	out, err := m.virtctl.Run("stop", name, "-n", namespace, "--grace-period=1", "--force=true")
	if err != nil {
		// Check if the error is because VM is already stopped
		if strings.Contains(string(out), "VM is not running") || strings.Contains(string(out), "already stopped") || strings.Contains(string(out), "not running") {
			log.Printf("VM %s is already stopped in namespace %s", name, namespace)
		} else {
			return fmt.Errorf("failed to stop VM %s in namespace %s: %s %w", name, namespace, out, err)
		}
	} else {
		log.Printf("VM %s stopped successfully in namespace %s", name, namespace)
	}

	// Now start the VM using the existing StartVM method which has proper error handling
	log.Printf("starting VM %s in namespace %s", name, namespace)
	err = m.StartVM(name, namespace)
	if err != nil {
		return fmt.Errorf("failed to start VM after restart: %w", err)
	}

	log.Printf("VM %s restarted successfully in namespace %s", name, namespace)
	return nil
}

// WaitVMHandler handles VM wait requests
func WaitVMHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	name := c.Param("name")
	if name == "" {
		respondWithError(c, http.StatusBadRequest, "VM name is required")
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		log.Printf("namespace is required")
		respondWithError(c, http.StatusBadRequest, "namespace is required")
		return
	}
	if !CheckNamespaceAccess(username, namespace) {
		respondWithError(c, http.StatusForbidden, "user does not have access to this namespace")
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
	log.Printf("waiting for VM %s to be ready in namespace %s", name, namespace)
	out, err := m.kubectl.Run("wait", "VirtualMachine", name, "-n", namespace, "--for=condition=ready", "--timeout=10m")
	if err != nil {
		return fmt.Errorf("failed to wait for VM %s in namespace %s: %s %w", name, namespace, out, err)
	}
	log.Printf("VM %s is ready in namespace %s", name, namespace)
	return nil
}
