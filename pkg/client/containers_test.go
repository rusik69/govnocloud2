package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// TestCreateContainer tests the CreateContainer function.
func TestCreateContainer(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.CreateContainer("test-container", "nginx", "default")
	if err != nil {
		t.Fatalf("error creating container: %v", err)
	}
}

// TestListContainers tests the ListContainers function.
func TestListContainers(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	containers, err := cli.ListContainers()
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}
	t.Logf("Containers: %v", containers)
}

// TestGetContainer tests the GetContainer function.
func TestGetContainer(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	container, err := cli.GetContainer("test-container")
	if err != nil {
		t.Fatalf("error getting container: %v", err)
	}
	t.Logf("Container: %v", container)
}

// TestDeleteContainer tests the DeleteContainer function.
func TestDeleteContainer(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.DeleteContainer("test-container")
	if err != nil {
		t.Fatalf("error deleting container: %v", err)
	}
}
