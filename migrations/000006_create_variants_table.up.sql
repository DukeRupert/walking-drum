-- Up migration: Create variants table
CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price_id UUID NOT NULL REFERENCES prices(id),
    stripe_price_id VARCHAR(255) NOT NULL,
    weight VARCHAR(50) NOT NULL,  -- "12oz", "3lb", "5lb"
    grind VARCHAR(50) NOT NULL,   -- "Whole Bean", "Drip Ground"
    active BOOLEAN DEFAULT TRUE,
    stock_level INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create a unique constraint to prevent duplicate variants
CREATE UNIQUE INDEX idx_variants_product_weight_grind ON variants(product_id, weight, grind);

-- Create indexes for common query patterns
CREATE INDEX idx_variants_product_id ON variants(product_id);
CREATE INDEX idx_variants_price_id ON variants(price_id);
CREATE INDEX idx_variants_stripe_price_id ON variants(stripe_price_id);
CREATE INDEX idx_variants_active ON variants(active);
