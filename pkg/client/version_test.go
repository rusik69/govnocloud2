package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

var serverHost = "10.0.0.1"
var serverPort = "6969"

// TestVersion tests the GetServerVersion function.
func TestVersion(t *testing.T) {
	ver, err := client.GetServerVersion(serverHost, serverPort)
	if err != nil {
		t.Error(err)
	}
	if ver != "v0.0.1" {
		t.Error("expected v0.0.1, got ", ver)
	}
}
