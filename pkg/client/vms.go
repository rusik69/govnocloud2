package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateVM creates a VM.
func CreateVM(host, port, name, image, size, namespace string) error {
	url := fmt.Sprintf("http://%s:%s/api/v1/vms", host, port)
	vm := types.VM{
		Name:      name,
		Image:     image,
		Size:      size,
		Namespace: namespace,
	}
	data, err := json.Marshal(vm)
	if err != nil {
		return fmt.Errorf("error marshaling VM: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating VM: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error creating VM: %s", resp.Status)
	}
	return nil
}

// ListVMs lists VMs.
func ListVMs(host, port string) ([]types.VM, error) {
	url := fmt.Sprintf("http://%s:%s/api/v1/vms", host, port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing VMs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error listing VMs: %s", resp.Status)
	}
	var vms []types.VM
	if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
		return nil, fmt.Errorf("error decoding VMs: %w", err)
	}
	return vms, nil
}

// GetVM gets a VM.
func GetVM(host, port, name string) (*types.VM, error) {
	url := fmt.Sprintf("http://%s:%s/api/v1/vms/?name=%s", host, port, name)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting VM: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting VM: %s", resp.Status)
	}
	var vm types.VM
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return nil, fmt.Errorf("error decoding VM: %w", err)
	}
	return &vm, nil
}

// DeleteVM deletes a VM.
func DeleteVM(host, port, name string) error {
	url := fmt.Sprintf("http://%s:%s/api/v1/vms/%s", host, port, name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting VM: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error deleting VM: %s", resp.Status)
	}
	return nil
}
