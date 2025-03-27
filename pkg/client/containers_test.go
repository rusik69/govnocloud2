package client_test

import (
	"log"
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

var cli *client.Client

func init() {
	cli = client.NewClient(testHost, testPort, testUser, testPassword)
	err := cli.CreateNamespace(testNamespace)
	if err != nil {
		log.Fatalf("error creating namespace: %v", err)
	}
}

const (
	testHost       = "localhost"
	testPort       = "6969"
	testNamespace  = "test"
	testNamespace2 = "test2"
	testUser       = "root"
	testPassword   = "password"
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
