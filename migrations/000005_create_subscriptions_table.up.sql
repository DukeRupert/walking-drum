-- Create subscriptions table with Stripe alignment
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    product_id UUID NOT NULL REFERENCES products(id),
    price_id UUID NOT NULL REFERENCES prices(id),
    
    -- Stripe-specific fields
    stripe_id VARCHAR(255) NOT NULL,
    stripe_item_id VARCHAR(255),
    
    -- Quantity and status
    quantity INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'past_due', 'incomplete', 'incomplete_expired', 'trialing', 'canceled', 'unpaid', 'paused')),
    
    -- Billing period
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    next_delivery_date TIMESTAMP WITH TIME ZONE,
    
    -- Cancellation details
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata and timestamps
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX idx_subscriptions_stripe_id ON subscriptions(stripe_id);
CREATE INDEX idx_subscriptions_stripe_item_id ON subscriptions(stripe_item_id);
CREATE INDEX idx_subscriptions_customer_id ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_current_period_end ON subscriptions(current_period_end);
CREATE INDEX idx_subscriptions_next_delivery_date ON subscriptions(next_delivery_date);