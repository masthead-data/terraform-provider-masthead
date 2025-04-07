package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ListDomains - Returns list of all data domains with pagination support
func (c *Client) ListDomains() ([]Domain, error) {
	var allDomains []Domain
	page := 1

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-domain/list?page=%d",
			c.HostURL, page), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}

		domainsResponse := DomainListResponse{}
		err = json.Unmarshal(body, &domainsResponse)
		if err != nil {
			return nil, err
		} else if domainsResponse.Error != nil {
			return nil, fmt.Errorf("error: %v. %v", domainsResponse.Error, domainsResponse.Message)
		}

		allDomains = append(allDomains, domainsResponse.Values...)

		// Break if we've retrieved all pages
		if len(domainsResponse.Values) == 0 || len(allDomains) >= domainsResponse.Pagination.Total {
			break
		}
		page++
	}

	return allDomains, nil
}

// CreateDomain - Create a new data domain in the system
func (c *Client) CreateDomain(domain Domain) (*Domain, error) {
	rb, err := json.Marshal(domain)
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

	domainResponse := DomainResponse{}
	err = json.Unmarshal(body, &domainResponse)
	if err != nil {
		return nil, err
	} else if domainResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", domainResponse.Error, domainResponse.Message)
	}

	return &domainResponse.Value, nil
}

// GetDomain - Get a specific data domain by ID
func (c *Client) GetDomain(domainID string) (*Domain, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domainID),
		nil,
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
	} else if domainResponse.Error != nil {
		return nil, fmt.Errorf("error: %v. %v", domainResponse.Error, domainResponse.Message)
	}

	return &domainResponse.Value, nil
}

// UpdateDomain - Update an existing data domain
func (c *Client) UpdateDomain(domain Domain) (*Domain, error) {
	if domain.UUID == "" {
		return nil, fmt.Errorf("domain UUID cannot be empty")
	}
	rb, err := json.Marshal(domain)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domain.UUID),
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

	return &domainResponse.Value, nil
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
	return err
}
