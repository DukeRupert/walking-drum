-- Drop indices
DROP INDEX IF EXISTS idx_webhook_events_processed;
DROP INDEX IF EXISTS idx_webhook_events_type;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_customer_id;
DROP INDEX IF EXISTS idx_customer_addresses_customer_id;
DROP INDEX IF EXISTS idx_customers_email;
DROP INDEX IF EXISTS idx_prices_active;
DROP INDEX IF EXISTS idx_prices_product_id;
DROP INDEX IF EXISTS idx_products_active;

-- Drop tables in reverse order to handle dependencies
DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS customer_addresses;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS products;
