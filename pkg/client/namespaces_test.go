package client_test

import (
	"testing"
)

func TestCreateNamespace(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreateNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}
	t.Logf("namespace created")
}

func TestGetNamespace(t *testing.T) {
	cli := setupTestClient(t)
	namespace, err := cli.GetNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error getting namespace: %v", err)
	}
	t.Logf("namespace: %v", namespace)
}

func TestListNamespaces(t *testing.T) {
	cli := setupTestClient(t)
	namespaces, err := cli.ListNamespaces()
	if err != nil {
		t.Fatalf("error listing namespaces: %v", err)
	}
	t.Logf("namespaces: %v", namespaces)
}

func TestDeleteNamespace(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error deleting namespace: %v", err)
	}
	t.Logf("namespace deleted")
}
