-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, password_hash, name, stripe_customer_id)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Test User', 'cus_test123'),
  ('22222222-2222-2222-2222-222222222222', 'admin@walkingdrum.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Admin User', 'cus_admin456');

-- Insert coffee products
INSERT INTO products (id, name, description, is_active, stripe_product_id)
VALUES
  ('33333333-3333-3333-3333-333333333333', 'Ethiopian Yirgacheffe', 'Light roast with floral and citrus notes', true, 'prod_eth001'),
  ('44444444-4444-4444-4444-444444444444', 'Colombian Supremo', 'Medium roast with caramel and nutty flavors', true, 'prod_col002'),
  ('55555555-5555-5555-5555-555555555555', 'Sumatra Dark', 'Dark roast with earthy and spicy tones', true, 'prod_sum003'),
  ('66666666-6666-6666-6666-666666666666', 'House Blend', 'Balanced medium roast, perfect for everyday drinking', true, 'prod_house004');

-- Insert product prices
INSERT INTO prices (id, product_id, amount, currency, interval_type, interval_count, is_active, stripe_price_id, nickname)
VALUES
  -- One-time purchase prices
  ('77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', 1499, 'usd', 'month', 0, true, 'price_eth_onetime', 'Ethiopian Yirgacheffe - One Time'),
  ('88888888-8888-8888-8888-888888888888', '44444444-4444-4444-4444-444444444444', 1399, 'usd', 'month', 0, true, 'price_col_onetime', 'Colombian Supremo - One Time'),
  ('99999999-9999-9999-9999-999999999999', '55555555-5555-5555-5555-555555555555', 1599, 'usd', 'month', 0, true, 'price_sum_onetime', 'Sumatra Dark - One Time'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '66666666-6666-6666-6666-666666666666', 1299, 'usd', 'month', 0, true, 'price_house_onetime', 'House Blend - One Time'),
  
  -- Subscription prices
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', 1199, 'usd', 'month', 1, true, 'price_eth_monthly', 'Ethiopian Yirgacheffe - Monthly'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444444', 1099, 'usd', 'month', 1, true, 'price_col_monthly', 'Colombian Supremo - Monthly'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', '55555555-5555-5555-5555-555555555555', 1299, 'usd', 'month', 1, true, 'price_sum_monthly', 'Sumatra Dark - Monthly'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '66666666-6666-6666-6666-666666666666', 999, 'usd', 'month', 1, true, 'price_house_monthly', 'House Blend - Monthly');

-- Insert active cart for Test User
INSERT INTO carts (id, user_id, created_at, updated_at, expires_at)
VALUES
  ('f47ac10b-58cc-4372-a567-0e02b2c3d479', '11111111-1111-1111-1111-111111111111', NOW(), NOW(), NOW() + INTERVAL '7 days');

-- Insert cart items
INSERT INTO cart_items (id, cart_id, product_id, price_id, quantity, unit_price, is_subscription, options)
VALUES
  ('550e8400-e29b-41d4-a716-446655440000', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', '33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 1, 1199, true, '{"grind_type": "medium", "size": "12oz"}'),
  ('6ba7b810-9dad-11d1-80b4-00c04fd430c8', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', '44444444-4444-4444-4444-444444444444', '88888888-8888-8888-8888-888888888888', 2, 1399, false, '{"grind_type": "whole_bean", "size": "1lb"}');

-- Insert active session cart
INSERT INTO carts (id, session_id, created_at, updated_at, expires_at)
VALUES
  ('6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b', 'test_session_12345', NOW(), NOW(), NOW() + INTERVAL '7 days');

-- Insert session cart items
INSERT INTO cart_items (id, cart_id, product_id, price_id, quantity, unit_price, is_subscription, options)
VALUES
  ('38400000-8cf0-11bd-b23e-10b96e4ef00d', '6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b', '55555555-5555-5555-5555-555555555555', '99999999-9999-9999-9999-999999999999', 1, 1599, false, '{"grind_type": "coarse", "size": "1lb"}');

-- Insert completed order for Test User
INSERT INTO orders (
  id, user_id, status, total_amount, currency, 
  created_at, updated_at, completed_at, 
  shipping_address, billing_address, 
  payment_intent_id, stripe_customer_id
)
VALUES
  (
    '7c9e6679-7425-40de-944b-e07fc1f90ae7', 
    '11111111-1111-1111-1111-111111111111',
    'completed',
    3097,
    'usd',
    NOW() - INTERVAL '15 days',
    NOW() - INTERVAL '14 days',
    NOW() - INTERVAL '14 days',
    '{"name": "Test User", "line1": "123 Main St", "city": "Anytown", "state": "CA", "postal_code": "12345", "country": "US"}',
    '{"name": "Test User", "line1": "123 Main St", "city": "Anytown", "state": "CA", "postal_code": "12345", "country": "US"}',
    'pi_completed12345',
    'cus_test123'
  );

-- Insert subscription created from the order
INSERT INTO subscriptions (
  id, user_id, price_id, quantity, status, 
  current_period_start, current_period_end,
  stripe_subscription_id, stripe_customer_id, collection_method
)
VALUES
  (
    '5fc0a1f2-dfb2-4d57-a4fc-1b7d868d8155',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    1,
    'active',
    NOW() - INTERVAL '15 days',
    NOW() + INTERVAL '15 days',
    'sub_eth789',
    'cus_test123',
    'charge_automatically'
  );

-- Insert order items for the completed order
INSERT INTO order_items (
  id, order_id, product_id, price_id, subscription_id,
  quantity, unit_price, total_price, is_subscription,
  created_at, updated_at, options
)
VALUES
  (
    '494c44a8-8e7c-4e79-a292-9cdd59e5e2b8',
    '7c9e6679-7425-40de-944b-e07fc1f90ae7',
    '33333333-3333-3333-3333-333333333333',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    '5fc0a1f2-dfb2-4d57-a4fc-1b7d868d8155',
    1,
    1199,
    1199,
    true,
    NOW() - INTERVAL '15 days',
    NOW() - INTERVAL '15 days',
    '{"grind_type": "medium", "size": "12oz"}'
  ),
  (
    '8c3d4b45-070c-4ede-8b1c-76e3351fa00a',
    '7c9e6679-7425-40de-944b-e07fc1f90ae7',
    '44444444-4444-4444-4444-444444444444',
    '88888888-8888-8888-8888-888888888888',
    NULL,
    2,
    949,
    1898,
    false,
    NOW() - INTERVAL '15 days',
    NOW() - INTERVAL '15 days',
    '{"grind_type": "fine", "size": "12oz"}'
  );

-- Insert pending order for Test User
INSERT INTO orders (
  id, user_id, status, total_amount, currency, 
  created_at, updated_at,
  shipping_address, billing_address, 
  payment_intent_id, stripe_customer_id
)
VALUES
  (
    '9ed66b58-f3c0-4dd7-8d5d-9c28e87b87a1',
    '11111111-1111-1111-1111-111111111111',
    'pending',
    1599,
    'usd',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days',
    '{"name": "Test User", "line1": "123 Main St", "city": "Anytown", "state": "CA", "postal_code": "12345", "country": "US"}',
    '{"name": "Test User", "line1": "123 Main St", "city": "Anytown", "state": "CA", "postal_code": "12345", "country": "US"}',
    'pi_pending67890',
    'cus_test123'
  );

-- Insert order items for the pending order
INSERT INTO order_items (
  id, order_id, product_id, price_id,
  quantity, unit_price, total_price, is_subscription,
  created_at, updated_at, options
)
VALUES
  (
    'a7183f9d-5d5a-4d5b-8209-769a1d3a691b',
    '9ed66b58-f3c0-4dd7-8d5d-9c28e87b87a1',
    '55555555-5555-5555-5555-555555555555',
    '99999999-9999-9999-9999-999999999999',
    1,
    1599,
    1599,
    false,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days',
    '{"grind_type": "coarse", "size": "1lb"}'
  );

-- Insert invoice for subscription
INSERT INTO invoices (
  id, user_id, subscription_id, status,
  amount_due, amount_paid, currency,
  stripe_invoice_id, period_start, period_end
)
VALUES
  (
    'b61f16be-c9df-4b5a-97da-2df98410bfe3',
    '11111111-1111-1111-1111-111111111111',
    '5fc0a1f2-dfb2-4d57-a4fc-1b7d868d8155',
    'paid',
    1199,
    1199,
    'usd',
    'in_sub_eth789',
    NOW() - INTERVAL '15 days',
    NOW() + INTERVAL '15 days'
  );

-- Insert upcoming invoice for subscription
INSERT INTO invoices (
  id, user_id, subscription_id, status,
  amount_due, amount_paid, currency,
  stripe_invoice_id, period_start, period_end
)
VALUES
  (
    'c5b5eaee-a20d-4c5c-8979-1a71c6679a5d',
    '11111111-1111-1111-1111-111111111111',
    '5fc0a1f2-dfb2-4d57-a4fc-1b7d868d8155',
    'draft',
    1199,
    0,
    'usd',
    'in_upcoming_eth789',
    NOW() + INTERVAL '15 days',
    NOW() + INTERVAL '45 days'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM invoices WHERE id IN (
  'b61f16be-c9df-4b5a-97da-2df98410bfe3',
  'c5b5eaee-a20d-4c5c-8979-1a71c6679a5d'
);

DELETE FROM order_items WHERE id IN (
  '494c44a8-8e7c-4e79-a292-9cdd59e5e2b8',
  '8c3d4b45-070c-4ede-8b1c-76e3351fa00a',
  'a7183f9d-5d5a-4d5b-8209-769a1d3a691b'
);

DELETE FROM orders WHERE id IN (
  '7c9e6679-7425-40de-944b-e07fc1f90ae7',
  '9ed66b58-f3c0-4dd7-8d5d-9c28e87b87a1'
);

DELETE FROM subscriptions WHERE id = '5fc0a1f2-dfb2-4d57-a4fc-1b7d868d8155';

DELETE FROM cart_items WHERE id IN (
  '550e8400-e29b-41d4-a716-446655440000',
  '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
  '38400000-8cf0-11bd-b23e-10b96e4ef00d'
);

DELETE FROM carts WHERE id IN (
  'f47ac10b-58cc-4372-a567-0e02b2c3d479',
  '6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b'
);

DELETE FROM prices WHERE id IN (
  '77777777-7777-7777-7777-777777777777',
  '88888888-8888-8888-8888-888888888888',
  '99999999-9999-9999-9999-999999999999',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'
);

DELETE FROM products WHERE id IN (
  '33333333-3333-3333-3333-333333333333',
  '44444444-4444-4444-4444-444444444444',
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666666'
);

DELETE FROM users WHERE id IN (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222'
);
-- +goose StatementEnd
