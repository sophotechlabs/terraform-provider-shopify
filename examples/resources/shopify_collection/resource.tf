# Manual collection
resource "shopify_collection" "featured" {
  title            = "Featured Products"
  handle           = "featured"
  description_html = "<p>Our handpicked featured products.</p>"
  sort_order       = "MANUAL"
}

# Smart collection with rules
resource "shopify_collection" "sale" {
  title            = "Sale Items"
  handle           = "sale"
  description_html = "<p>Products currently on sale.</p>"
  sort_order       = "PRICE_ASC"

  rule_set = {
    applied_disjunctively = false
    rules = [
      {
        column    = "IS_PRICE_REDUCED"
        relation  = "IS_SET"
        condition = "true"
      }
    ]
  }
}
