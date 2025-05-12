CREATE TABLE prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_id VARCHAR(255) UNIQUE NOT NULL,
    product_id UUID REFERENCES products(id),
    name VARCHAR(255),
    amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    type VARCHAR(10) NOT NULL CHECK (type IN ('one_time', 'recurring')),
    interval VARCHAR(10) NULL CHECK (interval IN ('week', 'month', 'year')),
    interval_count INTEGER NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Add constraint to ensure interval and interval_count are set when type is recurring
    CONSTRAINT check_recurring_fields CHECK (
        (type = 'one_time' AND interval IS NULL AND interval_count IS NULL) OR
        (type = 'recurring' AND interval IS NOT NULL AND interval_count IS NOT NULL)
    )
);