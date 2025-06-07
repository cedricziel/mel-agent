-- Add auto-persisting workflow drafts table
-- This enables continuous saving and testing without requiring explicit versions

-- Create workflow_versions table if it doesn't exist
CREATE TABLE IF NOT EXISTS workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    nodes JSONB NOT NULL DEFAULT '[]'::jsonb,
    edges JSONB NOT NULL DEFAULT '[]'::jsonb,
    layout JSONB,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(agent_id, version)
);

-- Create draft workflows table
CREATE TABLE workflow_drafts (
    agent_id UUID PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
    nodes JSONB NOT NULL DEFAULT '[]'::jsonb,
    edges JSONB NOT NULL DEFAULT '[]'::jsonb,
    layout JSONB,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for efficient queries
CREATE INDEX idx_workflow_drafts_updated_at ON workflow_drafts(updated_at);

-- Update trigger to maintain updated_at
CREATE OR REPLACE FUNCTION update_workflow_drafts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_workflow_drafts_updated_at
    BEFORE UPDATE ON workflow_drafts
    FOR EACH ROW
    EXECUTE FUNCTION update_workflow_drafts_updated_at();

-- Modify existing workflow_versions table to add deployment tracking
ALTER TABLE workflow_versions ADD COLUMN IF NOT EXISTS deployed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE workflow_versions ADD COLUMN IF NOT EXISTS is_deployed BOOLEAN DEFAULT FALSE;
ALTER TABLE workflow_versions ADD COLUMN IF NOT EXISTS deployment_notes TEXT;

-- Index for deployed versions
CREATE INDEX IF NOT EXISTS idx_workflow_versions_deployed ON workflow_versions(agent_id, is_deployed, deployed_at);

-- Function to get the latest deployed version
CREATE OR REPLACE FUNCTION get_latest_deployed_version(p_agent_id UUID)
RETURNS TABLE(
    agent_id UUID,
    version INTEGER,
    nodes JSONB,
    edges JSONB,
    layout JSONB,
    deployed_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        wv.agent_id,
        wv.version,
        wv.nodes,
        wv.edges,
        wv.layout,
        wv.deployed_at
    FROM workflow_versions wv
    WHERE wv.agent_id = p_agent_id 
      AND wv.is_deployed = true
    ORDER BY wv.deployed_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to deploy a version
CREATE OR REPLACE FUNCTION deploy_workflow_version(
    p_agent_id UUID,
    p_version INTEGER,
    p_deployment_notes TEXT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Mark all other versions as not deployed
    UPDATE workflow_versions 
    SET is_deployed = false 
    WHERE agent_id = p_agent_id;
    
    -- Deploy the specified version
    UPDATE workflow_versions 
    SET 
        is_deployed = true,
        deployed_at = NOW(),
        deployment_notes = p_deployment_notes
    WHERE agent_id = p_agent_id AND version = p_version;
    
    -- Return success if a row was updated
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;