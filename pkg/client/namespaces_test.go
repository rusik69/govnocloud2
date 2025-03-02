package client_test

import (
	"testing"
)

func TestCreateNamespace(t *testing.T) {
	err := cli.CreateNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}
	t.Logf("namespace created")
}

func TestGetNamespace(t *testing.T) {
	namespace, err := cli.GetNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error getting namespace: %v", err)
	}
	t.Logf("namespace: %v", namespace)
}

func TestListNamespaces(t *testing.T) {
	namespaces, err := cli.ListNamespaces()
	if err != nil {
		t.Fatalf("error listing namespaces: %v", err)
	}
	t.Logf("namespaces: %v", namespaces)
}

func TestDeleteNamespace(t *testing.T) {
	err := cli.DeleteNamespace(testNamespace2)
	if err != nil {
		t.Fatalf("error deleting namespace: %v", err)
	}
	t.Logf("namespace deleted")
}
