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

	// Example section: Data Product operations
	fmt.Println("\n=== Data Product Operations ===")
	dataProductExample(client)
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

	// Sample data for creating a data domain
	domainName := "Marketing Data"
	domainEmail := "marketing@example.com"
	slackChannel := "#marketing-data"

	// Call CreateDataDomain with sample data
	err := client.CreateDataDomain(domainName, domainEmail, &slackChannel)
	if err != nil {
		log.Printf("Error creating data domain: %v", err)
	} else {
		fmt.Printf("Data domain '%s' created successfully\n", domainName)
	}

	// Call ListDataDomains to retrieve a list of data domains
	domains, err := client.ListDataDomains()
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
		domain, err := client.GetDataDomain(domainID)
		if err != nil {
			log.Printf("Error getting data domain: %v", err)
		} else {
			fmt.Printf("Retrieved data domain: %s (ID: %s)\n", domain.Name, domain.ID)
		}

		// Update the data domain
		updatedName := domainName + " (Updated)"
		updatedSlackChannel := slackChannel + "-updates"
		err = client.UpdateDataDomain(domainID, updatedName, domainEmail, &updatedSlackChannel)
		if err != nil {
			log.Printf("Error updating data domain: %v", err)
		} else {
			fmt.Printf("Data domain updated to '%s'\n", updatedName)
		}

		// Delete the data domain
		err = client.DeleteDataDomain(domainID)
		if err != nil {
			log.Printf("Error deleting data domain: %v", err)
		} else {
			fmt.Printf("Data domain '%s' (ID: %s) deleted successfully\n", updatedName, domainID)
		}
	} else {
		fmt.Println("No domains found to perform additional operations")
	}
}

// dataProductExample demonstrates the Data Product API operations
func dataProductExample(client *masthead.Client) {
	// Sample data for creating a data product
	productName := "Customer Analytics"

	// Sample data assets
	dataAssets := []masthead.DataProductAsset{
		{
			Type: masthead.DataProductAssetTypeDataset,
			UUID: "1583db12-9ed3-3458-ad99-8c25413f6a5b",
		},
		{
			Type: masthead.DataProductAssetTypeTable,
			UUID: "5656f586-d9d5-3f7a-b9f2-06a44f72e5f2",
		},
	}

	// Optional fields
	dataDomainUUID := "a23422d9-5c7a-423c-afa7-5c5de18ff9df"
	description := "Customer analytics data product for marketing team"

	// Call CreateDataProduct with sample data
	err := client.CreateDataProduct(productName, dataAssets, &dataDomainUUID, &description)
	if err != nil {
		log.Printf("Error creating data product: %v", err)
	} else {
		fmt.Printf("Data product '%s' created successfully\n", productName)
	}

	// Call ListDataProducts to retrieve a list of data products
	products, err := client.ListDataProducts()
	if err != nil {
		log.Printf("Error listing data products: %v", err)
	} else {
		fmt.Println("List of data products:")

		// Variable to store product ID for later operations
		var productID string

		for _, product := range products {
			fmt.Printf("- ID: %s, Name: %s\n", product.ID, product.Name)
			if product.Description != "" {
				fmt.Printf("  Description: %s\n", product.Description)
			}
			fmt.Printf("  Data Assets: %d\n", len(product.DataAssets))

			// Store the first product ID for later use in examples
			if productID == "" {
				productID = product.ID
			}
		}

		// If we obtained a product ID from the list, use it for further operations
		if productID != "" {
			// Get a specific data product
			product, err := client.GetDataProduct(productID)
			if err != nil {
				log.Printf("Error getting data product: %v", err)
			} else {
				fmt.Printf("\nRetrieved data product: %s (ID: %s)\n", product.Name, product.ID)
				fmt.Printf("Data Assets: %d\n", len(product.DataAssets))
				for i, asset := range product.DataAssets {
					fmt.Printf("  Asset %d: Type=%s, UUID=%s\n", i+1, asset.Type, asset.UUID)
				}
			}

			// Update the data product
			updatedName := productName + " (Updated)"
			updatedDescription := description + " - with additional metrics"

			// Add an additional data asset for the update
			updatedAssets := append(dataAssets, masthead.DataProductAsset{
				Type: masthead.DataProductAssetTypeTable,
				UUID: "7777f586-d9d5-3f7a-b9f2-06a44f72e9a9",
			})

			err = client.UpdateDataProduct(productID, updatedName, updatedAssets, &dataDomainUUID, &updatedDescription)
			if err != nil {
				log.Printf("Error updating data product: %v", err)
			} else {
				fmt.Printf("\nData product updated to '%s' with %d assets\n", updatedName, len(updatedAssets))
			}

			// Delete the data product
			err = client.DeleteDataProduct(productID)
			if err != nil {
				log.Printf("Error deleting data product: %v", err)
			} else {
				fmt.Printf("\nData product '%s' (ID: %s) deleted successfully\n", updatedName, productID)
			}
		} else {
			fmt.Println("No data products found to perform additional operations")
		}
	}
}
