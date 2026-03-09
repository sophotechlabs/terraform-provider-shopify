package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hasVariantFields tests

func TestHasVariantFields_AllNull(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringNull(),
		CompareAtPrice: types.StringNull(),
		SKU:            types.StringNull(),
		Barcode:        types.StringNull(),
	}

	result := hasVariantFields(plan)
	if result {
		t.Error("hasVariantFields should return false when all variant fields are null")
	}
}

func TestHasVariantFields_AllUnknown(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringUnknown(),
		CompareAtPrice: types.StringUnknown(),
		SKU:            types.StringUnknown(),
		Barcode:        types.StringUnknown(),
	}

	result := hasVariantFields(plan)
	if result {
		t.Error("hasVariantFields should return false when all variant fields are unknown")
	}
}

func TestHasVariantFields_PriceSet(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringValue("29.99"),
		CompareAtPrice: types.StringNull(),
		SKU:            types.StringNull(),
		Barcode:        types.StringNull(),
	}

	result := hasVariantFields(plan)
	if !result {
		t.Error("hasVariantFields should return true when price is set")
	}
}

func TestHasVariantFields_SKUSet(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringNull(),
		CompareAtPrice: types.StringNull(),
		SKU:            types.StringValue("TEST-001"),
		Barcode:        types.StringNull(),
	}

	result := hasVariantFields(plan)
	if !result {
		t.Error("hasVariantFields should return true when SKU is set")
	}
}

func TestHasVariantFields_CompareAtPriceSet(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringNull(),
		CompareAtPrice: types.StringValue("39.99"),
		SKU:            types.StringNull(),
		Barcode:        types.StringNull(),
	}

	result := hasVariantFields(plan)
	if !result {
		t.Error("hasVariantFields should return true when compare_at_price is set")
	}
}

func TestHasVariantFields_BarcodeSet(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringNull(),
		CompareAtPrice: types.StringNull(),
		SKU:            types.StringNull(),
		Barcode:        types.StringValue("1234567890123"),
	}

	result := hasVariantFields(plan)
	if !result {
		t.Error("hasVariantFields should return true when barcode is set")
	}
}

func TestHasVariantFields_AllSet(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringValue("29.99"),
		CompareAtPrice: types.StringValue("39.99"),
		SKU:            types.StringValue("TEST-001"),
		Barcode:        types.StringValue("1234567890123"),
	}

	result := hasVariantFields(plan)
	if !result {
		t.Error("hasVariantFields should return true when all variant fields are set")
	}
}

func TestHasVariantFields_MixedNullAndUnknown(t *testing.T) {
	plan := ProductResourceModel{
		Price:          types.StringNull(),
		CompareAtPrice: types.StringUnknown(),
		SKU:            types.StringNull(),
		Barcode:        types.StringUnknown(),
	}

	result := hasVariantFields(plan)
	if result {
		t.Error("hasVariantFields should return false when fields are mix of null and unknown")
	}
}

// nullableString tests

func TestNullableString_NilReturnsNull(t *testing.T) {
	result := nullableString(nil)

	if !result.IsNull() {
		t.Errorf("nullableString(nil) should be null, got %q", result.ValueString())
	}
}

func TestNullableString_ValueReturnsString(t *testing.T) {
	value := "hello"
	result := nullableString(&value)

	if result.IsNull() {
		t.Error("nullableString with value should not be null")
	}

	expected := "hello"
	if result.ValueString() != expected {
		t.Errorf("nullableString value = %q, want %q", result.ValueString(), expected)
	}
}

func TestNullableString_EmptyStringReturnsValue(t *testing.T) {
	value := ""
	result := nullableString(&value)

	if result.IsNull() {
		t.Error("nullableString with empty string should not be null")
	}
	if result.ValueString() != "" {
		t.Errorf("nullableString value = %q, want empty string", result.ValueString())
	}
}

// formatUserErrors tests

func TestFormatUserErrors_SingleError(t *testing.T) {
	errors := []userError{
		{Field: []string{"title"}, Message: "Title can't be blank"},
	}

	result := formatUserErrors(errors)
	expected := "Title can't be blank"

	if result != expected {
		t.Errorf("formatUserErrors = %q, want %q", result, expected)
	}
}

func TestFormatUserErrors_MultipleErrors(t *testing.T) {
	errors := []userError{
		{Field: []string{"title"}, Message: "Title can't be blank"},
		{Field: []string{"status"}, Message: "Status is invalid"},
	}

	result := formatUserErrors(errors)
	expected := "Title can't be blank; Status is invalid"

	if result != expected {
		t.Errorf("formatUserErrors = %q, want %q", result, expected)
	}
}

func TestFormatUserErrors_EmptyErrors(t *testing.T) {
	result := formatUserErrors([]userError{})
	expected := ""

	if result != expected {
		t.Errorf("formatUserErrors = %q, want %q", result, expected)
	}
}

func TestFormatUserErrors_ThreeErrors(t *testing.T) {
	errors := []userError{
		{Field: []string{"title"}, Message: "First"},
		{Field: []string{"status"}, Message: "Second"},
		{Field: []string{"handle"}, Message: "Third"},
	}

	result := formatUserErrors(errors)
	expected := "First; Second; Third"

	if result != expected {
		t.Errorf("formatUserErrors = %q, want %q", result, expected)
	}
}

// ProductResource metadata tests

func TestProductResource_Metadata(t *testing.T) {
	r := &ProductResource{}
	req := resource.MetadataRequest{ProviderTypeName: "shopify"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "shopify_product"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, expected)
	}
}

func TestProductResource_Metadata_DifferentProvider(t *testing.T) {
	r := &ProductResource{}
	req := resource.MetadataRequest{ProviderTypeName: "test"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "test_product"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, expected)
	}
}

// ProductResource schema tests

func TestProductResource_Schema_RequiredAttributes(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	requiredAttrs := []string{"title", "status"}
	for _, name := range requiredAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("schema missing required attribute %q", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("attribute %q should be required", name)
		}
	}
}

func TestProductResource_Schema_ComputedAttributes(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	computedOnlyAttrs := []string{"id", "default_variant_id"}
	for _, name := range computedOnlyAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("schema missing attribute %q", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %q should be computed", name)
		}
	}
}

func TestProductResource_Schema_OptionalComputedAttributes(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	optionalComputedAttrs := []string{
		"handle", "vendor", "product_type", "description_html",
		"price", "compare_at_price", "sku", "barcode",
	}
	for _, name := range optionalComputedAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("schema missing attribute %q", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %q should be optional", name)
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %q should be computed", name)
		}
	}
}

func TestProductResource_Schema_TagsAttribute(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	attr, ok := resp.Schema.Attributes["tags"]
	if !ok {
		t.Fatal("schema missing 'tags' attribute")
	}
	if !attr.IsOptional() {
		t.Error("tags attribute should be optional")
	}
}

func TestProductResource_Schema_AllAttributesPresent(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	expectedAttrs := []string{
		"id", "title", "handle", "status", "vendor", "product_type",
		"description_html", "tags", "default_variant_id",
		"price", "compare_at_price", "sku", "barcode",
	}

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

func TestProductResource_Schema_Description(t *testing.T) {
	r := &ProductResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	expected := "Manages a Shopify product and its default variant."
	if resp.Schema.Description != expected {
		t.Errorf("Description = %q, want %q", resp.Schema.Description, expected)
	}
}

// NewProductResource tests

func TestNewProductResource_ReturnsResource(t *testing.T) {
	r := NewProductResource()
	if r == nil {
		t.Fatal("NewProductResource should not return nil")
	}

	_, ok := r.(*ProductResource)
	if !ok {
		t.Error("NewProductResource should return *ProductResource")
	}
}

// Configure tests

func TestProductResource_Configure_NilProviderData(t *testing.T) {
	r := &ProductResource{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Configure with nil ProviderData should not produce errors")
	}
	if r.client != nil {
		t.Error("client should be nil when ProviderData is nil")
	}
}

func TestProductResource_Configure_ValidClient(t *testing.T) {
	client := NewClient("test.myshopify.com", "token", "2025-04")
	r := &ProductResource{}
	req := resource.ConfigureRequest{ProviderData: client}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure should not produce errors, got: %v", resp.Diagnostics.Errors())
	}
	if r.client != client {
		t.Error("client should be set to provided Client")
	}
}

func TestProductResource_Configure_WrongType(t *testing.T) {
	r := &ProductResource{}
	req := resource.ConfigureRequest{ProviderData: "not a client"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure should produce error for wrong provider data type")
	}
}
