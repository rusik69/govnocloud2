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
