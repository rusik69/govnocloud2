package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// TestCreateVolume tests the CreateVolume function
func TestCreateVolume(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	err := client.CreateVolume("test", testNamespace, "1Gi")
	if err != nil {
		t.Fatalf("error creating volume: %v", err)
	}
	t.Logf("volume created")
}

// TestGetVolume tests the GetVolume function
func TestGetVolume(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	volume, err := client.GetVolume("test", testNamespace)
	if err != nil {
		t.Fatalf("error getting volume: %v", err)
	}
	t.Logf("volume: %v", volume)
}

// TestListVolumes tests the ListVolumes function
func TestListVolumes(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	volumes, err := client.ListVolumes(testNamespace)
	if err != nil {
		t.Fatalf("error listing volumes: %v", err)
	}
	t.Logf("volumes: %v", volumes)
}

// TestDeleteVolume tests the DeleteVolume function
func TestDeleteVolume(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	err := client.DeleteVolume("test", testNamespace)
	if err != nil {
		t.Fatalf("error deleting volume: %v", err)
	}
	t.Logf("volume deleted")
}
