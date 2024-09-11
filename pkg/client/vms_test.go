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
