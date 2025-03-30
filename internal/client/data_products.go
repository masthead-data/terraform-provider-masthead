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
	DataProductAssetTypeDataset DataProductAssetType = "DATASET"
	DataProductAssetTypeTable DataProductAssetType = "TABLE"
)

type DataProductAsset struct {
	Type DataProductAssetType `json:"type"`
	UUID string               `json:"uuid"`
}

type Subscribers struct {
	Values []interface{} `json:"values"` // TODO: replace with a concrete type
	Total  int           `json:"total"`
}

type DataProduct struct {
	UUID             string             `json:"uuid"`
	Name           string             `json:"name"`
	DataDomainUUID string             `json:"_dataDomainUuid"`
	Description    string             `json:"description"`
	Domain         *Domain            `json:"domain"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	DataAssets     []DataProductAsset `json:"dataAssets"`
}

// DataProductResponse represents the response from the create/update data product API
type DataProductResponse struct {
	Value DataProduct `json:"value"`
	Extra  interface{} `json:"extra"`
	Error  interface{} `json:"error"`
}

// DataProductsResponse represents the response from the list data products API
type DataProductsResponse struct {
	Values []DataProduct `json:"values"`
	Extra  interface{}   `json:"extra"`
	Error  interface{}   `json:"error"`
	Pagination struct {
		Total int `json:"total"`
		Page  int `json:"page"`
	} `json:"pagination"`
}

// ListDataProducts - Returns list of all data products with pagination
func (c *Client) ListDataProducts() ([]DataProduct, error) {
	var allProducts []DataProduct
	page := 1
	morePages := true

	for morePages {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-product/list?page=%d&limit=100",
			c.HostURL, page), nil)
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

		allProducts = append(allProducts, productsResponse.Values...)

		// Check if we've fetched all pages
		totalItems := productsResponse.Pagination.Total
		itemsFetched := len(allProducts)
		if itemsFetched >= totalItems {
			morePages = false
		} else {
			page++
		}
	}

	return allProducts, nil
}

// CreateDataProduct - Create a new data product in the system
func (c *Client) CreateDataProduct(dataProduct DataProduct) (*DataProduct, error) {

	rb, err := json.Marshal(dataProduct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-product", c.HostURL),
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	createdProduct := &DataProduct{}
	err = json.Unmarshal(body, createdProduct)
	if err != nil {
		return nil, err
	}

	return createdProduct, nil
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
func (c *Client) UpdateDataProduct(dataProduct DataProduct) (*DataProduct, error) {
	rb, err := json.Marshal(dataProduct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/clientApi/data-product/%s", c.HostURL, dataProduct.UUID),
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	updatedProduct := &DataProduct{}
	err = json.Unmarshal(body, updatedProduct)
	if err != nil {
		return nil, err
	}

	return updatedProduct, nil
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
