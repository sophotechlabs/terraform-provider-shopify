resource "shopify_product" "example" {
  title            = "Example Product"
  handle           = "example-product"
  status           = "DRAFT"
  vendor           = "My Brand"
  product_type     = "Accessories"
  description_html = "<p>An example product managed by Terraform.</p>"
  tags             = ["example", "terraform"]

  price            = "29.99"
  compare_at_price = "39.99"
  sku              = "EXAMPLE-001"
  barcode          = "1234567890123"
}
