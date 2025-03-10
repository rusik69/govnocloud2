package client_test

import (
	"testing"
)

func TestCreateClickhouse(t *testing.T) {
	err := cli.CreateClickhouse("test-clickhouse", testNamespace, 1)
	if err != nil {
		t.Fatalf("error creating clickhouse cluster: %v", err)
	}
}

func TestGetClickhouse(t *testing.T) {
	clickhouse, err := cli.GetClickhouse("test-clickhouse", testNamespace)
	if err != nil {
		t.Fatalf("error getting clickhouse cluster: %v", err)
	}
	t.Logf("clickhouse cluster: %v", clickhouse)
}

func TestListClickhouse(t *testing.T) {
	clickhouseClusters, err := cli.ListClickhouse(testNamespace)
	if err != nil {
		t.Fatalf("error listing clickhouse clusters: %v", err)
	}
	t.Logf("clickhouse clusters: %v", clickhouseClusters)
}

func TestDeleteClickhouse(t *testing.T) {
	err := cli.DeleteClickhouse("test-clickhouse", testNamespace)
	if err != nil {
		t.Fatalf("error deleting clickhouse cluster: %v", err)
	}
}
