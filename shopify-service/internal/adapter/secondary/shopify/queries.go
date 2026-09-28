package shopify

const (
	pageSize       = 10
	nestedPageSize = 250
)

const productFields = `
fragment ProductFields on Product {
  id
  title
  handle
  status
  vendor
  productType
  tags
  createdAt
  updatedAt
  variants(first: 50) {
    pageInfo { hasNextPage endCursor }
    nodes { ...VariantFields }
  }
}` + variantFields

const variantFields = `
fragment VariantFields on ProductVariant {
  id title sku price inventoryQuantity
}`

const listProductsQuery = `
query ListProducts($first: Int!, $after: String) {
  products(first: $first, after: $after) {
    pageInfo { hasNextPage endCursor }
    nodes { ...ProductFields }
  }
}` + productFields

const getProductQuery = `
query GetProduct($id: ID!) {
  product(id: $id) { ...ProductFields }
}` + productFields

const createProductMutation = `
mutation CreateProduct($product: ProductCreateInput!) {
  productCreate(product: $product) {
    product { ...ProductFields }
    userErrors { field message }
  }
}` + productFields

const orderFields = `
fragment OrderFields on Order {
  id
  name
  email
  displayFinancialStatus
  displayFulfillmentStatus
  currencyCode
  totalPriceSet { shopMoney { amount } }
  cancelledAt
  processedAt
  createdAt
  updatedAt
  lineItems(first: 50) {
    pageInfo { hasNextPage endCursor }
    nodes { ...LineItemFields }
  }
}` + lineItemFields

const lineItemFields = `
fragment LineItemFields on LineItem {
  id
  title
  sku
  quantity
  variant { id }
  originalUnitPriceSet { shopMoney { amount } }
}`

const listOrdersQuery = `
query ListOrders($first: Int!, $after: String) {
  orders(first: $first, after: $after, sortKey: PROCESSED_AT) {
    pageInfo { hasNextPage endCursor }
    nodes { ...OrderFields }
  }
}` + orderFields

const getOrderQuery = `
query GetOrder($id: ID!) {
  order(id: $id) { ...OrderFields }
}` + orderFields

const productVariantsQuery = `
query ProductVariants($id: ID!, $first: Int!, $after: String) {
  product(id: $id) {
    variants(first: $first, after: $after) {
      pageInfo { hasNextPage endCursor }
      nodes { ...VariantFields }
    }
  }
}` + variantFields

const orderLineItemsQuery = `
query OrderLineItems($id: ID!, $first: Int!, $after: String) {
  order(id: $id) {
    lineItems(first: $first, after: $after) {
      pageInfo { hasNextPage endCursor }
      nodes { ...LineItemFields }
    }
  }
}` + lineItemFields
