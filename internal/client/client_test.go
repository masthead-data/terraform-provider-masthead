package masthead

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClient is a placeholder test function to make the test file valid
func TestClient(t *testing.T) {
	// Retrieve API token from the MASTHEAD_API_TOKEN environment variable
	apiToken := os.Getenv("MASTHEAD_API_TOKEN")
	assert.NotEmpty(t, apiToken, "API token should not be empty")

	// Instantiate a new Masthead API client using the retrieved token
	apiClient, err := NewClient(&apiToken)
	assert.NoError(t, err, "Client creation should not return an error")


	t.Log("Masthead API client created successfully")

	// Call the example function to demonstrate API operations
	apiClientExample(apiClient, t)
	fmt.Println("Example usage completed successfully")
}

// userExample demonstrates the User API operations
func apiClientExample(client *Client, t *testing.T) {
	testUser := User{
		Email: "testuser@example.com",
		Role:  "USER",
	}

	user, err := client.CreateUser(testUser)
	assert.NoError(t, err, "User creation should not return an error")
	if err == nil {
		t.Logf("User %s created successfully with role %s\n", user.Email, user.Role)
	}

	// Call ListUsers to retrieve a list of users
	users, err := client.ListUsers()
	assert.NoError(t, err, "User listing should not return an error")
	if err == nil {
		t.Logf("List of users:")
		for _, user := range users {
			fmt.Printf("- Email: %s, Role: %s\n", user.Email, user.Role)
		}
	}

	// Sample data for updating a user's role
	testUser.Role = "OWNER"

	// Call UpdateUserRole for a user
	user, err = client.UpdateUserRole(testUser)
	assert.NoError(t, err, "User role update should not return an error")
	if err == nil {
		t.Logf("User %s role updated to %s\n", user.Email, user.Role)
	}

	// Call DeleteUser for a user
	err = client.DeleteUser(user.Email)
	assert.NoError(t, err, "User deletion should not return an error")
	if err == nil {
		t.Logf("User %s deleted successfully\n", user.Email)
	}

	// domainExample demonstrates the Data Domain API operations

	// Sample data for creating a data domain
	testDomain := Domain{
		Name:             "API Test Domain",
		Email:            "domain@example.com",
		SlackChannelName: "10x-infra",
	}

	// Call CreateDomain with sample data
	domain, err := client.CreateDomain(testDomain)
	assert.NoError(t, err, "Data domain creation should not return an error")
	if err == nil {
		fmt.Printf("Data domain '%s' created successfully\n", domain.Name)
		testDomain.UUID = domain.UUID
	}

	// Call ListDomains to retrieve a list of data domains
	domains, err := client.ListDomains()
	assert.NoError(t, err, "Data domain listing should not return an error")
	if err == nil {
		t.Logf("List of data domains:")
		for _, domain := range domains {
			fmt.Printf("- ID: %s, Name: %s, Email: %s\n", domain.UUID, domain.Name, domain.Email)
			if domain.SlackChannelName != "" {
				fmt.Printf("  Slack Channel: %s\n", domain.SlackChannelName)
			}

		}
	}

	assert.NotEmpty(t, testDomain.UUID, "Test data domain UUID should not be empty")
	if testDomain.UUID != "" {
		// Get a specific domain
		domain, err = client.GetDomain(testDomain.UUID)
		assert.NoError(t, err, "Data domain retrieval should not return an error")
		if err == nil {
			t.Logf("Retrieved data domain: %s (ID: %s)\n", domain.Name, domain.UUID)
		}

		// Update the data domain
		testDomain.Name = testDomain.Name + " (Updated)"
		domain, err = client.UpdateDomain(testDomain)
		assert.NoError(t, err, "Data domain update should not return an error")
		if err == nil {
			t.Logf("Data domain updated to '%s'\n", domain.Name)
		}
	}

	// dataProductExample demonstrates the Data Product API operations
	// Sample data assets
	dataAssets := []DataProductAsset{
		{
			Type: DataProductAssetTypeDataset,
			UUID: "1583db12-9ed3-3458-ad99-8c25413f6a5b",
		},
		{
			Type: DataProductAssetTypeTable,
			UUID: "5656f586-d9d5-3f7a-b9f2-06a44f72e5f2",
		},
	}

	// Sample data for creating a data product
	testProduct := DataProduct{
		Name:           "Test Product",
		Description:    "Data Product for API testing",
		DataDomainUUID: testDomain.UUID,
		DataAssets:     dataAssets,
	}

	// Call CreateDataProduct with sample data
	dataProduct, err := client.CreateDataProduct(testProduct)
	assert.NoError(t, err, "Data product creation should not return an error")
	if err == nil {
		t.Logf("Data product '%s' created successfully\n", dataProduct.Name)

		// Store the product ID for later use
		testProduct.UUID = dataProduct.UUID

	}

	// Call ListDataProducts to retrieve a list of data products
	dataProducts, err := client.ListDataProducts()
	assert.NoError(t, err, "Data product listing should not return an error")
	if err == nil {
		t.Logf("List of data products:")
		for _, product := range dataProducts {
			fmt.Printf("- ID: %s, Name: %s\n", product.UUID, product.Name)
			if product.Description != "" {
				fmt.Printf("  Description: %s\n", product.Description)
			}
		}
	}

	// If we obtained an ID after creating a product, use it for further operations
	assert.NotEmpty(t, testProduct.UUID, "Test data product UUID should not be empty")
	if testProduct.UUID != "" {
		// Get a specific data product
		dataProduct, err := client.GetDataProduct(testProduct.UUID)
		assert.NoError(t, err, "Data product retrieval should not return an error")
		if err == nil {
			t.Logf("\nRetrieved data product: %s (ID: %s)\n", dataProduct.Name, dataProduct.UUID)
			fmt.Printf("Data Assets: %d\n", len(dataProduct.DataAssets))
			for i, asset := range dataProduct.DataAssets {
				fmt.Printf("  Asset %d: Type=%s, UUID=%s\n", i+1, asset.Type, asset.UUID)
			}
		}

		// Update the data product
		testProduct.Name = testProduct.Name + " (Updated)"
		testProduct.Description = testProduct.Description + " - with updated description"

		// Add an additional data asset for the update
		testProduct.DataAssets = append(testProduct.DataAssets, DataProductAsset{
			Type: DataProductAssetTypeTable,
			UUID: "7777f586-d9d5-3f7a-b9f2-06a44f72e9a9",
		})

		dataProduct, err = client.UpdateDataProduct(testProduct)
		assert.NoError(t, err, "Data product update should not return an error")
		if err == nil {
			t.Logf("\nData product updated to '%s' with %d assets\n", dataProduct.Name, len(dataProduct.DataAssets))
		}

		// Delete the data product
		err = client.DeleteDataProduct(testProduct.UUID)
		assert.NoError(t, err, "Data product deletion should not return an error")
		if err == nil {
			t.Logf("\nData product '%s' (ID: %s) deleted successfully\n", testProduct.Name, testProduct.UUID)
		}

		// Delete the data domain
		err = client.DeleteDomain(testDomain.UUID)
		assert.NoError(t, err, "Data domain deletion should not return an error")
		if err == nil {
			t.Logf("Data domain '%s' (ID: %s) deleted successfully\n", testDomain.Name, testDomain.UUID)
		}
	}
}
