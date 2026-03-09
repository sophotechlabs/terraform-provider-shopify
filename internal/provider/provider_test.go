package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// New tests

func TestNew_ReturnsFactory(t *testing.T) {
	factory := New("1.0.0")
	if factory == nil {
		t.Fatal("New should not return nil")
	}

	p := factory()
	if p == nil {
		t.Fatal("factory should not return nil provider")
	}

	sp, ok := p.(*ShopifyProvider)
	if !ok {
		t.Fatal("factory should return *ShopifyProvider")
	}
	if sp.version != "1.0.0" {
		t.Errorf("version = %q, want %q", sp.version, "1.0.0")
	}
}

func TestNew_DevVersion(t *testing.T) {
	factory := New("dev")
	p := factory()
	sp := p.(*ShopifyProvider)

	if sp.version != "dev" {
		t.Errorf("version = %q, want %q", sp.version, "dev")
	}
}

// Metadata tests

func TestShopifyProvider_Metadata(t *testing.T) {
	p := &ShopifyProvider{version: "1.0.0"}
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "shopify" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shopify")
	}
	if resp.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", resp.Version, "1.0.0")
	}
}

// Schema tests

func TestShopifyProvider_Schema_Attributes(t *testing.T) {
	p := &ShopifyProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	expectedAttrs := []string{"store_url", "access_token", "api_version"}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}

	expectedCount := len(expectedAttrs)
	actualCount := len(resp.Schema.Attributes)
	if actualCount != expectedCount {
		t.Errorf("schema has %d attributes, want %d", actualCount, expectedCount)
	}
}

func TestShopifyProvider_Schema_AllOptional(t *testing.T) {
	p := &ShopifyProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// All attributes optional (env var fallback).
	for name, attr := range resp.Schema.Attributes {
		if !attr.IsOptional() {
			t.Errorf("attribute %q should be optional (env var fallback)", name)
		}
	}
}

func TestShopifyProvider_Schema_AccessTokenSensitive(t *testing.T) {
	p := &ShopifyProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if !resp.Schema.Attributes["access_token"].IsSensitive() {
		t.Error("access_token attribute should be sensitive")
	}
}

func TestShopifyProvider_Schema_Description(t *testing.T) {
	p := &ShopifyProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	expected := "Manage Shopify store resources via the GraphQL Admin API."
	if resp.Schema.Description != expected {
		t.Errorf("Description = %q, want %q", resp.Schema.Description, expected)
	}
}

// Resources tests

func TestShopifyProvider_Resources(t *testing.T) {
	p := &ShopifyProvider{}
	resources := p.Resources(context.Background())

	if len(resources) != 1 {
		t.Fatalf("len(Resources) = %d, want 1", len(resources))
	}

	r := resources[0]()
	if _, ok := r.(*ProductResource); !ok {
		t.Error("first resource should be *ProductResource")
	}
}

// DataSources tests

func TestShopifyProvider_DataSources(t *testing.T) {
	p := &ShopifyProvider{}
	dataSources := p.DataSources(context.Background())

	if dataSources != nil {
		t.Errorf("DataSources should be nil, got %v", dataSources)
	}
}
