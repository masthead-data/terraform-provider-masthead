package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SlackChannel struct {
	ID   string `json:"channelId"`
	Name string `json:"channelName"`
}

// Domain represents a data domain in the system
type Domain struct {
	UUID           string `json:"uuid,omitempty"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	SlackChannelName string `json:"slackChannelName,omitempty"`
	SlackChannel SlackChannel `json:"slackChannel,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// DomainResponse represents the response from the create/update domain API
type DomainResponse struct {
	Value Domain `json:"value"`
	Extra  interface{} `json:"extra"`
	Error  interface{} `json:"error"`
}

// DomainsResponse represents the response from the list domains API
type DomainsResponse struct {
	Values []Domain `json:"values"`
	Extra  interface{}  `json:"extra"`
	Error  interface{}  `json:"error"`
}

// ListDomains - Returns list of all data domains
func (c *Client) ListDomains() ([]Domain, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-domain/list",
		c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	domainsResponse := DomainsResponse{}
	err = json.Unmarshal(body, &domainsResponse)
	if err != nil {
		return nil, err
	}

	return domainsResponse.Values, nil
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

	domainResponse := &Domain{}
	err = json.Unmarshal(body, domainResponse)
	if err != nil {
		return nil, err
	}

	return domainResponse, nil
}

// GetDomain - Get a specific data domain by ID
func (c *Client) GetDomain(domainID string) (*Domain, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domainID),
		nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	domainResponse := &Domain{}
	err = json.Unmarshal(body, domainResponse)
	if err != nil {
		return nil, err
	}

	return domainResponse, nil
}

// UpdateDomain - Update an existing data domain
func (c *Client) UpdateDomain(domain Domain) (*Domain, error) {
	rb, err := json.Marshal(domain)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domain.UUID),
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	domainResponse := &Domain{}
	err = json.Unmarshal(body, domainResponse)
	if err != nil {
		return nil, err
	}

	return domainResponse, nil
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
