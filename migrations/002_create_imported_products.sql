-- Imported Products Table
-- Stores products imported from marketplaces that can be mapped to internal products

CREATE TABLE IF NOT EXISTS marketplace.imported_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    external_product_id VARCHAR(100) NOT NULL,
    external_sku VARCHAR(100),
    name VARCHAR(500) NOT NULL,
    description TEXT,
    price DECIMAL(12, 2),
    stock INTEGER DEFAULT 0,
    category_id VARCHAR(100),
    status VARCHAR(50), -- NORMAL, BANNED, DELETED, UNLIST
    image_url TEXT,
    is_mapped BOOLEAN DEFAULT false,
    mapped_to_product_id UUID, -- Internal product ID if mapped
    imported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_connection_imported_product UNIQUE (connection_id, external_product_id)
);

CREATE INDEX idx_imported_products_connection ON marketplace.imported_products(connection_id);
CREATE INDEX idx_imported_products_is_mapped ON marketplace.imported_products(is_mapped);
CREATE INDEX idx_imported_products_external_id ON marketplace.imported_products(external_product_id);

-- Apply update trigger
CREATE TRIGGER update_imported_products_updated_at
    BEFORE UPDATE ON marketplace.imported_products
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

COMMENT ON TABLE marketplace.imported_products IS 'Products imported from marketplaces awaiting mapping to internal products';
