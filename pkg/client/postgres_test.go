package client_test

import (
	"testing"
)

func TestCreatePostgres(t *testing.T) {
	err := cli.CreatePostgres("test-db", testNamespace, "small", 1, 1)
	if err != nil {
		t.Fatalf("error creating postgres cluster: %v", err)
	}
}

func TestGetPostgres(t *testing.T) {
	db, err := cli.GetPostgres("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error getting postgres cluster: %v", err)
	}
	t.Logf("postgres cluster: %v", db)
}

func TestListPostgres(t *testing.T) {
	dbs, err := cli.ListPostgres(testNamespace)
	if err != nil {
		t.Fatalf("error listing postgres clusters: %v", err)
	}
	t.Logf("postgres clusters: %v", dbs)
}

func TestDeletePostgres(t *testing.T) {
	err := cli.DeletePostgres("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error deleting postgres cluster: %v", err)
	}
}
