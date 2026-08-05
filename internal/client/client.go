package masthead

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// HostURL - Default Masthead URL
const HostURL string = "https://metadata.mastheadata.com"

// TokenEnvVar - Environment variable for the Masthead API token
const TokenEnvVar string = "MASTHEAD_API_TOKEN"

type Client struct {
	HostURL        string
	HTTPClient     *http.Client
	Token          string
	productsCache  map[string]*DataProduct
	domainsCache   map[string]*DataDomain
	cacheMutex     sync.RWMutex
	productsWarmed bool
	domainsWarmed  bool
}

func NewClient(token *string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	}
	c := Client{
		HTTPClient:    &http.Client{Timeout: timeout, Transport: tr},
		HostURL:       HostURL,
		productsCache: make(map[string]*DataProduct),
		domainsCache:  make(map[string]*DataDomain),
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
		return nil, formatAPIError(res.StatusCode, body)
	}

	return body, err
}

// ErrEmptyValue indicates a 200 response whose value payload was null/empty —
// typically an upstream timeout masked as success; server state may or may not
// have been mutated.
var ErrEmptyValue = errors.New("API returned success with an empty value — likely an upstream timeout; verify server state before retrying")

type APIErrorDetail struct {
	Type      string `json:"type"`
	Project   string `json:"project,omitempty"`
	Dataset   string `json:"dataset,omitempty"`
	Table     string `json:"table,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	Reason    string `json:"reason"`
	DeletedAt string `json:"deletedAt,omitempty"`
}

type apiErrorBody struct {
	Message string           `json:"message"`
	Errors  []APIErrorDetail `json:"errors"`
}

// formatAPIError renders a non-200 response body. Bodies carrying a structured
// errors[] array become a per-asset diagnostic; anything else falls back to the raw dump.
func formatAPIError(status int, body []byte) error {
	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err != nil || len(parsed.Errors) == 0 {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "status: %d: %s:", status, parsed.Message)
	for _, e := range parsed.Errors {
		identity := e.UUID
		if e.Table != "" {
			identity = fmt.Sprintf("%s.%s.%s", e.Project, e.Dataset, e.Table)
		} else if e.Dataset != "" {
			identity = fmt.Sprintf("%s.%s", e.Project, e.Dataset)
		}
		fmt.Fprintf(&b, "\n  - %s %s — %s", e.Type, identity, e.Reason)
		if e.DeletedAt != "" {
			fmt.Fprintf(&b, " (%s)", e.DeletedAt)
		}
	}
	return errors.New(b.String())
}
