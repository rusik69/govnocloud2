package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

func TestCreateNamespace(t *testing.T) {
	client := client.NewClient("http://localhost:6969")
	err := client.CreateNamespace("test")
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}
	t.Logf("namespace created")
}

func TestGetNamespace(t *testing.T) {
	client := client.NewClient("http://localhost:6969")
	namespace, err := client.GetNamespace("test")
	if err != nil {
		t.Fatalf("error getting namespace: %v", err)
	}
	t.Logf("namespace: %v", namespace)
}

func TestListNamespaces(t *testing.T) {
	client := client.NewClient("http://localhost:6969")
	namespaces, err := client.ListNamespaces()
	if err != nil {
		t.Fatalf("error listing namespaces: %v", err)
	}
	t.Logf("namespaces: %v", namespaces)
}

func TestDeleteNamespace(t *testing.T) {
	client := client.NewClient("http://localhost:6969")
	err := client.DeleteNamespace("test")
	if err != nil {
		t.Fatalf("error deleting namespace: %v", err)
	}
	t.Logf("namespace deleted")
}
