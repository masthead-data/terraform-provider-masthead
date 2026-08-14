package masthead

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func validateServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestValidateDataProduct(t *testing.T) {
	token := "t"

	t.Run("errors returned", func(t *testing.T) {
		s := validateServer(200, `{"values":[{"type":"TABLE","project":"p","dataset":"d","table":"x","reason":"NOT_FOUND"}]}`)
		defer s.Close()
		c, _ := NewClient(&token, 0)
		c.HostURL = s.URL
		details, supported, err := c.ValidateDataProduct(DataProduct{Name: "p"})
		if err != nil || !supported || len(details) != 1 || details[0].Reason != "NOT_FOUND" {
			t.Errorf("got details=%v supported=%v err=%v", details, supported, err)
		}
	})

	t.Run("clean product", func(t *testing.T) {
		s := validateServer(200, `{"values":[]}`)
		defer s.Close()
		c, _ := NewClient(&token, 0)
		c.HostURL = s.URL
		details, supported, err := c.ValidateDataProduct(DataProduct{Name: "p"})
		if err != nil || !supported || len(details) != 0 {
			t.Errorf("got details=%v supported=%v err=%v", details, supported, err)
		}
	})

	t.Run("old backend 404 means unsupported", func(t *testing.T) {
		s := validateServer(404, `{"message":"Not Found"}`)
		defer s.Close()
		c, _ := NewClient(&token, 0)
		c.HostURL = s.URL
		_, supported, err := c.ValidateDataProduct(DataProduct{Name: "p"})
		if err != nil || supported {
			t.Errorf("want unsupported without error, got supported=%v err=%v", supported, err)
		}
	})

	t.Run("server error propagates", func(t *testing.T) {
		s := validateServer(500, `boom`)
		defer s.Close()
		c, _ := NewClient(&token, 0)
		c.HostURL = s.URL
		_, _, err := c.ValidateDataProduct(DataProduct{Name: "p"})
		if err == nil {
			t.Error("want error on 500")
		}
	})
}
