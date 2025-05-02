-- Drop indices in reverse order
DROP INDEX IF EXISTS idx_webhook_events_processed;
DROP INDEX IF EXISTS idx_webhook_events_type;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_customer_id;
DROP INDEX IF EXISTS idx_customer_addresses_customer_id;
DROP INDEX IF EXISTS idx_customers_email;
DROP INDEX IF EXISTS idx_product_prices_active;
DROP INDEX IF EXISTS idx_product_prices_product_id;
DROP INDEX IF EXISTS idx_products_active;