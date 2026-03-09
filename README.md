# terraform-provider-shopify

[![CI](https://github.com/sophotechlabs/terraform-provider-shopify/actions/workflows/ci.yml/badge.svg)](https://github.com/sophotechlabs/terraform-provider-shopify/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/sophotechlabs/terraform-provider-shopify/branch/main/graph/badge.svg)](https://codecov.io/gh/sophotechlabs/terraform-provider-shopify)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/sophotechlabs/terraform-provider-shopify/badge)](https://scorecard.dev/viewer/?uri=github.com/sophotechlabs/terraform-provider-shopify)
[![Go](https://img.shields.io/badge/go-1.24-blue.svg)](https://go.dev/)
[![Terraform](https://img.shields.io/badge/terraform-%3E%3D1.0-purple.svg)](https://www.terraform.io/)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-green.svg)](LICENSE)

Terraform provider for managing Shopify store resources via the GraphQL Admin API.

## Features (v0.1.0)

- Product management (create, update, delete, import)
- Default variant management (price, SKU, barcode, compare-at price)
- Plan/apply workflow with drift detection
- Environment variable and HCL configuration

## Quick Start

```hcl
terraform {
  required_providers {
    shopify = {
      source = "sophotechlabs/shopify"
    }
  }
}

provider "shopify" {
  store_url = "your-store.myshopify.com"
  # access_token via SHOPIFY_ACCESS_TOKEN env var
}

resource "shopify_product" "example" {
  title  = "Example Product"
  status = "DRAFT"
  price  = "29.99"
}
```

## Authentication

The provider requires a Shopify Admin API access token with product scopes.

### Creating a Custom App

1. Go to **Shopify Admin** > **Settings** > **Apps and sales channels**
2. Click **Develop apps** > **Create an app**
3. Under **Configuration**, set Admin API scopes:
   - `write_products` — create, update, delete products
   - `read_products` — read product data
4. Click **Install app** and copy the **Admin API access token**

### Provider Configuration

The provider accepts configuration via HCL or environment variables:

| HCL Attribute    | Environment Variable       | Required | Description                          |
|------------------|----------------------------|----------|--------------------------------------|
| `store_url`      | `SHOPIFY_STORE_URL`        | Yes      | Store URL (e.g. `your-store.myshopify.com`) |
| `access_token`   | `SHOPIFY_ACCESS_TOKEN`     | Yes      | Admin API access token (`shpat_...`) |
| `api_version`    | —                          | No       | API version (default: `2025-04`)     |

Environment variables are used as fallbacks when HCL attributes are not set.

```bash
export SHOPIFY_STORE_URL="your-store.myshopify.com"
export SHOPIFY_ACCESS_TOKEN="shpat_xxxxxxxxxxxxxxxxxxxxx"
```

## Resources

### shopify_product

Manages a Shopify product and its default variant.

```hcl
resource "shopify_product" "ring" {
  title            = "Silver Ring"
  handle           = "silver-ring"
  status           = "ACTIVE"
  vendor           = "My Brand"
  product_type     = "Jewelry"
  description_html = "<p>A beautiful silver ring.</p>"
  tags             = ["jewelry", "silver", "ring"]

  price            = "49.99"
  compare_at_price = "59.99"
  sku              = "RING-001"
  barcode          = "1234567890123"
}
```

#### Argument Reference

| Attribute          | Type         | Required | Description                                    |
|--------------------|--------------|----------|------------------------------------------------|
| `title`            | string       | Yes      | Product title                                  |
| `status`           | string       | Yes      | `ACTIVE`, `DRAFT`, or `ARCHIVED`               |
| `handle`           | string       | No       | URL-friendly handle (auto-generated from title) |
| `vendor`           | string       | No       | Product vendor                                 |
| `product_type`     | string       | No       | Product type                                   |
| `description_html` | string       | No       | Product description in HTML                    |
| `tags`             | list(string) | No       | Product tags                                   |
| `price`            | string       | No       | Default variant price                          |
| `compare_at_price` | string       | No       | Default variant compare-at price               |
| `sku`              | string       | No       | Default variant SKU                            |
| `barcode`          | string       | No       | Default variant barcode (EAN-13)               |

#### Attribute Reference

| Attribute            | Description                      |
|----------------------|----------------------------------|
| `id`                 | Shopify product GID              |
| `default_variant_id` | GID of the default variant       |

#### Import

Import existing products using their Shopify GID:

```bash
terraform import shopify_product.ring gid://shopify/Product/123456789
```

## Development

### Building from Source

```bash
git clone https://github.com/sophotechlabs/terraform-provider-shopify.git
cd terraform-provider-shopify
go build -o terraform-provider-shopify
```

### Local Development with dev_overrides

Add to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "sophotechlabs/shopify" = "/path/to/terraform-provider-shopify"
  }
  direct {}
}
```

Then run `terraform plan` / `terraform apply` without `terraform init`.

### Running Tests

```bash
go test ./...
```

### Generating Documentation

```bash
go generate ./...
```

## License

[Apache-2.0](LICENSE)

