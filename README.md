# terraform-provider-shopify

Terraform provider for managing Shopify store resources via the GraphQL Admin API.

## Features

- Product management (create, update, delete) with full drift detection
- Default variant management (price, SKU, barcode, compare-at price)
- Environment variable and HCL configuration for credentials

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (to build the provider)

## Usage

```hcl
terraform {
  required_providers {
    shopify = {
      source  = "sophotechlabs/shopify"
      version = "~> 0.1"
    }
  }
}

provider "shopify" {
  store_url    = "your-store.myshopify.com"
  access_token = var.shopify_access_token
}

resource "shopify_product" "example" {
  title  = "Example Product"
  handle = "example-product"
  status = "DRAFT"

  price            = "29.99"
  compare_at_price = "39.99"
  sku              = "EX-001"
  barcode          = "1234567890123"
}
```

## Authentication

1. In your Shopify Admin, go to **Settings > Apps and sales channels > Develop apps**
2. Create an app and configure Admin API scopes:
   - `write_products`
   - `read_products`
3. Install the app and copy the Admin API access token

Set credentials via environment variables or provider config:

```bash
export SHOPIFY_STORE_URL="your-store.myshopify.com"
export SHOPIFY_ACCESS_TOKEN="shpat_xxxxx"
```

## Resources

### shopify_product

Manages a Shopify product and its default variant.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `title` | string | yes | Product title |
| `status` | string | yes | ACTIVE, DRAFT, or ARCHIVED |
| `handle` | string | no | URL-friendly handle (auto-generated from title if omitted) |
| `vendor` | string | no | Product vendor |
| `product_type` | string | no | Product type |
| `description_html` | string | no | Product description in HTML |
| `tags` | list(string) | no | Product tags |
| `price` | string | no | Default variant price |
| `compare_at_price` | string | no | Default variant compare-at price |
| `sku` | string | no | Default variant SKU |
| `barcode` | string | no | Default variant barcode |

#### Attributes

| Name | Description |
|------|-------------|
| `id` | Shopify product GID |
| `default_variant_id` | Default variant GID |

## Development

```bash
# Build
go build ./...

# Install locally
go install .

# Run tests
go test ./... -v
```

### Local testing with dev_overrides

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/sophotechlabs/shopify" = "/path/to/go/bin"
  }
  direct {}
}
```

## License

MPL-2.0
