package provider

const queryProductCreate = `
mutation productCreate($input: ProductInput!) {
  productCreate(input: $input) {
    product {
      id
      title
      handle
      status
      vendor
      productType
      descriptionHtml
      tags
      variants(first: 1) {
        edges {
          node {
            id
            price
            compareAtPrice
            barcode
            sku
          }
        }
      }
    }
    userErrors {
      field
      message
    }
  }
}
`

const queryProductUpdate = `
mutation productUpdate($input: ProductInput!) {
  productUpdate(input: $input) {
    product {
      id
      title
      handle
      status
      vendor
      productType
      descriptionHtml
      tags
    }
    userErrors {
      field
      message
    }
  }
}
`

const queryProductDelete = `
mutation productDelete($input: ProductDeleteInput!) {
  productDelete(input: $input) {
    deletedProductId
    userErrors {
      field
      message
    }
  }
}
`

const queryProductVariantsBulkUpdate = `
mutation productVariantsBulkUpdate($productId: ID!, $variants: [ProductVariantsBulkInput!]!) {
  productVariantsBulkUpdate(productId: $productId, variants: $variants) {
    productVariants {
      id
      price
      compareAtPrice
      barcode
      sku
    }
    userErrors {
      field
      message
    }
  }
}
`

const queryProductRead = `
query product($id: ID!) {
  product(id: $id) {
    id
    title
    handle
    status
    vendor
    productType
    descriptionHtml
    tags
    variants(first: 1) {
      edges {
        node {
          id
          price
          compareAtPrice
          barcode
          sku
        }
      }
    }
  }
}
`
