package client_test

import (
	"os"
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// Test configuration constants
const (
	testHost       = "localhost"
	testPort       = "6969"
	testNamespace  = "test"
	testNamespace2 = "test2"
	testUser       = "root"
	testPassword   = "password"
)

// setupTestClient initializes test client and setup test namespace
func setupTestClient(t *testing.T) *client.Client {
	// Check if integration tests should be skipped
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		t.Skip("Skipping integration test (SKIP_INTEGRATION_TESTS=1)")
	}

	cli := client.NewClient(testHost, testPort, testUser, testPassword)

	// Try to create test namespace, skip test if server is not available
	err := cli.CreateNamespace(testNamespace)
	if err != nil {
		t.Skipf("Skipping test: server not available at %s:%s - %v", testHost, testPort, err)
	}

	return cli
}
