package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// User represents a user in the system
type User struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UsersResponse represents the response from the list users API
type UsersResponse struct {
	Values []User      `json:"values"`
	Extra  interface{} `json:"extra"`
	Error  interface{} `json:"error"`
}

// ListUsers - Returns list of all users
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

	return usersResponse.Values, nil
}

// CreateUser - Create a new user in the system
func (c *Client) CreateUser(email, role string) error {
	userReq := struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}{
		Email: email,
		Role:  role,
	}

	rb, err := json.Marshal(userReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/clientApi/user", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// UpdateUserRole - Update an existing user's role
// UpdateUserRole changes the role of a user in the Masthead system.
// It takes the user's email address and the new role to assign.
// Returns an error if the request cannot be created or if the API returns an error response.
func (c *Client) UpdateUserRole(email, role string) error {
	userReq := struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}{
		Email: email,
		Role:  role,
	}

	rb, err := json.Marshal(userReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/clientApi/user/role", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

// DeleteUser - Remove a user from the system by their email address
func (c *Client) DeleteUser(email string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/clientApi/user/%s", c.HostURL, email), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
