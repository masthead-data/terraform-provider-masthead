package masthead

import (
	"encoding/json"
	"fmt"
	"io"
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

		allProducts = append(allProducts, productsResponse.DataProducts...)

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

	if productResponse.DataProduct.UUID == "" {
		return nil, ErrEmptyValue
	}

	c.cacheMutex.Lock()
	if c.productsCache != nil {
		c.productsCache[productResponse.DataProduct.UUID] = &productResponse.DataProduct
	}
	c.cacheMutex.Unlock()

	return &productResponse.DataProduct, nil
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

	if productResponse.DataProduct.UUID == "" {
		return nil, ErrEmptyValue
	}

	return &productResponse.DataProduct, nil
}

// GetCachedOrFetchDataProduct retrieves a product from the in-memory cache, bulk-populating
// the cache on the first call via ListDataProducts().
func (c *Client) GetCachedOrFetchDataProduct(productID string) (*DataProduct, error) {
	c.cacheMutex.RLock()
	if c.productsWarmed {
		p, found := c.productsCache[productID]
		c.cacheMutex.RUnlock()
		if found {
			return p, nil
		}
		// If not in cache, fallback to GetDataProduct
		return c.GetDataProduct(productID)
	}
	c.cacheMutex.RUnlock()

	// Warm cache under write lock
	c.cacheMutex.Lock()
	if !c.productsWarmed {
		// Mark the bulk warm-up as attempted regardless of outcome. If the list call
		// is left un-flagged on failure, every subsequent read re-runs the full
		// paginated list under this write lock before falling back — strictly worse
		// than going straight to the per-UUID fetch below.
		c.productsWarmed = true
		if allProducts, err := c.ListDataProducts(); err == nil {
			for i := range allProducts {
				c.productsCache[allProducts[i].UUID] = &allProducts[i]
			}
		}
	}
	p, found := c.productsCache[productID]
	c.cacheMutex.Unlock()

	if found {
		return p, nil
	}
	return c.GetDataProduct(productID)
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

	if productResponse.DataProduct.UUID == "" {
		return nil, ErrEmptyValue
	}

	c.cacheMutex.Lock()
	if c.productsCache != nil {
		c.productsCache[productResponse.DataProduct.UUID] = &productResponse.DataProduct
	}
	c.cacheMutex.Unlock()

	return &productResponse.DataProduct, nil
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
	if err != nil {
		return err
	}

	c.cacheMutex.Lock()
	if c.productsCache != nil {
		delete(c.productsCache, productID)
	}
	c.cacheMutex.Unlock()

	return nil
}

// ValidateDataProduct dry-runs asset validation server-side. supported=false
// means the backend has no validate endpoint (pre-rollout) — callers should
// skip plan-time validation, not fail.
func (c *Client) ValidateDataProduct(dataProduct DataProduct) ([]APIErrorDetail, bool, error) {
	rb, err := json.Marshal(dataProduct)
	if err != nil {
		return nil, false, err
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-product/validate", c.HostURL),
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, false, err
	}
	if c.Token != "" {
		req.Header.Set("X-API-TOKEN", c.Token)
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusMethodNotAllowed {
		return nil, false, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, true, formatAPIError(res.StatusCode, body)
	}
	var parsed struct {
		Values []APIErrorDetail `json:"values"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, true, err
	}
	return parsed.Values, true, nil
}
