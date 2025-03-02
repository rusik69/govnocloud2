package client_test

import (
	"testing"
)

func TestCreateDB(t *testing.T) {
	err := cli.CreateDB("test-db", testNamespace, "postgres", "small")
	if err != nil {
		t.Fatalf("error creating DB: %v", err)
	}
}

func TestGetDB(t *testing.T) {
	db, err := cli.GetDB("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error getting DB: %v", err)
	}
	t.Logf("DB: %v", db)
}

func TestListDBs(t *testing.T) {
	dbs, err := cli.ListDBs(testNamespace)
	if err != nil {
		t.Fatalf("error listing DBs: %v", err)
	}
	t.Logf("DBs: %v", dbs)
}

func TestDeleteDB(t *testing.T) {
	err := cli.DeleteDB("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error deleting DB: %v", err)
	}
}
