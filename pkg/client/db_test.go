package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

func TestCreateDB(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.CreateDB("test-db", "postgres", "small", testNamespace)
	if err != nil {
		t.Fatalf("error creating DB: %v", err)
	}
}

func TestGetDB(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	db, err := cli.GetDB("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error getting DB: %v", err)
	}
	t.Logf("DB: %v", db)
}

func TestListDBs(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	dbs, err := cli.ListDBs(testNamespace)
	if err != nil {
		t.Fatalf("error listing DBs: %v", err)
	}
	t.Logf("DBs: %v", dbs)
}

func TestDeleteDB(t *testing.T) {
	cli := client.NewClient(testHost, testPort)
	err := cli.DeleteDB("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error deleting DB: %v", err)
	}
}
