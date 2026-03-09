package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// mockShopifyServer creates an httptest server that simulates the Shopify GraphQL API.
// It tracks product state to support full CRUD lifecycle testing.
func mockShopifyServer(t *testing.T) *httptest.Server {
	t.Helper()

	var mu sync.Mutex
	products := map[string]*productData{}
	variantCounter := 100
	productCounter := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		body, _ := io.ReadAll(r.Body)
		var req graphQLRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to unmarshal request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "productCreate"):
			handleProductCreate(w, req, products, &productCounter, &variantCounter)

		case strings.Contains(req.Query, "productUpdate"):
			handleProductUpdate(w, req, products)

		case strings.Contains(req.Query, "productDelete"):
			handleProductDelete(w, req, products)

		case strings.Contains(req.Query, "productVariantsBulkUpdate"):
			handleVariantUpdate(w, req, products)

		case strings.Contains(req.Query, "query product"):
			handleProductRead(w, req, products)

		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, `{"errors":[{"message":"unknown query"}]}`)
		}
	}))
}

func handleProductCreate(w http.ResponseWriter, req graphQLRequest, products map[string]*productData, productCounter, variantCounter *int) {
	input, _ := req.Variables["input"].(map[string]any)
	*productCounter++
	*variantCounter++

	productID := fmt.Sprintf("gid://shopify/Product/%d", *productCounter)
	variantID := fmt.Sprintf("gid://shopify/ProductVariant/%d", *variantCounter)

	title, _ := input["title"].(string)
	status, _ := input["status"].(string)
	handle := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	if h, ok := input["handle"].(string); ok {
		handle = h
	}

	vendor := ""
	if v, ok := input["vendor"].(string); ok {
		vendor = v
	}
	productType := ""
	if pt, ok := input["productType"].(string); ok {
		productType = pt
	}
	descriptionHTML := ""
	if d, ok := input["descriptionHtml"].(string); ok {
		descriptionHTML = d
	}

	var tags []string
	if t, ok := input["tags"].([]any); ok {
		for _, tag := range t {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	defaultPrice := "0.00"

	product := &productData{
		ID:              productID,
		Title:           title,
		Handle:          handle,
		Status:          status,
		Vendor:          vendor,
		ProductType:     productType,
		DescriptionHTML: descriptionHTML,
		Tags:            tags,
	}
	product.Variants.Edges = []struct {
		Node variantData `json:"node"`
	}{
		{Node: variantData{ID: variantID, Price: defaultPrice}},
	}

	products[productID] = product

	resp := map[string]any{
		"data": map[string]any{
			"productCreate": map[string]any{
				"product":    product,
				"userErrors": []any{},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleProductUpdate(w http.ResponseWriter, req graphQLRequest, products map[string]*productData) {
	input, _ := req.Variables["input"].(map[string]any)
	id, _ := input["id"].(string)

	product, ok := products[id]
	if !ok {
		resp := map[string]any{
			"data": map[string]any{
				"productUpdate": map[string]any{
					"product":    nil,
					"userErrors": []map[string]any{{"field": []string{"id"}, "message": "Product not found"}},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if title, ok := input["title"].(string); ok {
		product.Title = title
	}
	if status, ok := input["status"].(string); ok {
		product.Status = status
	}
	if handle, ok := input["handle"].(string); ok {
		product.Handle = handle
	}
	if vendor, ok := input["vendor"].(string); ok {
		product.Vendor = vendor
	}
	if pt, ok := input["productType"].(string); ok {
		product.ProductType = pt
	}
	if d, ok := input["descriptionHtml"].(string); ok {
		product.DescriptionHTML = d
	}
	if t, ok := input["tags"].([]any); ok {
		var tags []string
		for _, tag := range t {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
		product.Tags = tags
	}

	// Return product without variants (matches real update query).
	respProduct := map[string]any{
		"id":              product.ID,
		"title":           product.Title,
		"handle":          product.Handle,
		"status":          product.Status,
		"vendor":          product.Vendor,
		"productType":     product.ProductType,
		"descriptionHtml": product.DescriptionHTML,
		"tags":            product.Tags,
	}

	resp := map[string]any{
		"data": map[string]any{
			"productUpdate": map[string]any{
				"product":    respProduct,
				"userErrors": []any{},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleProductDelete(w http.ResponseWriter, req graphQLRequest, products map[string]*productData) {
	input, _ := req.Variables["input"].(map[string]any)
	id, _ := input["id"].(string)

	delete(products, id)

	resp := map[string]any{
		"data": map[string]any{
			"productDelete": map[string]any{
				"deletedProductId": id,
				"userErrors":       []any{},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleVariantUpdate(w http.ResponseWriter, req graphQLRequest, products map[string]*productData) {
	productID, _ := req.Variables["productId"].(string)
	variants, _ := req.Variables["variants"].([]any)

	product, ok := products[productID]
	if !ok || len(variants) == 0 {
		resp := map[string]any{
			"data": map[string]any{
				"productVariantsBulkUpdate": map[string]any{
					"productVariants": []any{},
					"userErrors":      []map[string]any{{"field": []string{"productId"}, "message": "Product not found"}},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	variantInput, _ := variants[0].(map[string]any)
	variantID, _ := variantInput["id"].(string)

	// Update variant in product data.
	if len(product.Variants.Edges) > 0 {
		v := &product.Variants.Edges[0].Node
		if price, ok := variantInput["price"].(string); ok {
			v.Price = price
		}
		if sku, ok := variantInput["sku"].(string); ok {
			v.SKU = &sku
		}
		if barcode, ok := variantInput["barcode"].(string); ok {
			v.Barcode = &barcode
		}
		if cap, ok := variantInput["compareAtPrice"].(string); ok {
			v.CompareAtPrice = &cap
		}
	}

	variant := product.Variants.Edges[0].Node

	respVariant := map[string]any{
		"id":    variantID,
		"price": variant.Price,
	}
	if variant.SKU != nil {
		respVariant["sku"] = *variant.SKU
	}
	if variant.Barcode != nil {
		respVariant["barcode"] = *variant.Barcode
	}
	if variant.CompareAtPrice != nil {
		respVariant["compareAtPrice"] = *variant.CompareAtPrice
	}

	resp := map[string]any{
		"data": map[string]any{
			"productVariantsBulkUpdate": map[string]any{
				"productVariants": []any{respVariant},
				"userErrors":      []any{},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleProductRead(w http.ResponseWriter, req graphQLRequest, products map[string]*productData) {
	id, _ := req.Variables["id"].(string)

	product, ok := products[id]
	if !ok {
		resp := map[string]any{
			"data": map[string]any{
				"product": nil,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := map[string]any{
		"data": map[string]any{
			"product": product,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// testProviderFactoriesWithServer creates provider factories that point to a mock server.
func testProviderFactoriesWithServer(serverURL string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"shopify": providerserver.NewProtocol6WithError(
			&ShopifyProvider{version: "test", mockEndpoint: serverURL},
		),
	}
}

func TestAccProductResource_BasicCreate(t *testing.T) {
	server := mockShopifyServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactoriesWithServer(server.URL),
		Steps: []resource.TestStep{
			{
				Config: `
provider "shopify" {
  store_url    = "test-store.myshopify.com"
  access_token = "shpat_test_token"
  api_version  = "2025-04"
}

resource "shopify_product" "test" {
  title  = "Test Product"
  status = "DRAFT"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("shopify_product.test", "id"),
					resource.TestCheckResourceAttr("shopify_product.test", "title", "Test Product"),
					resource.TestCheckResourceAttr("shopify_product.test", "status", "DRAFT"),
					resource.TestCheckResourceAttr("shopify_product.test", "handle", "test-product"),
					resource.TestCheckResourceAttrSet("shopify_product.test", "default_variant_id"),
				),
			},
		},
	})
}

func TestAccProductResource_WithVariantFields(t *testing.T) {
	server := mockShopifyServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactoriesWithServer(server.URL),
		Steps: []resource.TestStep{
			{
				Config: `
provider "shopify" {
  store_url    = "test-store.myshopify.com"
  access_token = "shpat_test_token"
}

resource "shopify_product" "test" {
  title  = "Product With Variants"
  status = "ACTIVE"
  price  = "49.99"
  sku    = "SKU-001"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("shopify_product.test", "title", "Product With Variants"),
					resource.TestCheckResourceAttr("shopify_product.test", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("shopify_product.test", "price", "49.99"),
					resource.TestCheckResourceAttr("shopify_product.test", "sku", "SKU-001"),
				),
			},
		},
	})
}

func TestAccProductResource_WithAllFields(t *testing.T) {
	server := mockShopifyServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactoriesWithServer(server.URL),
		Steps: []resource.TestStep{
			{
				Config: `
provider "shopify" {
  store_url    = "test-store.myshopify.com"
  access_token = "shpat_test_token"
}

resource "shopify_product" "test" {
  title            = "Full Product"
  handle           = "full-product"
  status           = "DRAFT"
  vendor           = "Test Vendor"
  product_type     = "Widget"
  description_html = "<p>A test product</p>"
  tags             = ["test", "widget"]
  price            = "29.99"
  compare_at_price = "39.99"
  sku              = "FULL-001"
  barcode          = "1234567890123"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("shopify_product.test", "title", "Full Product"),
					resource.TestCheckResourceAttr("shopify_product.test", "handle", "full-product"),
					resource.TestCheckResourceAttr("shopify_product.test", "status", "DRAFT"),
					resource.TestCheckResourceAttr("shopify_product.test", "vendor", "Test Vendor"),
					resource.TestCheckResourceAttr("shopify_product.test", "product_type", "Widget"),
					resource.TestCheckResourceAttr("shopify_product.test", "description_html", "<p>A test product</p>"),
					resource.TestCheckResourceAttr("shopify_product.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("shopify_product.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("shopify_product.test", "tags.1", "widget"),
					resource.TestCheckResourceAttr("shopify_product.test", "price", "29.99"),
					resource.TestCheckResourceAttr("shopify_product.test", "compare_at_price", "39.99"),
					resource.TestCheckResourceAttr("shopify_product.test", "sku", "FULL-001"),
					resource.TestCheckResourceAttr("shopify_product.test", "barcode", "1234567890123"),
				),
			},
		},
	})
}

func TestAccProductResource_Update(t *testing.T) {
	server := mockShopifyServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactoriesWithServer(server.URL),
		Steps: []resource.TestStep{
			{
				Config: `
provider "shopify" {
  store_url    = "test-store.myshopify.com"
  access_token = "shpat_test_token"
}

resource "shopify_product" "test" {
  title  = "Original Title"
  status = "DRAFT"
}
`,
				Check: resource.TestCheckResourceAttr("shopify_product.test", "title", "Original Title"),
			},
			{
				Config: `
provider "shopify" {
  store_url    = "test-store.myshopify.com"
  access_token = "shpat_test_token"
}

resource "shopify_product" "test" {
  title  = "Updated Title"
  status = "ACTIVE"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("shopify_product.test", "title", "Updated Title"),
					resource.TestCheckResourceAttr("shopify_product.test", "status", "ACTIVE"),
				),
			},
		},
	})
}
