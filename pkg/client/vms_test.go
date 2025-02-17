package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
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
	if len(vms) == 0 {
		t.Fatalf("no VMs found")
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

func TestStopVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.StopVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error stopping VM: %v", err)
	}
}

func TestStartVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.StartVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error starting VM: %v", err)
	}
}

func TestRestartVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.RestartVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error restarting VM: %v", err)
	}
}

func TestDeleteVM(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.DeleteVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error deleting VM: %v", err)
	}
}
