package main

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/client"
)

func main() {
	// Create a new client with JWT authentication
	// The client will automatically generate JWT tokens for each request
	c := client.NewClient("localhost", "6969", "admin", "password")

	// Example: List VMs using JWT authentication
	fmt.Println("Listing VMs with JWT authentication...")
	vms, err := c.ListVMs("default")
	if err != nil {
		log.Printf("Error listing VMs: %v", err)
		return
	}

	fmt.Printf("Found %d VMs:\n", len(vms))
	for _, vm := range vms {
		fmt.Printf("- %s\n", vm)
	}

	// Example: List users using JWT authentication
	fmt.Println("\nListing users with JWT authentication...")
	users, err := c.ListUsers()
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("- %s\n", user.Name)
	}

	// Example: List nodes using JWT authentication
	fmt.Println("\nListing nodes with JWT authentication...")
	nodes, err := c.ListNodes()
	if err != nil {
		log.Printf("Error listing nodes: %v", err)
		return
	}

	fmt.Printf("Found %d nodes:\n", len(nodes))
	for _, node := range nodes {
		fmt.Printf("- %s\n", node)
	}

	fmt.Println("\nJWT authentication is working correctly!")
}
