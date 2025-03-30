package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// User represents a user in the system
type User struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	LastActivity string `json:"lastActivity,omitempty"`
	Status int `json:"status,omitempty"` // Status can be 0 - Pending, 1 - Active.
	Role  string `json:"role"` // Role can be "OWNER" or "USER"
}

// UserResponse represents the response from the create/update user API
type UserResponse struct {
	User User        `json:"value"`
	Extra interface{} `json:"extra"`
	Error interface{} `json:"error"`
}

// UsersResponse represents the response from the list users API
type UsersResponse struct {
	Users []User      `json:"values"`
	Extra  interface{} `json:"extra"`
	Error  interface{} `json:"error"`
}

func (c *Client) ListUsers() ([]User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/user/list", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	usersResponse := UsersResponse{}
	err = json.Unmarshal(body, &usersResponse)
	if err != nil {
		return nil, err
	}

	return usersResponse.Users, nil
}


func (c *Client) CreateUser(user User) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/clientApi/user", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	userResponse := UserResponse{}
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	return &userResponse.User, nil
}

func (c *Client) UpdateUserRole(user User) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/clientApi/user/role", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req) // Changed from `body, err = c.doRequest(req)` to `body, err := c.doRequest(req)`
	if err != nil {
		return nil, err
	}

	userResponse := UserResponse{}
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	return &userResponse.User, nil
}

// DeleteUser - Remove a user by the email address
func (c *Client) DeleteUser(email string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/clientApi/user/%s", c.HostURL, email), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
