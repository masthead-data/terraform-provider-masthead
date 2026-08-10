package masthead

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

// TestWarmUpAttemptedOnce pins the failure path of the bulk cache warm-up: a failing
// list endpoint must be attempted exactly once, after which every read falls back to
// the per-UUID endpoint. Retrying the full paginated list per resource is strictly
// worse than never caching at all.
func TestWarmUpAttemptedOnce(t *testing.T) {
	var mu sync.Mutex
	listCalls, getCalls := 0, 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/clientApi/data-product/list", "/clientApi/data-domain/list":
			listCalls++
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
		default:
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"value":{"uuid":%q,"name":"n"}}`, "u1")))
		}
	}))
	defer server.Close()

	token := "test-token"

	for _, tc := range []struct {
		name string
		get  func(*Client, string) error
	}{
		{"products", func(c *Client, id string) error {
			_, err := c.GetCachedOrFetchDataProduct(id)
			return err
		}},
		{"domains", func(c *Client, id string) error {
			_, err := c.GetCachedOrFetchDomain(id)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			listCalls, getCalls = 0, 0
			mu.Unlock()

			client, err := NewClient(&token, 0)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			client.HostURL = server.URL

			for i := 0; i < 5; i++ {
				if err := tc.get(client, "u1"); err != nil {
					t.Fatalf("read %d: %v", i, err)
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if listCalls != 1 {
				t.Errorf("list calls = %d, want 1 (warm-up must not retry per read)", listCalls)
			}
			if getCalls != 5 {
				t.Errorf("per-UUID calls = %d, want 5", getCalls)
			}
		})
	}
}

// TestListSendsPageAndLimit pins that both pagination params are sent. The API
// forwards page/limit to the backend only when both are present; sending page alone
// silently returns the first page every time, so a multi-page warm-up would loop over
// duplicates and never reach later pages.
func TestListSendsPageAndLimit(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		body string
		call func(*Client) error
	}{
		{
			name: "products",
			path: "/clientApi/data-product/list",
			body: `{"values":[],"pagination":{"total":0}}`,
			call: func(c *Client) error {
				_, err := c.ListDataProducts()
				return err
			},
		},
		{
			name: "domains",
			path: "/clientApi/data-domain/list",
			body: `{"values":[],"pagination":{"total":0}}`,
			call: func(c *Client) error {
				_, err := c.ListDomains()
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotQuery url.Values

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tc.path {
					gotQuery = r.URL.Query()
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			token := "test-token"
			client, err := NewClient(&token, 0)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			client.HostURL = server.URL

			if err := tc.call(client); err != nil {
				t.Fatalf("list: %v", err)
			}
			if gotQuery.Get("page") == "" {
				t.Errorf("page param missing, got %v", gotQuery)
			}
			if gotQuery.Get("limit") == "" {
				t.Errorf("limit param missing, got %v", gotQuery)
			}
		})
	}
}
