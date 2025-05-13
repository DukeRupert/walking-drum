-- Add UUID extension if not already installed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    origin VARCHAR(255),
    roast_level VARCHAR(50),
    -- Remove the fixed grind column since it will now be in options
    -- grind VARCHAR(50),
    stock_level INTEGER DEFAULT 0,
    -- Remove the fixed weight column since it will now be in options
    -- weight INTEGER,
    flavor_notes TEXT,
    active BOOLEAN DEFAULT TRUE,
    -- Add options column to store product options as JSONB
    options JSONB DEFAULT '{}',
    -- Add a flag for subscription capability
    allow_subscription BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);