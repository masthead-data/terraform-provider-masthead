// filepath: /Users/maxostapenko/Documents/GitHub/terraform-provider-masthead/internal/client/data_products.go
package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DataProductAssetType represents the type of a data asset
type DataProductAssetType string

const (
	// DataProductAssetTypeDataset represents a dataset asset type
	DataProductAssetTypeDataset DataProductAssetType = "DATASET"
	// DataProductAssetTypeTable represents a table asset type
	DataProductAssetTypeTable DataProductAssetType = "TABLE"
)

// DataProductAsset represents a data asset in a data product
type DataProductAsset struct {
	Type DataProductAssetType `json:"type"`
	UUID string               `json:"uuid"`
}

// DataProduct represents a data product in the system
type DataProduct struct {
	ID             string             `json:"id,omitempty"`
	Name           string             `json:"name"`
	DataAssets     []DataProductAsset `json:"dataAssets"`
	DataDomainUUID string             `json:"_dataDomainUuid,omitempty"`
	Description    string             `json:"description,omitempty"`
}

// DataProductsResponse represents the response from the list data products API
type DataProductsResponse struct {
	Values []DataProduct `json:"values"`
	Extra  interface{}   `json:"extra"`
	Error  interface{}   `json:"error"`
}

// ListDataProducts - Returns list of all data products
func (c *Client) ListDataProducts() ([]DataProduct, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-product",
		c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	productsResponse := DataProductsResponse{}
	err = json.Unmarshal(body, &productsResponse)
	if err != nil {
		return nil, err
	}

	return productsResponse.Values, nil
}

// CreateDataProduct - Create a new data product in the system
func (c *Client) CreateDataProduct(name string, dataAssets []DataProductAsset,
	dataDomainUUID *string, description *string) error {

	dataProduct := DataProduct{
		Name:       name,
		DataAssets: dataAssets,
	}

	if dataDomainUUID != nil {
		dataProduct.DataDomainUUID = *dataDomainUUID
	}

	if description != nil {
		dataProduct.Description = *description
	}

	rb, err := json.Marshal(dataProduct)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-product", c.HostURL),
		strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// GetDataProduct - Get a specific data product by ID
func (c *Client) GetDataProduct(productID string) (*DataProduct, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/clientApi/data-product/%s", c.HostURL, productID),
		nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	product := &DataProduct{}
	err = json.Unmarshal(body, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateDataProduct - Update an existing data product
func (c *Client) UpdateDataProduct(productID string, name string, dataAssets []DataProductAsset,
	dataDomainUUID *string, description *string) error {

	dataProduct := DataProduct{
		ID:         productID,
		Name:       name,
		DataAssets: dataAssets,
	}

	if dataDomainUUID != nil {
		dataProduct.DataDomainUUID = *dataDomainUUID
	}

	if description != nil {
		dataProduct.Description = *description
	}

	rb, err := json.Marshal(dataProduct)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/clientApi/data-product/%s", c.HostURL, productID),
		strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// DeleteDataProduct - Remove a data product from the system by ID
func (c *Client) DeleteDataProduct(productID string) error {
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/clientApi/data-product/%s", c.HostURL, productID),
		nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
