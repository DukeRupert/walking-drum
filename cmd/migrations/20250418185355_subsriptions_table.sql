-- +goose Up
-- +goose StatementBegin
CREATE TYPE subscription_status AS ENUM ('active', 'past_due', 'canceled', 'unpaid', 'trialing', 'incomplete', 'incomplete_expired');

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    price_id UUID NOT NULL REFERENCES prices(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    status subscription_status NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cancel_at TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    trial_start TIMESTAMP WITH TIME ZONE,
    trial_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    collection_method VARCHAR(50) NOT NULL DEFAULT 'charge_automatically',
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_price_id ON subscriptions(price_id);
CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE subscriptions;
DROP TYPE subscription_status;
-- +goose StatementEnd
