package masthead

import (
	"fmt"
	"time"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	UserRoleOwner UserRole = "OWNER"
	UserRoleUser  UserRole = "USER"
)

// User represents a user in the system
type User struct {
	Email string   `json:"email"`
	Role  UserRole `json:"role"` // Role can be "OWNER" or "USER"
}

// UserResponse represents the response from the create/update user API
type UserResponse struct {
	User  User        `json:"value"`
	Extra interface{} `json:"extra"`
	Error interface{} `json:"error"`
}

// UsersResponse represents the response from the list users API
type UsersResponse struct {
	Users []User      `json:"values"`
	Extra interface{} `json:"extra"`
	Error interface{} `json:"error"`
}

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
type DataDomain struct {
	UUID             string       `json:"uuid"`
	Name             string       `json:"name"`
	Email            string       `json:"email"`
	SlackChannelName string       `json:"slackChannelName,omitempty"`
	SlackChannel     SlackChannel `json:"slackChannel"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

// DomainResponse represents the response from the create/update domain API
type DomainResponse struct {
	DataDomain DataDomain `json:"value"`
	Error      error      `json:"error,omitempty"`
	Message    string     `json:"message,omitempty"`
}

// DomainsResponse represents the response from the list domains API
type DomainListResponse struct {
	DataDomains []DataDomain `json:"values"`
	Pagination  Pagination   `json:"pagination"`
	Extra       interface{}  `json:"extra,omitempty"`
	Error       error        `json:"error,omitempty"`
	Message     string       `json:"message,omitempty"`
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
	DataDomainUUID string             `json:"dataDomainUuid"`
	Description    string             `json:"description"`
	DataDomain     *DataDomain        `json:"domain"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	DataAssets     []DataProductAsset `json:"dataAssets"`
}

// Validate checks if the DataProduct struct is valid
func (dp *DataProduct) Validate() error {
	if dp.DataDomain == nil {
		return fmt.Errorf("domain field is nil in DataProduct with UUID: %s", dp.UUID)
	}
	return nil
}

// DataProductResponse represents the response from the create/update data product API
type DataProductResponse struct {
	DataProduct DataProduct `json:"value"`
	Error       error       `json:"error,omitempty"`
	Message     string      `json:"message,omitempty"`
}

// DataProductListResponse represents the response from the list data products API
type DataProductListResponse struct {
	DataProducts []DataProduct `json:"values"`
	Pagination   Pagination    `json:"pagination"`
	Error        error         `json:"error,omitempty"`
	Message      error         `json:"message,omitempty"`
}
