package client_test

import (
	"testing"
	"time"
)

func TestCreateVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreateVM("test-vm", "ubuntu24", "small", testNamespace)
	if err != nil {
		t.Fatalf("error creating VM: %v", err)
	}
}

func TestWaitVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.WaitVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error waiting for VM: %v", err)
	}
}

func TestListVMs(t *testing.T) {
	cli := setupTestClient(t)
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
	cli := setupTestClient(t)
	vm, err := cli.GetVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error getting VM: %v", err)
	}
	t.Logf("VM: %v", vm)
}

func TestStopVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.StopVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error stopping VM: %v", err)
	}
	time.Sleep(10 * time.Second)
}

func TestStartVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.StartVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error starting VM: %v", err)
	}
}

func TestRestartVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.RestartVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error restarting VM: %v", err)
	}
}

func TestDeleteVM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteVM("test-vm", testNamespace)
	if err != nil {
		t.Fatalf("error deleting VM: %v", err)
	}
}
