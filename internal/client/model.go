package masthead

import (
	"fmt"
	"time"
)

// Pagination represents pagination details in API responses
type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
}

type SlackChannel struct {
	Name string `json:"channelName"`
	ID   string `json:"channelId"`
}

// Domain represents a data domain in the system
type Domain struct {
	UUID             string       `json:"uuid"`
	Name             string       `json:"name"`
	Email            string       `json:"email"`
	SlackChannelName string       `json:"slackChannelName,omitempty"`
	SlackChannel     SlackChannel `json:"slackChannel,omitempty"`
	CreatedAt        time.Time    `json:"createdAt,omitempty"`
	UpdatedAt        time.Time    `json:"updatedAt,omitempty"`
}

// DomainResponse represents the response from the create/update domain API
type DomainResponse struct {
	Value   Domain `json:"value"`
	Error   error  `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// DomainsResponse represents the response from the list domains API
type DomainListResponse struct {
	Values     []Domain    `json:"values"`
	Pagination Pagination  `json:"pagination"`
	Extra      interface{} `json:"extra,omitempty"`
	Error      error       `json:"error,omitempty"`
	Message    string      `json:"message,omitempty"`
}

// DataProductAssetType represents the type of a data asset
type DataProductAssetType string

const (
	DataProductAssetTypeDataset DataProductAssetType = "DATASET"
	DataProductAssetTypeTable   DataProductAssetType = "TABLE"
)

type AlertType string

const (
	AlertTypeRegular  AlertType = "REGULAR"
	AlertTypeCritical AlertType = "CRITICAL"
)

type DataProductAsset struct {
	Type      DataProductAssetType `json:"type"`
	UUID      string               `json:"uuid"`
	Project   string               `json:"project"`
	Dataset   string               `json:"dataset"`
	Table     string               `json:"table,omitempty"`
	AlertType AlertType            `json:"alertType"`
}

type DataProduct struct {
	UUID           string             `json:"uuid"`
	Name           string             `json:"name"`
	DataDomainUUID string             `json:"dataDomainUuid,omitempty"`
	Description    string             `json:"description"`
	Domain         *Domain            `json:"domain"`
	CreatedAt      time.Time          `json:"createdAt,omitempty"`
	UpdatedAt      time.Time          `json:"updatedAt,omitempty"`
	DataAssets     []DataProductAsset `json:"dataAssets"`
}

// Validate checks if the DataProduct struct is valid
func (dp *DataProduct) Validate() error {
	if dp.Domain == nil {
		return fmt.Errorf("domain field is nil in DataProduct with UUID: %s", dp.UUID)
	}
	return nil
}

// DataProductResponse represents the response from the create/update data product API
type DataProductResponse struct {
	Value   DataProduct `json:"value"`
	Error   error       `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// DataProductListResponse represents the response from the list data products API
type DataProductListResponse struct {
	Values     []DataProduct `json:"values"`
	Pagination Pagination    `json:"pagination"`
	Error      error         `json:"error,omitempty"`
	Message    error         `json:"message,omitempty"`
}
