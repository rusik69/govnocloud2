package client_test

import (
	"testing"
)

// TestCreateVolume tests the CreateVolume function
func TestCreateVolume(t *testing.T) {
	err := cli.CreateVolume("test", testNamespace, "1Gi")
	if err != nil {
		t.Fatalf("error creating volume: %v", err)
	}
	t.Logf("volume created")
}

// TestGetVolume tests the GetVolume function
func TestGetVolume(t *testing.T) {
	volume, err := cli.GetVolume("test", testNamespace)
	if err != nil {
		t.Fatalf("error getting volume: %v", err)
	}
	t.Logf("volume: %v", volume)
}

// TestListVolumes tests the ListVolumes function
func TestListVolumes(t *testing.T) {
	volumes, err := cli.ListVolumes(testNamespace)
	if err != nil {
		t.Fatalf("error listing volumes: %v", err)
	}
	t.Logf("volumes: %v", volumes)
}

// TestDeleteVolume tests the DeleteVolume function
func TestDeleteVolume(t *testing.T) {
	err := cli.DeleteVolume("test", testNamespace)
	if err != nil {
		t.Fatalf("error deleting volume: %v", err)
	}
	t.Logf("volume deleted")
}
