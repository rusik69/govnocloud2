package client_test

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/client"
)

// Shared test client instance
var cli *client.Client

// Test configuration constants
const (
	testHost       = "localhost"
	testPort       = "6969"
	testNamespace  = "test"
	testNamespace2 = "test2"
	testUser       = "root"
	testPassword   = "password"
)

// Initialize test client and setup test namespace
func init() {
	cli = client.NewClient(testHost, testPort, testUser, testPassword)
	err := cli.CreateNamespace(testNamespace)
	if err != nil {
		log.Fatalf("error creating namespace: %v", err)
	}
}
