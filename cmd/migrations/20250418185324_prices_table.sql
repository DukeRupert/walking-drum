-- +goose Up
-- +goose StatementBegin
CREATE TYPE billing_interval AS ENUM ('day', 'week', 'month', 'year');

CREATE TABLE prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    interval_type billing_interval NOT NULL,
    interval_count INTEGER NOT NULL DEFAULT 1,
    trial_period_days INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    stripe_price_id VARCHAR(255) UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    nickname VARCHAR(255),
    metadata JSONB
);

CREATE INDEX idx_prices_product_id ON prices(product_id);
CREATE INDEX idx_prices_stripe_price_id ON prices(stripe_price_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE prices;
DROP TYPE billing_interval;
-- +goose StatementEnd
