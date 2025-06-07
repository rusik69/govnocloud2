package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// UserManager handles user operations for k3s
type UserManager struct {
	etcdClient *clientv3.Client
}

// NewUserManager creates a new UserManager for k3s
func NewUserManager() (*UserManager, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}
	return &UserManager{etcdClient: etcdClient}, nil
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

	// Store password separately if provided
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

	// Store password separately if provided (plain text)
	if password != "" {
		err = m.SetUserPassword(name, password)
		if err != nil {
			return fmt.Errorf("failed to set user password: %w", err)
		}
	}

	return nil
}

// SetUserPassword sets a user's password (plain text)
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

	// Store password in plain text
	key := fmt.Sprintf("/users/%s/password", name)
	_, err = m.etcdClient.Put(ctx, key, password)
	if err != nil {
		return fmt.Errorf("failed to store user password in etcd: %w", err)
	}

	return nil
}

// GetUserPassword gets a user's password (plain text)
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

// CreateRootUser creates a root user
func CreateRootUser(password string) error {
	log.Printf("Creating root user...")

	// Create user manager instance
	userManager, err := NewUserManager()
	if err != nil {
		return fmt.Errorf("failed to create user manager: %w", err)
	}
	defer userManager.etcdClient.Close()

	// Check if root user already exists
	existingUser, err := userManager.GetUser("root")
	if err != nil {
		log.Printf("Error checking for existing root user: %v", err)
	}

	if existingUser != nil {
		log.Printf("Root user already exists, skipping creation")
		// Verify the existing user's password
		storedPassword, err := userManager.GetUserPassword("root")
		if err != nil {
			log.Printf("Error getting existing root user password: %v", err)
		} else {
			log.Printf("Existing root user password: '%s'", storedPassword)
		}
		return nil
	}

	log.Printf("Creating root user with password '%s'", password)
	user := types.User{
		Name:     "root",
		IsAdmin:  true,
		Password: password, // Default password
	}
	err = userManager.CreateUser("root", user)
	if err != nil {
		log.Printf("Failed to create root user: %v", err)
		return fmt.Errorf("failed to create root user: %w", err)
	}

	log.Printf("Root user created successfully with password '%s'", password)

	// Verify the user was created correctly
	createdUser, err := userManager.GetUser("root")
	if err != nil {
		log.Printf("Error verifying created root user: %v", err)
	} else if createdUser != nil {
		log.Printf("Verified root user exists: IsAdmin=%v", createdUser.IsAdmin)
		storedPassword, err := userManager.GetUserPassword("root")
		if err != nil {
			log.Printf("Error getting created root user password: %v", err)
		} else {
			log.Printf("Verified root user password: '%s'", storedPassword)
		}
	}

	return nil
}
