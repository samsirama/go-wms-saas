-- WMS SaaS Initial Schema

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_sku ON products(sku);

-- Stocks with version-based optimistic locking
CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id)
);

CREATE INDEX idx_stocks_product_id ON stocks(product_id);

-- Audit trail for all stock changes
CREATE TABLE IF NOT EXISTS stock_mutations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    mutation_type VARCHAR(50) NOT NULL,
    quantity BIGINT NOT NULL,
    previous_qty BIGINT NOT NULL,
    new_qty BIGINT NOT NULL,
    reference_id VARCHAR(255),
    reference_type VARCHAR(50),
    notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stock_mutations_product_id ON stock_mutations(product_id);
CREATE INDEX idx_stock_mutations_reference ON stock_mutations(reference_id, reference_type);
CREATE INDEX idx_stock_mutations_created_at ON stock_mutations(created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stocks_updated_at BEFORE UPDATE ON stocks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data
INSERT INTO products (sku, name, description) VALUES
    ('LAPTOP-001', 'MacBook Pro 16"', 'Apple MacBook Pro 16-inch M2'),
    ('MOUSE-001', 'Logitech MX Master 3', 'Wireless ergonomic mouse'),
    ('KEYBOARD-001', 'Keychron K2', 'Mechanical keyboard')
ON CONFLICT (sku) DO NOTHING;

INSERT INTO stocks (product_id, quantity)
SELECT id, 100 FROM products WHERE sku IN ('LAPTOP-001', 'MOUSE-001', 'KEYBOARD-001')
ON CONFLICT (product_id) DO NOTHING;
