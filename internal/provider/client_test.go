package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// NewClient tests

func TestNewClient_SetsFields(t *testing.T) {
	client := NewClient("test-store.myshopify.com", "shpat_test123", "2025-04")

	if client.StoreURL != "test-store.myshopify.com" {
		t.Errorf("StoreURL = %q, want %q", client.StoreURL, "test-store.myshopify.com")
	}
	if client.AccessToken != "shpat_test123" {
		t.Errorf("AccessToken = %q, want %q", client.AccessToken, "shpat_test123")
	}
	if client.APIVersion != "2025-04" {
		t.Errorf("APIVersion = %q, want %q", client.APIVersion, "2025-04")
	}

	expected := "https://test-store.myshopify.com/admin/api/2025-04/graphql.json"
	if client.Endpoint != expected {
		t.Errorf("Endpoint = %q, want %q", client.Endpoint, expected)
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestNewClient_EndpointFormat(t *testing.T) {
	tests := []struct {
		name       string
		storeURL   string
		apiVersion string
		expected   string
	}{
		{
			name:       "standard store URL",
			storeURL:   "my-shop.myshopify.com",
			apiVersion: "2025-04",
			expected:   "https://my-shop.myshopify.com/admin/api/2025-04/graphql.json",
		},
		{
			name:       "different API version",
			storeURL:   "other-store.myshopify.com",
			apiVersion: "2024-10",
			expected:   "https://other-store.myshopify.com/admin/api/2024-10/graphql.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.storeURL, "token", tt.apiVersion)
			if client.Endpoint != tt.expected {
				t.Errorf("Endpoint = %q, want %q", client.Endpoint, tt.expected)
			}
		})
	}
}

// Execute tests

func TestExecute_SuccessfulQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers.
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-Shopify-Access-Token") != "shpat_test" {
			t.Errorf("X-Shopify-Access-Token = %q, want shpat_test", r.Header.Get("X-Shopify-Access-Token"))
		}

		// Verify request body structure.
		body, _ := io.ReadAll(r.Body)
		var req graphQLRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		if req.Query != "{ shop { name } }" {
			t.Errorf("Query = %q, want %q", req.Query, "{ shop { name } }")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"shop": {"name": "Test Store"}}}`))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:    server.URL,
		AccessToken: "shpat_test",
		HTTPClient:  server.Client(),
	}

	result, err := client.Execute("{ shop { name } }", nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.Data == nil {
		t.Fatal("result.Data is nil")
	}

	var data map[string]map[string]string
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal result data: %v", err)
	}

	expected := "Test Store"
	if data["shop"]["name"] != expected {
		t.Errorf("shop.name = %q, want %q", data["shop"]["name"], expected)
	}
}

func TestExecute_WithVariables(t *testing.T) {
	var receivedVars map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req graphQLRequest
		_ = json.Unmarshal(body, &req)
		receivedVars = req.Variables

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	variables := map[string]any{
		"id": "gid://shopify/Product/123",
	}
	_, err := client.Execute("query($id: ID!) { product(id: $id) { id } }", variables)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if receivedVars["id"] != "gid://shopify/Product/123" {
		t.Errorf("variables.id = %v, want gid://shopify/Product/123", receivedVars["id"])
	}
}

func TestExecute_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	result, err := client.Execute("{ shop { name } }", nil)
	if err == nil {
		t.Fatal("Execute should return error for non-200 status")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}

	expectedMsg := "shopify API returned status 500"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("error = %q, should contain %q", err.Error(), expectedMsg)
	}
}

func TestExecute_GraphQLErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": null, "errors": [{"message": "Access denied"}]}`))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	result, err := client.Execute("{ shop { name } }", nil)
	if err == nil {
		t.Fatal("Execute should return error for GraphQL errors")
	}
	if result != nil {
		t.Error("result should be nil on GraphQL error")
	}

	expectedMsg := "graphql errors: Access denied"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestExecute_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	result, err := client.Execute("{ shop { name } }", nil)
	if err == nil {
		t.Fatal("Execute should return error for invalid JSON response")
	}
	if result != nil {
		t.Error("result should be nil on unmarshal error")
	}

	expectedMsg := "unmarshaling response"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("error = %q, should contain %q", err.Error(), expectedMsg)
	}
}

func TestExecute_ConnectionError(t *testing.T) {
	client := &Client{
		Endpoint:   "http://127.0.0.1:1", // Nothing listening.
		HTTPClient: &http.Client{},
	}

	result, err := client.Execute("{ shop { name } }", nil)
	if err == nil {
		t.Fatal("Execute should return error for connection failure")
	}
	if result != nil {
		t.Error("result should be nil on connection error")
	}

	expectedMsg := "executing request"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("error = %q, should contain %q", err.Error(), expectedMsg)
	}
}

func TestExecute_MultipleGraphQLErrors_ReturnsFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := `{"data": null, "errors": [{"message": "First error"}, {"message": "Second error"}]}`
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	_, err := client.Execute("{ shop { name } }", nil)
	if err == nil {
		t.Fatal("Execute should return error")
	}

	expectedMsg := "graphql errors: First error"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestExecute_ProductCreateResponse(t *testing.T) {
	responseJSON := `{
		"data": {
			"productCreate": {
				"product": {
					"id": "gid://shopify/Product/123456789",
					"title": "Test Product",
					"handle": "test-product",
					"status": "DRAFT",
					"vendor": "Test Vendor",
					"productType": "Widget",
					"descriptionHtml": "<p>A test product</p>",
					"tags": ["test", "widget"],
					"variants": {
						"edges": [{
							"node": {
								"id": "gid://shopify/ProductVariant/987654321",
								"price": "29.99",
								"compareAtPrice": "39.99",
								"barcode": "1234567890123",
								"sku": "TEST-001"
							}
						}]
					}
				},
				"userErrors": []
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	result, err := client.Execute(queryProductCreate, map[string]any{
		"input": map[string]any{"title": "Test Product", "status": "DRAFT"},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var data productCreateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal product create data: %v", err)
	}

	product := data.ProductCreate.Product
	if product == nil {
		t.Fatal("product should not be nil")
	}
	if product.ID != "gid://shopify/Product/123456789" {
		t.Errorf("product.ID = %q, want gid://shopify/Product/123456789", product.ID)
	}
	if product.Title != "Test Product" {
		t.Errorf("product.Title = %q, want %q", product.Title, "Test Product")
	}
	if product.Handle != "test-product" {
		t.Errorf("product.Handle = %q, want %q", product.Handle, "test-product")
	}
	if product.Status != "DRAFT" {
		t.Errorf("product.Status = %q, want %q", product.Status, "DRAFT")
	}
	if len(product.Tags) != 2 {
		t.Fatalf("len(product.Tags) = %d, want 2", len(product.Tags))
	}
	if product.Tags[0] != "test" || product.Tags[1] != "widget" {
		t.Errorf("product.Tags = %v, want [test widget]", product.Tags)
	}
	if len(product.Variants.Edges) != 1 {
		t.Fatalf("len(product.Variants.Edges) = %d, want 1", len(product.Variants.Edges))
	}

	variant := product.Variants.Edges[0].Node
	if variant.ID != "gid://shopify/ProductVariant/987654321" {
		t.Errorf("variant.ID = %q, want gid://shopify/ProductVariant/987654321", variant.ID)
	}
	if variant.Price != "29.99" {
		t.Errorf("variant.Price = %q, want %q", variant.Price, "29.99")
	}
	if len(data.ProductCreate.UserErrors) != 0 {
		t.Errorf("userErrors should be empty, got %v", data.ProductCreate.UserErrors)
	}
}

func TestExecute_ProductCreateWithUserErrors(t *testing.T) {
	responseJSON := `{
		"data": {
			"productCreate": {
				"product": null,
				"userErrors": [
					{"field": ["title"], "message": "Title can't be blank"},
					{"field": ["status"], "message": "Status is invalid"}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	result, err := client.Execute(queryProductCreate, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var data productCreateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if data.ProductCreate.Product != nil {
		t.Error("product should be nil when userErrors present")
	}
	if len(data.ProductCreate.UserErrors) != 2 {
		t.Fatalf("len(userErrors) = %d, want 2", len(data.ProductCreate.UserErrors))
	}
	if data.ProductCreate.UserErrors[0].Message != "Title can't be blank" {
		t.Errorf("userErrors[0].Message = %q, want %q", data.ProductCreate.UserErrors[0].Message, "Title can't be blank")
	}
}
