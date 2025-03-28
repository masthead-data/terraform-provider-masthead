package main

import (
	"fmt"
	"log"
	"os"

	"github.com/masthead-data/terraform-provider-masthead/internal/client"
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
	client, err := masthead.NewClient(&apiToken)
	if err != nil {
		// Log a fatal error and exit if client creation fails
		log.Fatalf("Error creating client: %v", err)
	}

	// Example section: User operations
	fmt.Println("=== User Operations ===")
	userExample(client)

	// Example section: Domain operations 
	fmt.Println("\n=== Domain Operations ===")
	domainExample(client)
}

// userExample demonstrates the User API operations
func userExample(client *masthead.Client) {
	// Sample data for creating a user
	userEmail := "testuser@example.com"
	userRole := "USER"

	// Call CreateUser with sample data
	err := client.CreateUser(userEmail, userRole)
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

// domainExample demonstrates the Data Domain API operations
func domainExample(client *masthead.Client) {
	// Sample tenant and project IDs 
	tenantID := "example-tenant"
	projectID := "example-project"
	
	// Sample data for creating a data domain
	domainName := "Marketing Data"
	domainEmail := "marketing@example.com"
	slackChannel := "#marketing-data"

	// Call CreateDataDomain with sample data
	err := client.CreateDataDomain(tenantID, projectID, domainName, domainEmail, &slackChannel)
	if err != nil {
		log.Printf("Error creating data domain: %v", err)
	} else {
		fmt.Printf("Data domain '%s' created successfully\n", domainName)
	}

	// Call ListDataDomains to retrieve a list of data domains
	domains, err := client.ListDataDomains(tenantID, projectID)
	if err != nil {
		log.Printf("Error listing data domains: %v", err)
	} else {
		fmt.Println("List of data domains:")
		for _, domain := range domains {
			fmt.Printf("- ID: %s, Name: %s, Email: %s\n", domain.ID, domain.Name, domain.Email)
			if domain.SlackChannel != "" {
				fmt.Printf("  Slack Channel: %s\n", domain.SlackChannel)
			}
			
			// Store the first domain ID for later use in examples
			if len(domainID) == 0 {
				domainID = domain.ID
			}
		}
	}

	// If we obtained a domain ID from the list, use it for further operations
	var domainID string
	if len(domains) > 0 {
		domainID = domains[0].ID
		
		// Get a specific domain
		domain, err := client.GetDataDomain(tenantID, projectID, domainID)
		if err != nil {
			log.Printf("Error getting data domain: %v", err)
		} else {
			fmt.Printf("Retrieved data domain: %s (ID: %s)\n", domain.Name, domain.ID)
		}
		
		// Update the data domain
		updatedName := domainName + " (Updated)"
		updatedSlackChannel := slackChannel + "-updates"
		err = client.UpdateDataDomain(tenantID, projectID, domainID, updatedName, domainEmail, &updatedSlackChannel)
		if err != nil {
			log.Printf("Error updating data domain: %v", err)
		} else {
			fmt.Printf("Data domain updated to '%s'\n", updatedName)
		}
		
		// Delete the data domain
		err = client.DeleteDataDomain(tenantID, projectID, domainID)
		if err != nil {
			log.Printf("Error deleting data domain: %v", err)
		} else {
			fmt.Printf("Data domain '%s' (ID: %s) deleted successfully\n", updatedName, domainID)
		}
	} else {
		fmt.Println("No domains found to perform additional operations")
	}
}
