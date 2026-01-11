-- Marketplace Integration Schema
-- Run this migration to create all marketplace-related tables

-- Create marketplace schema
CREATE SCHEMA IF NOT EXISTS marketplace;

-- =====================================================
-- MARKETPLACE CONNECTIONS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(50) NOT NULL, -- 'shopee', 'tiktok'
    shop_id VARCHAR(100) NOT NULL,
    shop_name VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL, -- Encrypted
    refresh_token TEXT, -- Encrypted
    token_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_platform_shop UNIQUE (platform, shop_id)
);

CREATE INDEX idx_connections_platform ON marketplace.connections(platform);
CREATE INDEX idx_connections_is_active ON marketplace.connections(is_active);

-- =====================================================
-- PRODUCT MAPPINGS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.product_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    internal_product_id UUID NOT NULL,
    external_product_id VARCHAR(100) NOT NULL,
    external_sku VARCHAR(100),
    sync_status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'synced', 'error'
    sync_error TEXT,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_connection_internal_product UNIQUE (connection_id, internal_product_id),
    CONSTRAINT unique_connection_external_product UNIQUE (connection_id, external_product_id)
);

CREATE INDEX idx_product_mappings_connection ON marketplace.product_mappings(connection_id);
CREATE INDEX idx_product_mappings_internal_product ON marketplace.product_mappings(internal_product_id);
CREATE INDEX idx_product_mappings_sync_status ON marketplace.product_mappings(sync_status);

-- =====================================================
-- VARIANT MAPPINGS TABLE (for products with variants)
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.variant_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_mapping_id UUID NOT NULL REFERENCES marketplace.product_mappings(id) ON DELETE CASCADE,
    internal_variant_id UUID NOT NULL,
    external_variant_id VARCHAR(100) NOT NULL,
    external_sku VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_product_mapping_variant UNIQUE (product_mapping_id, internal_variant_id)
);

CREATE INDEX idx_variant_mappings_product ON marketplace.variant_mappings(product_mapping_id);
CREATE INDEX idx_variant_mappings_internal_variant ON marketplace.variant_mappings(internal_variant_id);

-- =====================================================
-- CATEGORY MAPPINGS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.category_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    internal_category_id UUID NOT NULL,
    internal_category_name VARCHAR(255),
    external_category_id VARCHAR(100) NOT NULL,
    external_category_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_connection_internal_category UNIQUE (connection_id, internal_category_id)
);

CREATE INDEX idx_category_mappings_connection ON marketplace.category_mappings(connection_id);

-- =====================================================
-- SYNC JOBS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL, -- 'product_push', 'inventory_sync', 'order_import'
    payload JSONB DEFAULT '{}', -- Job payload data
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    total_items INTEGER DEFAULT 0,
    processed_items INTEGER DEFAULT 0,
    failed_items INTEGER DEFAULT 0,
    error_message TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sync_jobs_connection ON marketplace.sync_jobs(connection_id);
CREATE INDEX idx_sync_jobs_status ON marketplace.sync_jobs(status);
CREATE INDEX idx_sync_jobs_created_at ON marketplace.sync_jobs(created_at DESC);

-- =====================================================
-- MARKETPLACE ORDERS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    external_order_id VARCHAR(100) NOT NULL,
    internal_order_id UUID, -- Link to main order system after import
    platform VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'MYR',
    order_data JSONB DEFAULT '{}', -- Raw order data from marketplace
    buyer_info JSONB DEFAULT '{}',
    shipping_info JSONB DEFAULT '{}',
    synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_connection_external_order UNIQUE (connection_id, external_order_id)
);

CREATE INDEX idx_orders_connection ON marketplace.orders(connection_id);
CREATE INDEX idx_orders_external_order_id ON marketplace.orders(external_order_id);
CREATE INDEX idx_orders_status ON marketplace.orders(status);
CREATE INDEX idx_orders_created_at ON marketplace.orders(created_at DESC);

-- =====================================================
-- INVENTORY SYNC LOG TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.inventory_sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES marketplace.connections(id) ON DELETE CASCADE,
    product_mapping_id UUID REFERENCES marketplace.product_mappings(id) ON DELETE SET NULL,
    internal_product_id UUID NOT NULL,
    previous_quantity INTEGER,
    new_quantity INTEGER NOT NULL,
    sync_status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_inventory_logs_connection ON marketplace.inventory_sync_logs(connection_id);
CREATE INDEX idx_inventory_logs_created_at ON marketplace.inventory_sync_logs(created_at DESC);

-- =====================================================
-- WEBHOOK EVENTS TABLE (for debugging/audit)
-- =====================================================
CREATE TABLE IF NOT EXISTS marketplace.webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_platform ON marketplace.webhook_events(platform);
CREATE INDEX idx_webhook_events_processed ON marketplace.webhook_events(processed);
CREATE INDEX idx_webhook_events_created_at ON marketplace.webhook_events(created_at DESC);

-- =====================================================
-- UPDATE TRIGGER FUNCTION
-- =====================================================
CREATE OR REPLACE FUNCTION marketplace.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers
CREATE TRIGGER update_connections_updated_at
    BEFORE UPDATE ON marketplace.connections
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

CREATE TRIGGER update_product_mappings_updated_at
    BEFORE UPDATE ON marketplace.product_mappings
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

CREATE TRIGGER update_category_mappings_updated_at
    BEFORE UPDATE ON marketplace.category_mappings
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

CREATE TRIGGER update_sync_jobs_updated_at
    BEFORE UPDATE ON marketplace.sync_jobs
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON marketplace.orders
    FOR EACH ROW EXECUTE FUNCTION marketplace.update_updated_at_column();

-- =====================================================
-- COMMENTS FOR DOCUMENTATION
-- =====================================================
COMMENT ON TABLE marketplace.connections IS 'Stores marketplace OAuth connections';
COMMENT ON TABLE marketplace.product_mappings IS 'Maps internal products to marketplace products';
COMMENT ON TABLE marketplace.variant_mappings IS 'Maps internal product variants to marketplace variants';
COMMENT ON TABLE marketplace.category_mappings IS 'Maps internal categories to marketplace categories';
COMMENT ON TABLE marketplace.sync_jobs IS 'Background sync job queue';
COMMENT ON TABLE marketplace.orders IS 'Orders imported from marketplaces';
COMMENT ON TABLE marketplace.inventory_sync_logs IS 'Audit log for inventory syncs';
COMMENT ON TABLE marketplace.webhook_events IS 'Webhook events for debugging';
