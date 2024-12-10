package k3s

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// RookConfig holds configuration for Rook installation
type RookConfig struct {
	Version     string
	Namespace   string
	BaseURL     string
	Manifests   []string
	ClusterSpec ClusterSpec
}

// ClusterSpec holds Ceph cluster configuration
type ClusterSpec struct {
	CephVersion struct {
		Image string
	}
	DataDirHostPath string
	Mon             MonConfig
	Dashboard       DashboardConfig
	Storage         StorageConfig
	Placement       PlacementConfig
	Resources       ResourceRequirements
}

type MonConfig struct {
	Count                int
	AllowMultiplePerNode bool
}

type DashboardConfig struct {
	Enabled bool
}

type StorageConfig struct {
	UseAllNodes   bool
	UseAllDevices bool
	Config        map[string]string
	Nodes         []StorageNode
}

type StorageNode struct {
	Devices []DeviceSpec
}

type DeviceSpec struct {
	Name string
}

type PlacementConfig struct {
	All NodeAffinityConfig
}

type NodeAffinityConfig struct {
	NodeAffinity NodeAffinityRule
}

type NodeAffinityRule struct {
	RequiredDuringSchedulingIgnoredDuringExecution NodeSelectorTerm
}

type NodeSelectorTerm struct {
	MatchExpressions []MatchExpression
}

type MatchExpression struct {
	Key      string
	Operator string
}

type ResourceRequirements struct {
	Mgr ResourceSpec
	Mon ResourceSpec
	Osd ResourceSpec
}

type ResourceSpec struct {
	Limits   ResourceList
	Requests ResourceList
}

type ResourceList struct {
	CPU    string
	Memory string
}

// NewRookConfig creates a default Rook configuration
func NewRookConfig() *RookConfig {
	const version = "release-1.15"
	baseURL := fmt.Sprintf("https://raw.githubusercontent.com/rook/rook/%s/deploy/examples", version)

	cfg := &RookConfig{
		Version:   version,
		Namespace: "rook-ceph",
		BaseURL:   baseURL,
		Manifests: []string{
			"operator.yaml",
			"crds.yaml",
			"common.yaml",
			"toolbox.yaml",
		},
	}

	// Initialize cluster specification
	cfg.ClusterSpec = ClusterSpec{
		CephVersion: struct{ Image string }{
			Image: "ceph/ceph:v16.2.6",
		},
		DataDirHostPath: "/var/lib/rook",
		Mon: MonConfig{
			Count:                1,
			AllowMultiplePerNode: false,
		},
		Dashboard: DashboardConfig{
			Enabled: true,
		},
		Storage: StorageConfig{
			UseAllNodes:   true,
			UseAllDevices: false,
			Config: map[string]string{
				"storeType": "bluestore",
			},
			Nodes: []StorageNode{
				{
					Devices: []DeviceSpec{
						{Name: "/dev/sda"},
					},
				},
			},
		},
		Resources: ResourceRequirements{
			Mgr: ResourceSpec{
				Limits: ResourceList{
					CPU:    "500m",
					Memory: "1024Mi",
				},
				Requests: ResourceList{
					CPU:    "500m",
					Memory: "1024Mi",
				},
			},
			Mon: ResourceSpec{
				Limits: ResourceList{
					CPU:    "500m",
					Memory: "1024Mi",
				},
				Requests: ResourceList{
					CPU:    "500m",
					Memory: "1024Mi",
				},
			},
			Osd: ResourceSpec{
				Limits: ResourceList{
					CPU:    "500m",
					Memory: "2048Mi",
				},
				Requests: ResourceList{
					CPU:    "500m",
					Memory: "2048Mi",
				},
			},
		},
	}

	return cfg
}

// InstallRook installs Rook to k3s cluster.
func InstallRook() error {
	cfg := NewRookConfig()
	return cfg.Install()
}

// Install performs the Rook installation
func (r *RookConfig) Install() error {
	if err := r.createNamespace(); err != nil {
		return err
	}

	if err := r.applyManifests(); err != nil {
		return err
	}

	if err := r.waitForOperator(); err != nil {
		return err
	}

	return r.deployCluster()
}

// createNamespace creates the Rook namespace
func (r *RookConfig) createNamespace() error {
	cmd := exec.Command("kubectl", "create", "namespace", r.Namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create namespace %s: %w", r.Namespace, err)
	}

	return nil
}

// applyManifests applies all required Rook manifests
func (r *RookConfig) applyManifests() error {
	var urls []string
	for _, manifest := range r.Manifests {
		urls = append(urls, fmt.Sprintf(" -f %s/%s", r.BaseURL, manifest))
	}

	cmd := exec.Command("kubectl", append([]string{"apply", "-f"}, urls...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	return nil
}

// waitForOperator waits for the Rook operator to be ready
func (r *RookConfig) waitForOperator() error {
	log.Println("Waiting for Rook Operator to be in Running state")

	cmd := exec.Command("kubectl", "wait", "--for=condition=available",
		"deployment/rook-ceph-operator", "-n", r.Namespace, "--timeout=600s")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("timeout waiting for operator: %w", err)
	}

	return nil
}

// deployCluster deploys the Ceph cluster
func (r *RookConfig) deployCluster() error {
	clusterConfig := r.generateClusterConfig()

	configPath := "/tmp/rook-cluster.yaml"
	if err := os.WriteFile(configPath, []byte(clusterConfig), 0644); err != nil {
		return fmt.Errorf("failed to write cluster config: %w", err)
	}

	cmd := exec.Command("kubectl", "apply", "-f", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy cluster: %w", err)
	}

	return nil
}

// generateClusterConfig generates the Ceph cluster configuration
func (r *RookConfig) generateClusterConfig() string {
	return fmt.Sprintf(`
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: %s
spec:
  cephVersion:
    image: %s
  dataDirHostPath: %s
  mon:
    count: %d
    allowMultiplePerNode: %v
  dashboard:
    enabled: %v
  storage:
    useAllNodes: %v
    useAllDevices: %v
    config:
      storeType: %s
    nodes:
    - devices:
      - name: %s
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
        cpu: "%s"
        memory: "%s"
      requests:
        cpu: "%s"
        memory: "%s"
    mon:
      limits:
        cpu: "%s"
        memory: "%s"
      requests:
        cpu: "%s"
        memory: "%s"
    osd:
      limits:
        cpu: "%s"
        memory: "%s"
      requests:
        cpu: "%s"
        memory: "%s"
`,
		r.Namespace,
		r.ClusterSpec.CephVersion.Image,
		r.ClusterSpec.DataDirHostPath,
		r.ClusterSpec.Mon.Count,
		r.ClusterSpec.Mon.AllowMultiplePerNode,
		r.ClusterSpec.Dashboard.Enabled,
		r.ClusterSpec.Storage.UseAllNodes,
		r.ClusterSpec.Storage.UseAllDevices,
		r.ClusterSpec.Storage.Config["storeType"],
		r.ClusterSpec.Storage.Nodes[0].Devices[0].Name,
		r.ClusterSpec.Resources.Mgr.Limits.CPU,
		r.ClusterSpec.Resources.Mgr.Limits.Memory,
		r.ClusterSpec.Resources.Mgr.Requests.CPU,
		r.ClusterSpec.Resources.Mgr.Requests.Memory,
		r.ClusterSpec.Resources.Mon.Limits.CPU,
		r.ClusterSpec.Resources.Mon.Limits.Memory,
		r.ClusterSpec.Resources.Mon.Requests.CPU,
		r.ClusterSpec.Resources.Mon.Requests.Memory,
		r.ClusterSpec.Resources.Osd.Limits.CPU,
		r.ClusterSpec.Resources.Osd.Limits.Memory,
		r.ClusterSpec.Resources.Osd.Requests.CPU,
		r.ClusterSpec.Resources.Osd.Requests.Memory,
	)
}
