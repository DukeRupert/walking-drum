-- +goose Up
-- +goose StatementBegin
-- Insert test users
INSERT INTO users (id, email, password_hash, name, stripe_customer_id)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Test User', 'cus_test123'),
  ('22222222-2222-2222-2222-222222222222', 'admin@walkingdrum.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Admin User', 'cus_admin456');

-- Insert test products
INSERT INTO products (id, name, description, stripe_product_id)
VALUES
  ('33333333-3333-3333-3333-333333333333', 'Basic Plan', 'Basic subscription with core features', 'prod_basic789'),
  ('44444444-4444-4444-4444-444444444444', 'Premium Plan', 'Premium subscription with advanced features', 'prod_premium012');

-- Insert test prices
INSERT INTO prices (id, product_id, amount, currency, interval_type, interval_count, stripe_price_id, nickname)
VALUES
  ('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', 1000, 'usd', 'month', 1, 'price_basic_monthly', 'Basic Monthly'),
  ('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', 10000, 'usd', 'year', 1, 'price_basic_yearly', 'Basic Yearly'),
  ('77777777-7777-7777-7777-777777777777', '44444444-4444-4444-4444-444444444444', 2500, 'usd', 'month', 1, 'price_premium_monthly', 'Premium Monthly'),
  ('88888888-8888-8888-8888-888888888888', '44444444-4444-4444-4444-444444444444', 25000, 'usd', 'year', 1, 'price_premium_yearly', 'Premium Yearly');

-- Insert test subscriptions
INSERT INTO subscriptions (
  id, user_id, price_id, status, 
  current_period_start, current_period_end,
  stripe_subscription_id, stripe_customer_id, collection_method
)
VALUES
  (
    '99999999-9999-9999-9999-999999999999',
    '11111111-1111-1111-1111-111111111111',
    '55555555-5555-5555-5555-555555555555',
    'active',
    NOW() - INTERVAL '15 days',
    NOW() + INTERVAL '15 days',
    'sub_test123',
    'cus_test123',
    'charge_automatically'
  );

-- Insert test invoices
INSERT INTO invoices (
  id, user_id, subscription_id, status,
  amount_due, amount_paid, currency,
  stripe_invoice_id, period_start, period_end
)
VALUES
  (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '11111111-1111-1111-1111-111111111111',
    '99999999-9999-9999-9999-999999999999',
    'paid',
    1000,
    1000,
    'usd',
    'in_test123',
    NOW() - INTERVAL '15 days',
    NOW() + INTERVAL '15 days'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Clean test data (in reverse order of dependencies)
DELETE FROM invoices WHERE id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
DELETE FROM subscriptions WHERE id = '99999999-9999-9999-9999-999999999999';
DELETE FROM prices WHERE id IN (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666666',
  '77777777-7777-7777-7777-777777777777',
  '88888888-8888-8888-8888-888888888888'
);
DELETE FROM products WHERE id IN (
  '33333333-3333-3333-3333-333333333333',
  '44444444-4444-4444-4444-444444444444'
);
DELETE FROM users WHERE id IN (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222'
);
-- +goose StatementEnd
