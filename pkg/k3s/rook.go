package k3s

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const RookVersion = "release-1.15"

// InstallRook installs Rook to k3s cluster.
func InstallRook() error {
	// create rook-ceph namespace
	command := exec.Command("kubectl", "create", "namespace", "rook-ceph")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error creating Rook namespace: %w", err)
	}
	rookCommon := "https://raw.githubusercontent.com/rook/rook/" + RookVersion + "/deploy/examples/common.yaml"
	rootToolbox := "https://raw.githubusercontent.com/rook/rook/" + RookVersion + "/deploy/examples/toolbox.yaml"
	rookCrds := "https://raw.githubusercontent.com/rook/rook/" + RookVersion + "/deploy/examples/crds.yaml"
	rookOperator := "https://raw.githubusercontent.com/rook/rook/" + RookVersion + "/deploy/examples/operator.yaml"
	command = exec.Command("kubectl", "apply", "-f", rookOperator, "-f", rookCrds, "-f", rookCommon, "-f", rootToolbox)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing Rook operator: %w", err)
	}
	log.Println("Waiting for Rook Operator to be in Running state")
	command = exec.Command("kubectl", "wait", "--for=condition=available", "deployment/rook-ceph-operator", "-n", "rook-ceph", "--timeout=600s")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		log.Fatalf("error waiting for Rook operator to be in running state: %v", err)
	}
	clusterConfig := `
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: rook-ceph
spec:
  cephVersion:
    image: ceph/ceph:v16.2.6
  dataDirHostPath: /var/lib/rook
  mon:
    count: 1
    allowMultiplePerNode: false
  dashboard:
    enabled: true
  storage:
    useAllNodes: true
    useAllDevices: false
    config:
      storeType: bluestore
    nodes:
    - devices:
      - name: /dev/sda
  placement:
    all:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: kubernetes.io/hostname
              operator: Exists
  resources:
    mgr:
      limits:
        cpu: "500m"
        memory: "1024Mi"
      requests:
        cpu: "500m"
        memory: "1024Mi"
    mon:
      limits:
        cpu: "500m"
        memory: "1024Mi"
      requests:
        cpu: "500m"
        memory: "1024Mi"
    osd:
      limits:
        cpu: "500m"
        memory: "2048Mi"
      requests:
        cpu: "500m"
        memory: "2048Mi"
`
	// Write the cluster configuration to a file
	clusterConfigPath := "/tmp/rook-cluster.yaml"
	err := os.WriteFile(clusterConfigPath, []byte(clusterConfig), 0644)
	if err != nil {
		return err
	}

	// Step 3: Apply the Cluster Configuration
	command = exec.Command("kubectl", "apply", "-f", clusterConfigPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}
