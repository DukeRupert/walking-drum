-- +goose Up
-- +goose StatementBegin
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    object_id VARCHAR(255),
    object_type VARCHAR(255),
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    is_processed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_webhook_events_stripe_event_id ON webhook_events(stripe_event_id);
CREATE INDEX idx_webhook_events_event_type ON webhook_events(event_type);
CREATE INDEX idx_webhook_events_object_id ON webhook_events(object_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE webhook_events;
-- +goose StatementEnd
