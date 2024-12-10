package k3s

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// RookConfig holds configuration for Rook installation
type RookConfig struct {
	Version     string
	Namespace   string
	BaseURL     string
	Manifests   []string
	ClusterSpec ClusterSpec
	Host        string
	User        string
	Key         string
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
func NewRookConfig(host, user, key string) *RookConfig {
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
		Host: host,
		User: user,
		Key:  key,
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
func InstallRook(host, user, key string) error {
	cfg := NewRookConfig(host, user, key)
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
	cmd := fmt.Sprintf("kubectl create namespace %s", r.Namespace)
	log.Println(cmd)
	out, err := ssh.Run(cmd, r.Host, r.Key, r.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	log.Println(out)
	return nil
}

// applyManifests applies all required Rook manifests
func (r *RookConfig) applyManifests() error {
	var urls []string
	for _, manifest := range r.Manifests {
		urls = append(urls, fmt.Sprintf(" -f %s/%s", r.BaseURL, manifest))
	}
	cmd := fmt.Sprintf("kubectl apply %s", strings.Join(urls, " "))
	log.Println(cmd)
	out, err := ssh.Run(cmd, r.Host, r.Key, r.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}
	log.Println(out)
	return nil
}

// waitForOperator waits for the Rook operator to be ready
func (r *RookConfig) waitForOperator() error {
	log.Println("Waiting for Rook Operator to be in Running state")
	cmd := fmt.Sprintf("kubectl wait --for=condition=available deployment/rook-ceph-operator -n %s --timeout=600s", r.Namespace)
	log.Println(cmd)
	out, err := ssh.Run(cmd, r.Host, r.Key, r.User, "", true, 600)
	if err != nil {
		return fmt.Errorf("failed to wait for operator: %w", err)
	}
	log.Println(out)
	return nil
}

// deployCluster deploys the Ceph cluster
func (r *RookConfig) deployCluster() error {
	clusterConfig := r.generateClusterConfig()
	log.Println(clusterConfig)
	configPath := "/tmp/rook-cluster.yaml"
	if err := os.WriteFile(configPath, []byte(clusterConfig), 0644); err != nil {
		return fmt.Errorf("failed to write cluster config: %w", err)
	}
	err := ssh.Copy(configPath, "new-"+configPath, r.Host, r.User, r.Key)
	if err != nil {
		return fmt.Errorf("failed to copy cluster config: %w", err)
	}
	cmd := fmt.Sprintf("kubectl apply -f %s", "new-"+configPath)
	log.Println(cmd)
	out, err := ssh.Run(cmd, r.Host, r.Key, r.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to deploy cluster: %w", err)
	}
	log.Println(out)
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
