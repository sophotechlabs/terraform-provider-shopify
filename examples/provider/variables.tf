variable "store_url" {
  description = "Shopify store URL (e.g. your-store.myshopify.com)"
  type        = string
}

variable "access_token" {
  description = "Shopify Admin API access token"
  type        = string
  sensitive   = true
}
