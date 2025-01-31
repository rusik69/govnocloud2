package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

func TestCreateNamespace(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	err := client.CreateNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}
	t.Logf("namespace created")
}

func TestGetNamespace(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	namespace, err := client.GetNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error getting namespace: %v", err)
	}
	t.Logf("namespace: %v", namespace)
}

func TestListNamespaces(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	namespaces, err := client.ListNamespaces()
	if err != nil {
		t.Fatalf("error listing namespaces: %v", err)
	}
	t.Logf("namespaces: %v", namespaces)
}

func TestDeleteNamespace(t *testing.T) {
	client := client.NewClient(testHost, testPort)
	err := client.DeleteNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error deleting namespace: %v", err)
	}
	t.Logf("namespace deleted")
}
