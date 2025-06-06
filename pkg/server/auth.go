package server

import (
	"fmt"
	"log"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// CheckAuth verifies user authentication using HTTP Basic Auth
func CheckAuth(c *gin.Context) (bool, string, error) {
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		log.Printf("Authentication failed: missing basic auth credentials")
		return false, "", fmt.Errorf("missing basic auth credentials")
	}

	log.Printf("Attempting authentication for user: %s", username)

	// Verify the password against stored password (plain text)
	valid, err := userManager.VerifyPassword(username, password)
	if err != nil {
		log.Printf("Authentication error for user %s: %v", username, err)
		return false, "", fmt.Errorf("authentication error: %w", err)
	}

	if !valid {
		log.Printf("Authentication failed for user %s:%s invalid credentials", username, password)
		return false, "", fmt.Errorf("invalid credentials")
	}

	log.Printf("Authentication successful for user: %s", username)
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

	if user.IsAdmin {
		return true
	}

	if types.ReservedNamespaces[namespace] {
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

// CreateRootUser creates a root user
func CreateRootUser(password string) error {
	log.Printf("Creating root user...")

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
