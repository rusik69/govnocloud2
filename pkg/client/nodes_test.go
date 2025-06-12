package client_test

import (
	"testing"
	"time"
)

// TestListNodes tests the ListNodes method
func TestListNodes(t *testing.T) {
	cli := setupTestClient(t)
	nodes, err := cli.ListNodes()
	if err != nil {
		t.Fatalf("error listing nodes: %v", err)
	}
	t.Logf("Nodes: %v", nodes)
}

// TestGetNode tests the GetNode method
func TestGetNode(t *testing.T) {
	cli := setupTestClient(t)
	node, err := cli.GetNode("node-10-0-0-2")
	if err != nil {
		t.Fatalf("error getting node: %v", err)
	}
	t.Logf("Node: %v", node)
}

// TestDeleteNode tests the DeleteNode method
func TestDeleteNode(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteNode("node-10-0-0-2")
	if err != nil {
		t.Fatalf("error deleting node: %v", err)
	}
}

// TestAddNode tests the AddNode method
func TestAddNode(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.AddNode("node-10-0-0-2", "10.0.0.2", "10.0.0.1", "", "")
	if err != nil {
		t.Fatalf("error adding node: %v", err)
	}
	nodes, err := cli.ListNodes()
	if err != nil {
		t.Fatalf("error listing nodes: %v", err)
	}
	found := false
	for _, node := range nodes {
		if node == "node-10-0-0-2" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("node not found after adding")
	}
	time.Sleep(10 * time.Second)
}

// TestUpgradeNode tests the UpgradeNode method
func TestUpgradeNode(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.UpgradeNode("node-10-0-0-2")
	if err != nil {
		t.Fatalf("error upgrading node: %v", err)
	}
}
