package masthead

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyValueGuard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value": null}`))
	}))
	defer server.Close()

	token := "test-token"
	client, err := NewClient(&token, 0)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.HostURL = server.URL

	if _, err := client.CreateDataProduct(DataProduct{Name: "p"}); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("CreateDataProduct: want ErrEmptyValue, got %v", err)
	}
	if _, err := client.UpdateDataProduct(DataProduct{UUID: "u", Name: "p"}); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("UpdateDataProduct: want ErrEmptyValue, got %v", err)
	}
	if _, err := client.GetDataProduct("u"); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("GetDataProduct: want ErrEmptyValue, got %v", err)
	}
	if _, err := client.CreateDomain(DataDomain{Name: "d"}); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("CreateDomain: want ErrEmptyValue, got %v", err)
	}
	if _, err := client.GetDomain("u"); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("GetDomain: want ErrEmptyValue, got %v", err)
	}
}
