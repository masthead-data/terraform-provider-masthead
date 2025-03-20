package main

import (
	"fmt"
	"log"
	"os"

	"github.com/masthead-data/masthead-client-go"
)

// main function to demonstrate Masthead API client usage
func main() {
	// Retrieve API token from the MASTHEAD_API_TOKEN environment variable
	apiToken := os.Getenv("MASTHEAD_API_TOKEN")

	if apiToken == "" {
		// Log a fatal error and exit if the API token is not set
		log.Fatal("MASTHEAD_API_TOKEN environment variable not set")
	}

	// Instantiate a new Masthead API client using the retrieved token
	client, err := masthead.NewClient(nil, &apiToken)
	if err != nil {
		// Log a fatal error and exit if client creation fails
		log.Fatalf("Error creating client: %v", err)
	}

	// Sample data for creating a user
	userEmail := "testuser@example.com"
	userRole := "USER"

	// Call CreateUser with sample data
	err = client.CreateUser(userEmail, userRole)
	if err != nil {
		// Log an error if user creation fails
		log.Printf("Error creating user: %v", err)
	} else {
		// Print success message if user creation is successful
		fmt.Printf("User %s created successfully with role %s\n", userEmail, userRole)
	}

	// Call ListUsers to retrieve a list of users
	users, err := client.ListUsers()
	if err != nil {
		// Log an error if listing users fails
		log.Printf("Error listing users: %v", err)
	} else {
		// Print each user's email and role
		fmt.Println("List of users:")
		for _, user := range users {
			fmt.Printf("- Email: %s, Role: %s\n", user.Email, user.Role)
		}
	}

	// Sample data for updating a user's role
	newUserRole := "OWNER"

	// Call UpdateUserRole for a user
	err = client.UpdateUserRole(userEmail, newUserRole)
	if err != nil {
		// Log an error if updating the user role fails
		log.Printf("Error updating user role: %v", err)
	} else {
		// Print success message if the user role update is successful
		fmt.Printf("User %s role updated to %s\n", userEmail, newUserRole)
	}

	// Call DeleteUser for a user
	err = client.DeleteUser(userEmail)
	if err != nil {
		// Log an error if deleting the user fails
		log.Printf("Error deleting user: %v", err)
	} else {
		// Print success message if the user deletion is successful
		fmt.Printf("User %s deleted successfully\n", userEmail)
	}
}
