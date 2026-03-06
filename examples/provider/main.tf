terraform {
  required_providers {
    shopify = {
      source = "sophotechlabs/shopify"
    }
  }
}

provider "shopify" {
  store_url    = var.store_url
  access_token = var.access_token
}
