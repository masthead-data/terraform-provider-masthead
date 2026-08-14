package masthead

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccClient(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping integration test; set TF_ACC=1 to run")
	}

	// Retrieve API token from the MASTHEAD_API_TOKEN environment variable
	apiToken := os.Getenv("MASTHEAD_API_TOKEN")
	assert.NotEmpty(t, apiToken, "API token should not be empty")

	// Instantiate a new Masthead API client using the retrieved token
	apiClient, err := NewClient(&apiToken, 0)
	assert.NoError(t, err, "Client creation should not return an error")

	t.Log("Masthead API client created successfully")

	// Call the example function to demonstrate API operations
	apiClientExample(apiClient, t)
	fmt.Println("Completed successfully")
}

func TestFormatAPIError(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   []string
	}{
		{
			name:   "structured errors array",
			status: 400,
			body:   `{"message":"2 data assets cannot be resolved","errors":[{"type":"TABLE","project":"p","dataset":"d","table":"t1","reason":"DELETED","deletedAt":"2026-07-04T01:27:06Z"},{"type":"LOOKER_DASHBOARD","uuid":"abc-123","reason":"NOT_FOUND"}]}`,
			want:   []string{"2 data assets cannot be resolved", "TABLE p.d.t1 — DELETED (2026-07-04T01:27:06Z)", "LOOKER_DASHBOARD abc-123 — NOT_FOUND"},
		},
		{
			name:   "plain error body falls back to raw dump",
			status: 400,
			body:   `{"message":"Data asset not found."}`,
			want:   []string{"status: 400", "Data asset not found."},
		},
		{
			name:   "non-JSON body falls back to raw dump",
			status: 502,
			body:   `<html>bad gateway</html>`,
			want:   []string{"status: 502", "<html>bad gateway</html>"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatAPIError(tc.status, []byte(tc.body))
			for _, w := range tc.want {
				if !strings.Contains(got.Error(), w) {
					t.Errorf("error %q missing substring %q", got.Error(), w)
				}
			}
		})
	}
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
	testDomain := DataDomain{
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
			fmt.Printf("- ID: %s, Name: %s, Email: %s, Slack Channel: %s\n", domain.UUID, domain.Name, domain.Email, domain.SlackChannel)

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
		testDomain.SlackChannelName = ""
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
			Type:    DataProductAssetTypeDataset,
			Project: "httparchive",
			Dataset: "scratchspace",
		},
		{
			Type:    DataProductAssetTypeTable,
			Project: "httparchive",
			Dataset: "crawl",
			Table:   "pages",
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
			Type:    DataProductAssetTypeTable,
			Project: "httparchive",
			Dataset: "sample_data",
			Table:   "pages_10k",
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
