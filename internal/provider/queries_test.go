package provider

import (
	"strings"
	"testing"
)

// Verify GraphQL queries contain expected operation names and fields.
// These catch accidental query edits or truncations.

func TestQueryProductCreate_ContainsOperation(t *testing.T) {
	if !strings.Contains(queryProductCreate, "mutation productCreate") {
		t.Error("queryProductCreate should contain 'mutation productCreate'")
	}
	if !strings.Contains(queryProductCreate, "$input: ProductInput!") {
		t.Error("queryProductCreate should contain '$input: ProductInput!'")
	}
	if !strings.Contains(queryProductCreate, "variants(first: 1)") {
		t.Error("queryProductCreate should request first variant")
	}
	if !strings.Contains(queryProductCreate, "userErrors") {
		t.Error("queryProductCreate should request userErrors")
	}
}

func TestQueryProductUpdate_ContainsOperation(t *testing.T) {
	if !strings.Contains(queryProductUpdate, "mutation productUpdate") {
		t.Error("queryProductUpdate should contain 'mutation productUpdate'")
	}
	if !strings.Contains(queryProductUpdate, "$input: ProductInput!") {
		t.Error("queryProductUpdate should contain '$input: ProductInput!'")
	}
	if !strings.Contains(queryProductUpdate, "userErrors") {
		t.Error("queryProductUpdate should request userErrors")
	}
}

func TestQueryProductDelete_ContainsOperation(t *testing.T) {
	if !strings.Contains(queryProductDelete, "mutation productDelete") {
		t.Error("queryProductDelete should contain 'mutation productDelete'")
	}
	if !strings.Contains(queryProductDelete, "$input: ProductDeleteInput!") {
		t.Error("queryProductDelete should contain '$input: ProductDeleteInput!'")
	}
	if !strings.Contains(queryProductDelete, "deletedProductId") {
		t.Error("queryProductDelete should request deletedProductId")
	}
	if !strings.Contains(queryProductDelete, "userErrors") {
		t.Error("queryProductDelete should request userErrors")
	}
}

func TestQueryProductVariantsBulkUpdate_ContainsOperation(t *testing.T) {
	if !strings.Contains(queryProductVariantsBulkUpdate, "mutation productVariantsBulkUpdate") {
		t.Error("queryProductVariantsBulkUpdate should contain 'mutation productVariantsBulkUpdate'")
	}
	if !strings.Contains(queryProductVariantsBulkUpdate, "$productId: ID!") {
		t.Error("queryProductVariantsBulkUpdate should contain '$productId: ID!'")
	}
	if !strings.Contains(queryProductVariantsBulkUpdate, "$variants: [ProductVariantsBulkInput!]!") {
		t.Error("queryProductVariantsBulkUpdate should contain variant input type")
	}
	if !strings.Contains(queryProductVariantsBulkUpdate, "userErrors") {
		t.Error("queryProductVariantsBulkUpdate should request userErrors")
	}
}

func TestQueryProductRead_ContainsOperation(t *testing.T) {
	if !strings.Contains(queryProductRead, "query product") {
		t.Error("queryProductRead should contain 'query product'")
	}
	if !strings.Contains(queryProductRead, "$id: ID!") {
		t.Error("queryProductRead should contain '$id: ID!'")
	}
	if !strings.Contains(queryProductRead, "variants(first: 1)") {
		t.Error("queryProductRead should request first variant")
	}
}

func TestQueryProductRead_ContainsAllFields(t *testing.T) {
	expectedFields := []string{
		"id", "title", "handle", "status", "vendor",
		"productType", "descriptionHtml", "tags",
	}

	for _, field := range expectedFields {
		if !strings.Contains(queryProductRead, field) {
			t.Errorf("queryProductRead should contain field %q", field)
		}
	}
}

func TestQueryProductRead_ContainsVariantFields(t *testing.T) {
	variantFields := []string{"price", "compareAtPrice", "barcode", "sku"}

	for _, field := range variantFields {
		if !strings.Contains(queryProductRead, field) {
			t.Errorf("queryProductRead should contain variant field %q", field)
		}
	}
}
