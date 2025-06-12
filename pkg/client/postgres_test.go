package client_test

import (
	"testing"
)

func TestCreatePostgres(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreatePostgres("test-db", testNamespace, "small", 1, 1)
	if err != nil {
		t.Fatalf("error creating postgres: %v", err)
	}
}

func TestGetPostgres(t *testing.T) {
	cli := setupTestClient(t)
	db, err := cli.GetPostgres("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error getting postgres: %v", err)
	}
	t.Logf("Postgres: %v", db)
}

func TestListPostgres(t *testing.T) {
	cli := setupTestClient(t)
	dbs, err := cli.ListPostgres(testNamespace)
	if err != nil {
		t.Fatalf("error listing postgres: %v", err)
	}
	t.Logf("Postgres: %v", dbs)
}

func TestDeletePostgres(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeletePostgres("test-db", testNamespace)
	if err != nil {
		t.Fatalf("error deleting postgres: %v", err)
	}
}
