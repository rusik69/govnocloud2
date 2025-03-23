package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/crypto/bcrypt"
)

// UserManager handles user operations
type UserManager struct {
	etcdClient *clientv3.Client
}

// NewUserManager creates a new UserHandler
func NewUserManager() *UserManager {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to connect to etcd: %v", err)
	}
	return &UserManager{etcdClient: etcdClient}
}

// GetUser gets a user by name
func (m *UserManager) GetUser(name string) (*types.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/users/%s", name)
	resp, err := m.etcdClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, nil // User not found
	}

	var user types.User
	if err := json.Unmarshal(resp.Kvs[0].Value, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (m *UserManager) CreateUser(name string, user types.User) error {
	// Set the name to ensure consistency
	user.Name = name

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	existingUser, err := m.GetUser(name)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return fmt.Errorf("user already exists")
	}

	// Validate namespaces
	for _, ns := range user.Namespaces {
		if _, ok := types.ReservedNamespaces[ns]; ok {
			return fmt.Errorf("namespace %s is reserved", ns)
		}
	}

	// Hash password if provided
	password := user.Password
	user.Password = "" // Don't store password in user data

	// Store user data without password
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	key := fmt.Sprintf("/users/%s", name)
	_, err = m.etcdClient.Put(ctx, key, string(userData))
	if err != nil {
		return fmt.Errorf("failed to store user in etcd: %w", err)
	}

	// Store password separately if provided
	if password != "" {
		// Hash the password before storing
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		err = m.SetUserPassword(name, string(hashedPassword))
		if err != nil {
			return fmt.Errorf("failed to set user password: %w", err)
		}
	}

	return nil
}

// DeleteUser deletes a user by name
func (m *UserManager) DeleteUser(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete both user data and password
	userKey := fmt.Sprintf("/users/%s", name)

	// Use DeletePrefix to also delete any nested keys (like password)
	_, err := m.etcdClient.Delete(ctx, userKey, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to delete user from etcd: %w", err)
	}

	return nil
}

// ListUsers returns a list of all users
func (m *UserManager) ListUsers() ([]types.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := "/users/"
	resp, err := m.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list users from etcd: %w", err)
	}

	users := []types.User{}
	for _, kv := range resp.Kvs {
		// Skip password entries
		if len(kv.Key) > 9 && string(kv.Key[len(kv.Key)-9:]) == "/password" {
			continue
		}

		var user types.User
		if err := json.Unmarshal(kv.Value, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserPassword gets a user's password (hashed)
func (m *UserManager) GetUserPassword(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/users/%s/password", name)
	resp, err := m.etcdClient.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get user password from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("password not found for user %s", name)
	}

	return string(resp.Kvs[0].Value), nil
}

// SetUserPassword sets a user's password
func (m *UserManager) SetUserPassword(name, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify user exists
	user, err := m.GetUser(name)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user %s not found", name)
	}

	// Store password
	key := fmt.Sprintf("/users/%s/password", name)
	_, err = m.etcdClient.Put(ctx, key, password)
	if err != nil {
		return fmt.Errorf("failed to store user password in etcd: %w", err)
	}

	return nil
}

// AddNamespaceToUser adds a namespace to a user's list of accessible namespaces
func (m *UserManager) AddNamespaceToUser(name, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify user exists
	user, err := m.GetUser(name)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user %s not found", name)
	}

	// Check if namespace is reserved
	if _, ok := types.ReservedNamespaces[namespace]; ok {
		return fmt.Errorf("namespace %s is reserved", namespace)
	}

	// Check if namespace already exists in user's list
	for _, ns := range user.Namespaces {
		if ns == namespace {
			return nil // Namespace already added
		}
	}

	// Add namespace to user
	user.Namespaces = append(user.Namespaces, namespace)

	// Store updated user data
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	key := fmt.Sprintf("/users/%s", name)
	_, err = m.etcdClient.Put(ctx, key, string(userData))
	if err != nil {
		return fmt.Errorf("failed to update user in etcd: %w", err)
	}

	return nil
}

// RemoveNamespaceFromUser removes a namespace from a user's list of accessible namespaces
func (m *UserManager) RemoveNamespaceFromUser(name, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify user exists
	user, err := m.GetUser(name)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user %s not found", name)
	}

	// Remove namespace from user's list
	found := false
	updatedNamespaces := []string{}
	for _, ns := range user.Namespaces {
		if ns != namespace {
			updatedNamespaces = append(updatedNamespaces, ns)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("namespace %s not found in user's namespaces", namespace)
	}

	user.Namespaces = updatedNamespaces

	// Store updated user data
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	key := fmt.Sprintf("/users/%s", name)
	_, err = m.etcdClient.Put(ctx, key, string(userData))
	if err != nil {
		return fmt.Errorf("failed to update user in etcd: %w", err)
	}

	return nil
}

// VerifyPassword checks if the provided password matches the stored hash
func (m *UserManager) VerifyPassword(name, password string) (bool, error) {
	hashedPassword, err := m.GetUserPassword(name)
	if err != nil {
		return false, fmt.Errorf("failed to get stored password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("error comparing passwords: %w", err)
	}

	return true, nil
}

// CheckAuth verifies user authentication from request headers
func CheckAuth(c *gin.Context) (bool, string, error) {
	username := c.GetHeader("User")
	password := c.GetHeader("Password")

	if username == "" || password == "" {
		return false, "", fmt.Errorf("missing authentication headers")
	}

	// Verify the password against stored hash
	valid, err := userManager.VerifyPassword(username, password)
	if err != nil {
		return false, "", fmt.Errorf("authentication error: %w", err)
	}

	if !valid {
		return false, "", fmt.Errorf("invalid credentials")
	}

	return true, username, nil
}

// CheckNamespaceAccess checks if a user has access to a namespace
func CheckNamespaceAccess(username, namespace string) bool {
	user, err := userManager.GetUser(username)
	if err != nil {
		return false
	}

	if user == nil {
		return false
	}

	if types.ReservedNamespaces[namespace] && !user.IsAdmin {
		return false
	}

	return slices.Contains(user.Namespaces, namespace)
}

// CheckAdminAccess checks if a user is an admin
func CheckAdminAccess(username string) bool {
	user, err := userManager.GetUser(username)
	if err != nil {
		return false
	}

	if user == nil {
		return false
	}

	return user.IsAdmin
}

// ListUsersHandler handles requests to list users
func ListUsersHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	users, err := userManager.ListUsers()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to list users: %v", err))
		return
	}
	respondWithSuccess(c, users)
}

// GetUserHandler handles requests to get a user
func GetUserHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	user, err := userManager.GetUser(name)
	if err != nil {
		log.Printf("failed to get user: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get user: %v", err))
		return
	}
	if user == nil {
		respondWithError(c, http.StatusNotFound, "user not found")
		return
	}
	respondWithSuccess(c, user)
}

// CreateUserHandler handles requests to create a new user
func CreateUserHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	var user types.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("failed to bind user: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("failed to bind user: %v", err))
		return
	}
	err = userManager.CreateUser(name, user)
	if err != nil {
		log.Printf("failed to create user: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %v", err))
		return
	}
	respondWithSuccess(c, user)
}

// DeleteUserHandler handles requests to delete a user
func DeleteUserHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	err = userManager.DeleteUser(name)
	if err != nil {
		log.Printf("failed to delete user: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete user: %v", err))
		return
	}
	respondWithSuccess(c, nil)
}

// GetUserPasswordHandler handles requests to get a user's password
func GetUserPasswordHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	password, err := userManager.GetUserPassword(name)
	if err != nil {
		log.Printf("failed to get user password: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to get user password: %v", err))
		return
	}
	respondWithSuccess(c, password)
}

// SetUserPasswordHandler handles requests to set a user's password
func SetUserPasswordHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	if name == "" {
		log.Printf("name is required")
		respondWithError(c, http.StatusBadRequest, "name is required")
		return
	}
	var password string
	if err := c.ShouldBindJSON(&password); err != nil {
		log.Printf("failed to bind password: %v", err)
		respondWithError(c, http.StatusBadRequest, fmt.Sprintf("failed to bind password: %v", err))
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to hash password: %v", err))
		return
	}

	err = userManager.SetUserPassword(name, string(hashedPassword))
	if err != nil {
		log.Printf("failed to set user password: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to set user password: %v", err))
		return
	}
	respondWithSuccess(c, nil)
}

// AddNamespaceToUserHandler handles requests to add a namespace to a user
func AddNamespaceToUserHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		log.Printf("name and namespace are required")
		respondWithError(c, http.StatusBadRequest, "name and namespace are required")
		return
	}
	err = userManager.AddNamespaceToUser(name, namespace)
	if err != nil {
		log.Printf("failed to add namespace to user: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to add namespace to user: %v", err))
		return
	}
	respondWithSuccess(c, nil)
}

// RemoveNamespaceFromUserHandler handles requests to remove a namespace from a user
func RemoveNamespaceFromUserHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		log.Printf("name and namespace are required")
		respondWithError(c, http.StatusBadRequest, "name and namespace are required")
		return
	}
	err = userManager.RemoveNamespaceFromUser(name, namespace)
	if err != nil {
		log.Printf("failed to remove namespace from user: %v", err)
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to remove namespace from user: %v", err))
		return
	}
	respondWithSuccess(c, nil)
}
