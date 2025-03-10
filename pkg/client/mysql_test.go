package client_test

import (
	"testing"
)

func TestCreateMysql(t *testing.T) {
	err := cli.CreateMysql("test-mysql", testNamespace, 1, 1)
	if err != nil {
		t.Fatalf("error creating mysql cluster: %v", err)
	}
}

func TestGetMysql(t *testing.T) {
	mysql, err := cli.GetMysql("test-mysql", testNamespace)
	if err != nil {
		t.Fatalf("error getting mysql cluster: %v", err)
	}
	t.Logf("mysql cluster: %v", mysql)
}

func TestListMysql(t *testing.T) {
	mysqlClusters, err := cli.ListMysql(testNamespace)
	if err != nil {
		t.Fatalf("error listing mysql clusters: %v", err)
	}
	t.Logf("mysql clusters: %v", mysqlClusters)
}

func TestDeleteMysql(t *testing.T) {
	err := cli.DeleteMysql("test-mysql", testNamespace)
	if err != nil {
		t.Fatalf("error deleting mysql cluster: %v", err)
	}
}
