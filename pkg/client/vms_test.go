package client_test

import (
	"testing"
	"time"

	"github.com/rusik69/govnocloud2/pkg/client"
)

const (
	testHost      = "localhost"
	testPort      = "6969"
	testNamespace = "default"
)

func setupTestClient() *client.Client {
	return client.NewClient(testHost, testPort)
}

func TestVMOperations(t *testing.T) {
	c := setupTestClient()

	// Test VM parameters
	testVM := struct {
		name      string
		image     string
		size      string
		namespace string
	}{
		name:      "test-vm-" + time.Now().Format("20060102150405"),
		image:     "ubuntu:24.04",
		size:      "small",
		namespace: testNamespace,
	}

	// Test CreateVM
	t.Run("CreateVM", func(t *testing.T) {
		err := c.CreateVM(testVM.name, testVM.image, testVM.size, testVM.namespace)
		if err != nil {
			t.Fatalf("CreateVM failed: %v", err)
		}

		// Wait for VM to be created
		time.Sleep(5 * time.Second)
	})

	// Test ListVMs
	t.Run("ListVMs", func(t *testing.T) {
		vms, err := c.ListVMs()
		if err != nil {
			t.Fatalf("ListVMs failed: %v", err)
		}

		found := false
		for _, vm := range vms {
			if vm.Name == testVM.name {
				found = true
				if vm.Image != testVM.image {
					t.Errorf("expected image %s, got %s", testVM.image, vm.Image)
				}
				if vm.Size != testVM.size {
					t.Errorf("expected size %s, got %s", testVM.size, vm.Size)
				}
				if vm.Namespace != testVM.namespace {
					t.Errorf("expected namespace %s, got %s", testVM.namespace, vm.Namespace)
				}
				break
			}
		}
		if !found {
			t.Errorf("created VM %s not found in list", testVM.name)
		}
	})

	// Test GetVM
	t.Run("GetVM", func(t *testing.T) {
		vm, err := c.GetVM(testVM.name)
		if err != nil {
			t.Fatalf("GetVM failed: %v", err)
		}

		if vm.Name != testVM.name {
			t.Errorf("expected name %s, got %s", testVM.name, vm.Name)
		}
		if vm.Image != testVM.image {
			t.Errorf("expected image %s, got %s", testVM.image, vm.Image)
		}
		if vm.Size != testVM.size {
			t.Errorf("expected size %s, got %s", testVM.size, vm.Size)
		}
		if vm.Namespace != testVM.namespace {
			t.Errorf("expected namespace %s, got %s", testVM.namespace, vm.Namespace)
		}
	})

	// Test VM Status
	t.Run("CheckVMStatus", func(t *testing.T) {
		maxAttempts := 10
		interval := 5 * time.Second

		var lastStatus string
		for i := 0; i < maxAttempts; i++ {
			vm, err := c.GetVM(testVM.name)
			if err != nil {
				t.Logf("attempt %d: GetVM failed: %v", i+1, err)
				time.Sleep(interval)
				continue
			}

			lastStatus = vm.Status
			if vm.Status == "Running" {
				break
			}

			if i == maxAttempts-1 {
				t.Errorf("VM did not reach Running state, last status: %s", lastStatus)
			}

			time.Sleep(interval)
		}
	})

	// Test DeleteVM
	t.Run("DeleteVM", func(t *testing.T) {
		err := c.DeleteVM(testVM.name)
		if err != nil {
			t.Fatalf("DeleteVM failed: %v", err)
		}

		// Verify VM is deleted
		time.Sleep(5 * time.Second)
		vm, err := c.GetVM(testVM.name)
		if err == nil {
			t.Errorf("VM still exists after deletion: %+v", vm)
		}
	})

	// Test error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Test getting non-existent VM
		t.Run("GetNonExistentVM", func(t *testing.T) {
			_, err := c.GetVM("non-existent-vm")
			if err == nil {
				t.Error("expected error when getting non-existent VM")
			}
		})

		// Test creating VM with invalid parameters
		t.Run("CreateInvalidVM", func(t *testing.T) {
			err := c.CreateVM("", "", "", "")
			if err == nil {
				t.Error("expected error when creating VM with invalid parameters")
			}
		})

		// Test deleting non-existent VM
		t.Run("DeleteNonExistentVM", func(t *testing.T) {
			err := c.DeleteVM("non-existent-vm")
			if err == nil {
				t.Error("expected error when deleting non-existent VM")
			}
		})
	})
}

func TestVMValidation(t *testing.T) {
	c := setupTestClient()

	tests := []struct {
		name        string
		vmName      string
		image       string
		size        string
		namespace   string
		expectError bool
	}{
		{
			name:        "valid parameters",
			vmName:      "test-vm",
			image:       "ubuntu:20.04",
			size:        "small",
			namespace:   "default",
			expectError: false,
		},
		{
			name:        "empty name",
			vmName:      "",
			image:       "ubuntu:20.04",
			size:        "small",
			namespace:   "default",
			expectError: true,
		},
		{
			name:        "empty image",
			vmName:      "test-vm",
			image:       "",
			size:        "small",
			namespace:   "default",
			expectError: true,
		},
		{
			name:        "invalid size",
			vmName:      "test-vm",
			image:       "ubuntu:20.04",
			size:        "invalid",
			namespace:   "default",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.CreateVM(tt.vmName, tt.image, tt.size, tt.namespace)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateVM() error = %v, expectError %v", err, tt.expectError)
			}

			if err == nil {
				// Cleanup if VM was created
				defer func() {
					if err := c.DeleteVM(tt.vmName); err != nil {
						t.Logf("cleanup failed: %v", err)
					}
				}()
			}
		})
	}
}
