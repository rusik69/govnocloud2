package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// TestCreateVM tests the CreateVM function.
func TestCreateVM(t *testing.T) {
	err := client.CreateVM(serverHost, serverPort, "test", "ubuntu", "small", "default")
	if err != nil {
		t.Error(err)
	}
}

// TestListVMs tests the ListVMs function.
func TestListVMs(t *testing.T) {
	vms, err := client.ListVMs(serverHost, serverPort)
	if err != nil {
		t.Error(err)
	}
	if len(vms) == 0 {
		t.Error("no vms found")
	}
}

// TestGetVM tests the GetVM function.
func TestGetVM(t *testing.T) {
	vm, err := client.GetVM(serverHost, serverPort, "test")
	if err != nil {
		t.Error(err)
	}
	if vm.Name != "test" {
		t.Error("invalid vm name")
	}
}

// TestDeleteVM tests the DeleteVM function.
func TestDeleteVM(t *testing.T) {
	err := client.DeleteVM(serverHost, serverPort, "test")
	if err != nil {
		t.Error(err)
	}
}
