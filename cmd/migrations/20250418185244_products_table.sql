-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    stripe_product_id VARCHAR(255) UNIQUE,
    metadata JSONB
);

CREATE INDEX idx_products_stripe_product_id ON products(stripe_product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE products;
-- +goose StatementEnd
