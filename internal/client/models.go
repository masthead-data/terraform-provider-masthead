package client

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
