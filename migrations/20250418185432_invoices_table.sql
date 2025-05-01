-- +goose Up
-- +goose StatementBegin
CREATE TYPE invoice_status AS ENUM ('draft', 'open', 'paid', 'uncollectible', 'void');

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id),
    status invoice_status NOT NULL,
    amount_due BIGINT NOT NULL,
    amount_paid BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    invoice_pdf VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    stripe_invoice_id VARCHAR(255) UNIQUE,
    payment_intent_id VARCHAR(255),
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_stripe_invoice_id ON invoices(stripe_invoice_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE invoices;
DROP TYPE invoice_status;
-- +goose StatementEnd
