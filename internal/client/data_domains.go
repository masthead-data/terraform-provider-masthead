package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ListDomains - Returns list of all data domains with pagination support
func (c *Client) ListDomains() ([]DataDomain, error) {
	var allDataDomains []DataDomain
	page := 1

	for {
		// limit must be sent alongside page: the API forwards the pair only when both
		// are present, otherwise it silently falls back to page=1 and returns the same
		// first page on every iteration.
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-domain/list?page=%d&limit=100",
			c.HostURL, page), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}

		dataDomainsResponse := DomainListResponse{}
		err = json.Unmarshal(body, &dataDomainsResponse)
		if err != nil {
			return nil, err
		} else if dataDomainsResponse.Error != nil {
			return nil, fmt.Errorf("error: %v. %v", dataDomainsResponse.Error, dataDomainsResponse.Message)
		}

		allDataDomains = append(allDataDomains, dataDomainsResponse.DataDomains...)

		// Break if we've retrieved all pages
		if len(dataDomainsResponse.DataDomains) == 0 || len(allDataDomains) >= dataDomainsResponse.Pagination.Total {
			break
		}
		page++
	}

	return allDataDomains, nil
}

// CreateDomain - Create a new data domain in the system
func (c *Client) CreateDomain(dataDomain DataDomain) (*DataDomain, error) {
	rb, err := json.Marshal(dataDomain)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-domain", c.HostURL),
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataDomainResponse := DomainResponse{}
	err = json.Unmarshal(body, &dataDomainResponse)
	if err != nil {
		return nil, err
	} else if dataDomainResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", dataDomainResponse.Error, dataDomainResponse.Message)
	}

	if dataDomainResponse.DataDomain.UUID == "" {
		return nil, ErrEmptyValue
	}

	c.cacheMutex.Lock()
	if c.domainsCache != nil {
		c.domainsCache[dataDomainResponse.DataDomain.UUID] = &dataDomainResponse.DataDomain
	}
	c.cacheMutex.Unlock()

	return &dataDomainResponse.DataDomain, nil
}

// GetDomain - Get a specific data domain by ID
func (c *Client) GetDomain(dataDomainID string) (*DataDomain, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, dataDomainID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataDomainResponse := DomainResponse{}
	err = json.Unmarshal(body, &dataDomainResponse)
	if err != nil {
		return nil, err
	} else if dataDomainResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", dataDomainResponse.Error, dataDomainResponse.Message)
	}

	if dataDomainResponse.DataDomain.UUID == "" {
		return nil, ErrEmptyValue
	}

	return &dataDomainResponse.DataDomain, nil
}

// GetCachedOrFetchDomain retrieves a domain from in-memory cache, bulk-populating
// the cache on first call via ListDomains().
func (c *Client) GetCachedOrFetchDomain(dataDomainID string) (*DataDomain, error) {
	c.cacheMutex.RLock()
	if c.domainsWarmed {
		d, found := c.domainsCache[dataDomainID]
		c.cacheMutex.RUnlock()
		if found {
			return d, nil
		}
		return c.GetDomain(dataDomainID)
	}
	c.cacheMutex.RUnlock()

	c.cacheMutex.Lock()
	if !c.domainsWarmed {
		// Mark the bulk warm-up as attempted regardless of outcome. If the list call
		// is left un-flagged on failure, every subsequent read re-runs the full
		// paginated list under this write lock before falling back — strictly worse
		// than going straight to the per-UUID fetch below.
		c.domainsWarmed = true
		if allDomains, err := c.ListDomains(); err == nil {
			for i := range allDomains {
				c.domainsCache[allDomains[i].UUID] = &allDomains[i]
			}
		}
	}
	d, found := c.domainsCache[dataDomainID]
	c.cacheMutex.Unlock()

	if found {
		return d, nil
	}
	return c.GetDomain(dataDomainID)
}

// UpdateDomain - Update an existing data domain
func (c *Client) UpdateDomain(dataDomain DataDomain) (*DataDomain, error) {
	if dataDomain.UUID == "" {
		return nil, fmt.Errorf("domain UUID cannot be empty")
	}
	rb, err := json.Marshal(dataDomain)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, dataDomain.UUID),
		strings.NewReader(string(rb)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	domainResponse := DomainResponse{}
	err = json.Unmarshal(body, &domainResponse)
	if err != nil {
		return nil, err
	}
	if domainResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", domainResponse.Error, domainResponse.Message)
	}

	if domainResponse.DataDomain.UUID == "" {
		return nil, ErrEmptyValue
	}

	c.cacheMutex.Lock()
	if c.domainsCache != nil {
		c.domainsCache[domainResponse.DataDomain.UUID] = &domainResponse.DataDomain
	}
	c.cacheMutex.Unlock()

	return &domainResponse.DataDomain, nil
}

// DeleteDomain - Remove a data domain from the system by ID
func (c *Client) DeleteDomain(domainID string) error {
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domainID),
		nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	c.cacheMutex.Lock()
	if c.domainsCache != nil {
		delete(c.domainsCache, domainID)
	}
	c.cacheMutex.Unlock()

	return nil
}
