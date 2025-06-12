package client_test

import (
	"testing"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// User test variables
// These are already defined in containers_test.go:
// var cli *client.Client
// const testNamespace = "test"

// User-specific test constants
const (
	testNewUser     = "testuser"
	testNewPassword = "testpassword"
)

func TestCreateUser(t *testing.T) {
	cli := setupTestClient(t)
	user := types.User{
		Name:       testNewUser,
		Password:   testNewPassword,
		Namespaces: []string{testNamespace},
	}
	err := cli.CreateUser(testNewUser, user)
	if err != nil {
		t.Fatalf("error creating user: %v", err)
	}
	t.Logf("user created")
}

func TestGetUser(t *testing.T) {
	cli := setupTestClient(t)
	user, err := cli.GetUser(testNewUser)
	if err != nil {
		t.Fatalf("error getting user: %v", err)
	}
	t.Logf("user: %v", user)
}

func TestListUsers(t *testing.T) {
	cli := setupTestClient(t)
	users, err := cli.ListUsers()
	if err != nil {
		t.Fatalf("error listing users: %v", err)
	}
	t.Logf("users: %v", users)
}

func TestSetUserPassword(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.SetUserPassword(testNewUser, testNewPassword)
	if err != nil {
		t.Fatalf("error setting user password: %v", err)
	}
	t.Logf("user password set")
}

func TestAddNamespaceToUser(t *testing.T) {
	cli := setupTestClient(t)
	// Using the testNamespace variable defined in containers_test.go
	err := cli.AddNamespaceToUser(testNewUser, testNamespace)
	if err != nil {
		t.Fatalf("error adding namespace to user: %v", err)
	}
	t.Logf("namespace added to user")
}

func TestRemoveNamespaceFromUser(t *testing.T) {
	cli := setupTestClient(t)
	// Using the testNamespace variable defined in containers_test.go
	err := cli.RemoveNamespaceFromUser(testNewUser, testNamespace)
	if err != nil {
		t.Fatalf("error removing namespace from user: %v", err)
	}
	t.Logf("namespace removed from user")
}

func TestDeleteUser(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.DeleteUser(testNewUser)
	if err != nil {
		t.Fatalf("error deleting user: %v", err)
	}
	t.Logf("user deleted")
}
