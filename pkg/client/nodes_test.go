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
