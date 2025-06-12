package client_test

import (
	"testing"
)

func TestCreateMysql(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreateMysql("test-mysql", testNamespace, 1, 1)
	if err != nil {
		t.Fatalf("error creating mysql: %v", err)
	}
}

func TestGetMysql(t *testing.T) {
	cli := setupTestClient(t)
	mysql, err := cli.GetMysql("test-mysql", testNamespace)
	if err != nil {
		t.Fatalf("error getting mysql: %v", err)
	}
	t.Logf("MySQL: %v", mysql)
}

func TestListMysql(t *testing.T) {
	cli := setupTestClient(t)
	mysqlClusters, err := cli.ListMysql(testNamespace)
	if err != nil {
		t.Fatalf("error listing mysql: %v", err)
	}
	t.Logf("MySQL: %v", mysqlClusters)
}

func TestDeleteMysql(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteMysql("test-mysql", testNamespace)
	if err != nil {
		t.Fatalf("error deleting mysql: %v", err)
	}
}
