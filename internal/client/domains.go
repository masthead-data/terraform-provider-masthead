package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DataDomain represents a data domain in the system
type DataDomain struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	SlackChannel string `json:"_slackChannel,omitempty"`
}

// DataDomainsResponse represents the response from the list domains API
type DataDomainsResponse struct {
	Values []DataDomain `json:"values"`
	Extra  interface{}  `json:"extra"`
	Error  interface{}  `json:"error"`
}

// ListDataDomains - Returns list of all data domains
func (c *Client) ListDataDomains() ([]DataDomain, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-domain/list",
		c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	domainsResponse := DataDomainsResponse{}
	err = json.Unmarshal(body, &domainsResponse)
	if err != nil {
		return nil, err
	}

	return domainsResponse.Values, nil
}

// CreateDataDomain - Create a new data domain in the system
func (c *Client) CreateDataDomain(name, email string, slackChannelName *string) error {
	domainReq := DataDomain{
		Name:  name,
		Email: email,
	}

	if slackChannelName != nil {
		domainReq.SlackChannel = *slackChannelName
	}

	rb, err := json.Marshal(domainReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-domain", c.HostURL),
		strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// GetDataDomain - Get a specific data domain by ID
func (c *Client) GetDataDomain(domainID string) (*DataDomain, error) {
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

	domain := &DataDomain{}
	err = json.Unmarshal(body, domain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

// UpdateDataDomain - Update an existing data domain
func (c *Client) UpdateDataDomain(domainID string, name, email string, slackChannelName *string) error {
	domainReq := DataDomain{
		ID:    domainID,
		Name:  name,
		Email: email,
	}

	if slackChannelName != nil {
		domainReq.SlackChannel = *slackChannelName
	}

	rb, err := json.Marshal(domainReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domainID),
		strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// DeleteDataDomain - Remove a data domain from the system by ID
func (c *Client) DeleteDataDomain(domainID string) error {
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/clientApi/data-domain/%s", c.HostURL, domainID),
		nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
