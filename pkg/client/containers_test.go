package client_test

import (
	"testing"
)

// TestCreateContainer tests the CreateContainer function.
func TestCreateContainer(t *testing.T) {
	err := cli.CreateContainer("test-container", "k8s.gcr.io/pause", testNamespace, 1024, 1024, 1024, 80)
	if err != nil {
		t.Fatalf("error creating container: %v", err)
	}
}

// TestListContainers tests the ListContainers function.
func TestListContainers(t *testing.T) {
	containers, err := cli.ListContainers(testNamespace)
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}
	t.Logf("Containers: %v", containers)
}

// TestGetContainer tests the GetContainer function.
func TestGetContainer(t *testing.T) {
	container, err := cli.GetContainer("test-container", testNamespace)
	if err != nil {
		t.Fatalf("error getting container: %v", err)
	}
	t.Logf("Container: %v", container)
}

// TestDeleteContainer tests the DeleteContainer function.
func TestDeleteContainer(t *testing.T) {
	err := cli.DeleteContainer("test-container", testNamespace)
	if err != nil {
		t.Fatalf("error deleting container: %v", err)
	}
}
