-- Create messages table with partitioning
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE messages (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id) 
) PARTITION BY LIST (tenant_id);

-- Create index on tenant_id
CREATE INDEX idx_messages_tenant_id ON messages (tenant_id);
CREATE INDEX idx_messages_created_at ON messages (created_at);