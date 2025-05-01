-- +goose Up
-- +goose StatementBegin
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    price_id UUID REFERENCES prices(id),
    quantity INT NOT NULL DEFAULT 1,
    unit_price BIGINT NOT NULL,
    is_subscription BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    options JSONB, -- For coffee-specific options like grind type, roast level, etc.
    metadata JSONB
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);
CREATE INDEX idx_cart_items_price_id ON cart_items(price_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cart_items;
-- +goose StatementEnd
