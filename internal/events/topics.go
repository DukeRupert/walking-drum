// internal/events/topics.go
package events

// Product-related topics
const (
    TopicProductCreated     = "products.created"
    TopicProductUpdated     = "products.updated"
    TopicProductDeleted     = "products.deleted"
    TopicProductStockUpdated = "products.stock_updated"
    TopicProductLowStock    = "products.low_stock"
)

// Customer-related topics
const (
    TopicCustomerCreated    = "customers.created"
    TopicCustomerUpdated    = "customers.updated"
    TopicCustomerDeleted    = "customers.deleted"
)

// Subscription-related topics
const (
    TopicSubscriptionCreated  = "subscriptions.created"
    TopicSubscriptionUpdated  = "subscriptions.updated"
    TopicSubscriptionCanceled = "subscriptions.canceled"
    TopicSubscriptionPaused   = "subscriptions.paused"
    TopicSubscriptionResumed  = "subscriptions.resumed"
    TopicSubscriptionRenewed  = "subscriptions.renewed"
)

// Order-related topics
const (
    TopicOrderCreated        = "orders.created"
    TopicOrderStatusUpdated  = "orders.status_updated"
    TopicOrderShipped        = "orders.shipped"
    TopicOrderDelivered      = "orders.delivered"
)