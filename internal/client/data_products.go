package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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

		productsResponse := DataProductListResponse{}
		err = json.Unmarshal(body, &productsResponse)
		if err != nil {
			return nil, err
		} else if productsResponse.Error != nil {
			return nil, fmt.Errorf("error: %v. %v", productsResponse.Error, productsResponse.Message)
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

	productResponse := &DataProductResponse{}
	err = json.Unmarshal(body, productResponse)
	if err != nil {
		return nil, err
	} else if productResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", productResponse.Error, productResponse.Message)
	}

	return &productResponse.Value, nil
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

	productResponse := &DataProductResponse{}
	err = json.Unmarshal(body, productResponse)
	if productResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", productResponse.Error, productResponse.Message)
	} else if err != nil {
		return nil, err
	}

	return &productResponse.Value, nil
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

	productResponse := &DataProductResponse{}
	err = json.Unmarshal(body, productResponse)
	if productResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", productResponse.Error, productResponse.Message)
	} else if err != nil {
		return nil, err
	}

	return &productResponse.Value, nil
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
