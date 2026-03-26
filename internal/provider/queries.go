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

const queryCollectionCreate = `
mutation collectionCreate($input: CollectionInput!) {
  collectionCreate(input: $input) {
    collection {
      id
      title
      handle
      descriptionHtml
      sortOrder
      templateSuffix
      ruleSet {
        appliedDisjunctively
        rules {
          column
          condition
          relation
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

const queryCollectionUpdate = `
mutation collectionUpdate($input: CollectionInput!) {
  collectionUpdate(input: $input) {
    collection {
      id
      title
      handle
      descriptionHtml
      sortOrder
      templateSuffix
      ruleSet {
        appliedDisjunctively
        rules {
          column
          condition
          relation
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

const queryCollectionDelete = `
mutation collectionDelete($input: CollectionDeleteInput!) {
  collectionDelete(input: $input) {
    deletedCollectionId
    userErrors {
      field
      message
    }
  }
}
`

const queryCollectionRead = `
query collection($id: ID!) {
  collection(id: $id) {
    id
    title
    handle
    descriptionHtml
    sortOrder
    templateSuffix
    ruleSet {
      appliedDisjunctively
      rules {
        column
        condition
        relation
      }
    }
  }
}
`
