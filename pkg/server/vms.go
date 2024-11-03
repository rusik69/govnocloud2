package server

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateVMHandler creates a new VM.
func CreateVMHandler(c *gin.Context) {
	var vm types.VM
	if err := c.BindJSON(&vm); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	vmYaml := fmt.Sprintf(`
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: %s
spec:
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/size: %s
        kubevirt.io/domain: %s
    spec:
      domain:
        devices:
          disks:
            - name: containerdisk
              disk:
                bus: virtio
              interfaces:
              - name: default
                masquerade: {}
              resources:
                requests:
                  memory: %dM
cpu: %d
      networks:
	  - name: default
	  	pod: {}
	  volumes:
	  - name: containerdisk
	    containerDisk:
		  image: %s
	  - name: cloudinitdisk
	    cloudInitNoCloud:
		  userDataBase64: SGkuXG4=
`, vm.Name, vm.Size, vm.Name, types.VMSizes[vm.Size].RAM, types.VMSizes[vm.Size].CPU, vm.Image)
	vmFile, err := os.CreateTemp("", "vm.yaml")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer vmFile.Close()
	_, err = vmFile.WriteString(vmYaml)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	command := exec.Command("kubectl", "apply", "-f", vmFile.Name())
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM created"})
}
