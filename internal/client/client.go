package masthead

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HostURL - Default Masthead URL
const HostURL string = "https://metadata.mastheadata.com"

// TokenEnvVar - Environment variable for the Masthead API token
const TokenEnvVar string = "MASTHEAD_API_TOKEN"

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

func NewClient(token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		// Default Masthead URL
		HostURL: HostURL,
	}

	if token != nil {
		c.Token = *token
	} else if token := os.Getenv(TokenEnvVar); token != "" {
		// If the token is not provided, check for the environment variable
		// and set it as the token.
		// This allows the user to set the token in their environment
		// without having to pass it explicitly.
		// This is useful for CI/CD pipelines or other automated environments.
		// The environment variable is expected to be set as "MASTHEAD
		// _TOKEN" and will be used as the default token if not provided.
		c.Token = token
	} else {
		// If the token is not provided and the environment variable is not set,
		// return an error indicating that the token is required.
		return nil, fmt.Errorf("masthead API token is required. Set the token in the configuration or use the %s environment variable", TokenEnvVar)
	}

	return &c, nil
}

// doRequest performs an HTTP request and processes the response.
//
// It sets the authentication token in the request header if available,
// executes the request, and handles the response. If the response status
// is not OK (200), it returns an error with the status code and response body.
//
// Parameters:
//   - req: The HTTP request to be executed
//
// Returns:
//   - []byte: The response body as a byte slice
//   - error: An error if the request fails, the response cannot be read, or the status code is not 200
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.Token != "" {
		req.Header.Set("X-API-TOKEN", c.Token)
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
