package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// TestListNodes tests the ListNodes method
func TestListNodes(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	nodes, err := cli.ListNodes()
	if err != nil {
		t.Fatalf("error listing nodes: %v", err)
	}
	t.Logf("nodes: %v", nodes)
}

// TestGetNode tests the GetNode method
func TestGetNode(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	node, err := cli.GetNode("node-10-0-0-2")
	if err != nil {
		t.Fatalf("error getting node: %v", err)
	}
	t.Logf("node: %v", node)
}

// TestDeleteNode tests the DeleteNode method
func TestDeleteNode(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.DeleteNode("node-10-0-0-2")
	if err != nil {
		t.Fatalf("error deleting node: %v", err)
	}
}

// TestAddNode tests the AddNode method
func TestAddNode(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.AddNode("node-10-0-0-2", "10.0.0.2", "10.0.0.1", "", "")
	if err != nil {
		t.Fatalf("error adding node: %v", err)
	}
	nodes, err := cli.ListNodes()
	if err != nil {
		t.Fatalf("error listing nodes: %v", err)
	}
	t.Logf("nodes: %v", nodes)
	found := false
	for _, node := range nodes {
		if node == "node-10-0-0-2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("node not found")
	}
}

// TestUpgradeNode tests the UpgradeNode method
func TestUpgradeNode(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.UpgradeNode("10.0.0.2")
	if err != nil {
		t.Fatalf("error upgrading node: %v", err)
	}
}
