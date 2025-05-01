-- +goose Up
-- +goose StatementBegin
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    price_id UUID REFERENCES prices(id),
    subscription_id UUID REFERENCES subscriptions(id),
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    is_subscription BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    options JSONB,
    metadata JSONB
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_price_id ON order_items(price_id);
CREATE INDEX idx_order_items_subscription_id ON order_items(subscription_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE order_items;
-- +goose StatementEnd
