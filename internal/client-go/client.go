package masthead

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// HostURL - Default Masthead URL
const HostURL string = "https://metadata.mastheadata.com"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient -
func NewClient(host, token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		// Default Masthead URL
		HostURL: HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	if token != nil {
		c.Token = *token
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
