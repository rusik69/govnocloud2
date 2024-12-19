package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

const (
	testHost      = "localhost"
	testPort      = "6969"
	testNamespace = "default"
)

func TestCreateVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.CreateVM("test-vm", "ubuntu24", "small", testNamespace)
	if err != nil {
		t.Fatalf("error creating VM: %v", err)
	}
}

func TestListVMs(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	vms, err := cli.ListVMs(testNamespace)
	if err != nil {
		t.Fatalf("error listing VMs: %v", err)
	}
	t.Logf("VMs: %v", vms)
}

func TestGetVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	vm, err := cli.GetVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error getting VM: %v", err)
	}
	t.Logf("VM: %v", vm)
}

func TestDeleteVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.DeleteVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error deleting VM: %v", err)
	}
}
