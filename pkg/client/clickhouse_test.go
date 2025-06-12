package client_test

import (
	"testing"
)

func TestCreateClickhouse(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreateClickhouse("test-clickhouse", testNamespace, 1)
	if err != nil {
		t.Fatalf("error creating clickhouse: %v", err)
	}
}

func TestGetClickhouse(t *testing.T) {
	cli := setupTestClient(t)
	clickhouse, err := cli.GetClickhouse("test-clickhouse", testNamespace)
	if err != nil {
		t.Fatalf("error getting clickhouse: %v", err)
	}
	t.Logf("Clickhouse: %v", clickhouse)
}

func TestListClickhouse(t *testing.T) {
	cli := setupTestClient(t)
	clickhouseClusters, err := cli.ListClickhouse(testNamespace)
	if err != nil {
		t.Fatalf("error listing clickhouse: %v", err)
	}
	t.Logf("Clickhouse: %v", clickhouseClusters)
}

func TestDeleteClickhouse(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteClickhouse("test-clickhouse", testNamespace)
	if err != nil {
		t.Fatalf("error deleting clickhouse: %v", err)
	}
}
