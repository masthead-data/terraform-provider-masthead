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
	apiClient, err := masthead.NewClient(&apiToken)
	if err != nil {
		// Log a fatal error and exit if client creation fails
		log.Fatalf("Error creating client: %v", err)
	}

	// Example section: Data Product operations
	fmt.Println("\n=== API Client Operations ===")
	apiClientExample(apiClient)
}

// userExample demonstrates the User API operations
func apiClientExample(client *masthead.Client) {
	testUser := masthead.User{
		Email: "testuser@example.com",
		Role:  "USER",
	}

	// Call CreateUser with sample data
	user, err := client.CreateUser(testUser)
	if err != nil {
		// Log an error if user creation fails
		log.Printf("Error creating user: %v", err)
	} else {
		// Print success message if user creation is successful
		fmt.Printf("User %s created successfully with role %s\n", user.Email, user.Role)
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
	testUser.Role = "OWNER"

	// Call UpdateUserRole for a user
	user, err = client.UpdateUserRole(testUser)
	if err != nil {
		// Log an error if updating the user role fails
		log.Printf("Error updating user role: %v", err)
	} else {
		// Print success message if the user role update is successful
		fmt.Printf("User %s role updated to %s\n", user.Email, user.Role)
	}

	// Call DeleteUser for a user
	err = client.DeleteUser(user.Email)
	if err != nil {
		// Log an error if deleting the user fails
		log.Printf("Error deleting user: %v", err)
	} else {
		// Print success message if the user deletion is successful
		fmt.Printf("User %s deleted successfully\n", user.Email)
	}

	// domainExample demonstrates the Data Domain API operations

	// Variable to store domain ID for later operations
	var domainUUID string

	// Sample data for creating a data domain
	testDomain := masthead.Domain{
		Name:             "API Test Domain",
		Email:            "domain@example.com",
		SlackChannelName: "10x-infra",
	}

	// Call CreateDomain with sample data
	domain, err := client.CreateDomain(testDomain)
	if err != nil {
		log.Printf("Error creating data domain: %v", err)
	} else {
		fmt.Printf("Data domain '%s' created successfully\n", domain.Name)

		// Store the first domain ID for later use in examples
		if domainUUID == "" {
			domainUUID = domain.UUID
		}
	}

	// Call ListDomains to retrieve a list of data domains
	domains, err := client.ListDomains()
	if err != nil {
		log.Printf("Error listing data domains: %v", err)
	} else {
		fmt.Println("List of data domains:")
		for _, domain := range domains {
			fmt.Printf("- ID: %s, Name: %s, Email: %s\n", domain.UUID, domain.Name, domain.Email)
			if domain.SlackChannel.Name != "" {
				fmt.Printf("  Slack Channel: %s\n", domain.SlackChannel)
			}

		}
	}

	// If we obtained an ID after creating a domain, use it for further operations
	if domainUUID != "" {
		// Get a specific domain
		domain, err = client.GetDomain(domainUUID)
		if err != nil {
			log.Printf("Error getting data domain: %v", err)
		} else {
			fmt.Printf("Retrieved data domain: %s (ID: %s)\n", domain.Name, domain.UUID)
		}

		// Update the data domain
		testDomain.Name = testDomain.Name + " (Updated)"
		domain, err = client.UpdateDomain(testDomain)
		if err != nil {
			log.Printf("Error updating data domain: %v", err)
		} else {
			fmt.Printf("Data domain updated to '%s'\n", domain.Name)
		}

		// Delete the data domain
		err = client.DeleteDomain(domainUUID)
		if err != nil {
			log.Printf("Error deleting data domain: %v", err)
		} else {
			fmt.Printf("Data domain '%s' (ID: %s) deleted successfully\n", testDomain.Name, domainUUID)
		}
	} else {
		fmt.Println("No data domain available to perform additional operations")
	}

	// dataProductExample demonstrates the Data Product API operations
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

	// Sample data for creating a data product
	testProduct := masthead.DataProduct{
		Name:           "Test Product",
		Description:    "Data Product for API testing",
		DataDomainUUID: domainUUID,
		DataAssets:     dataAssets,
	}

	// Call CreateDataProduct with sample data
	dataProduct, err := client.CreateDataProduct(testProduct)
	if err != nil {
		log.Printf("Error creating data product: %v", err)
	} else {
		fmt.Printf("Data product '%s' created successfully\n", dataProduct.Name)

		// Store the product ID for later use
		testProduct.UUID = dataProduct.UUID

	}

	// Call ListDataProducts to retrieve a list of data products
	dataProducts, err := client.ListDataProducts()
	if err != nil {
		log.Printf("Error listing data products: %v", err)
	} else {
		fmt.Println("List of data products:")
		for _, product := range dataProducts {
			fmt.Printf("- ID: %s, Name: %s\n", product.UUID, product.Name)
			if product.Description != "" {
				fmt.Printf("  Description: %s\n", product.Description)
			}
		}
	}

	// If we obtained an ID after creating a product, use it for further operations
	if testProduct.UUID != "" {
		// Get a specific data product
		dataProduct, err := client.GetDataProduct(testProduct.UUID)
		if err != nil {
			log.Printf("Error getting data product: %v", err)
		} else {
			fmt.Printf("\nRetrieved data product: %s (ID: %s)\n", dataProduct.Name, dataProduct.UUID)
			fmt.Printf("Data Assets: %d\n", len(dataProduct.DataAssets))
			for i, asset := range dataProduct.DataAssets {
				fmt.Printf("  Asset %d: Type=%s, UUID=%s\n", i+1, asset.Type, asset.UUID)
			}
		}

		// Update the data product
		testProduct.Name = testProduct.Name + " (Updated)"
		testProduct.Description = testProduct.Description + " - with updated description"

		// Add an additional data asset for the update
		testProduct.DataAssets = append(testProduct.DataAssets, masthead.DataProductAsset{
			Type: masthead.DataProductAssetTypeTable,
			UUID: "7777f586-d9d5-3f7a-b9f2-06a44f72e9a9",
		})

		dataProduct, err = client.UpdateDataProduct(testProduct)
		if err != nil {
			log.Printf("Error updating data product: %v", err)
		} else {
			fmt.Printf("\nData product updated to '%s' with %d assets\n", dataProduct.Name, len(dataProduct.DataAssets))
		}

		// Delete the data product
		err = client.DeleteDataProduct(testProduct.UUID)
		if err != nil {
			log.Printf("Error deleting data product: %v", err)
		} else {
			fmt.Printf("\nData product '%s' (ID: %s) deleted successfully\n", testProduct.Name, testProduct.UUID)
		}
	} else {
		fmt.Println("No data product available to perform additional operations")
	}
}
