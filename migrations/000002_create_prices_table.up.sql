CREATE TABLE prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_id VARCHAR(255) UNIQUE NOT NULL,
    product_id UUID REFERENCES products(id),
    nickname VARCHAR(255),
    unit_amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    recurring_interval VARCHAR(10) NOT NULL,
    recurring_interval_count INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);