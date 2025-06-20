-- Migration 018: Add workflow versions and drafts functionality
-- This replaces the missing agent_drafts and agent_versions tables 
-- with proper workflow-based versions and drafts

-- Create workflow_versions table for versioning support
CREATE TABLE IF NOT EXISTS workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    is_current BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workflow_id, version_number)
);

-- Create workflow_drafts table for auto-persistence
CREATE TABLE IF NOT EXISTS workflow_drafts (
    workflow_id UUID PRIMARY KEY REFERENCES workflows(id) ON DELETE CASCADE,
    definition JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_workflow_versions_workflow_id ON workflow_versions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_versions_current ON workflow_versions(workflow_id, is_current);
CREATE INDEX IF NOT EXISTS idx_workflow_versions_created_at ON workflow_versions(created_at);
CREATE INDEX IF NOT EXISTS idx_workflow_drafts_updated_at ON workflow_drafts(updated_at);

-- Update trigger for workflow_drafts updated_at
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

-- Function to get the current version of a workflow
CREATE OR REPLACE FUNCTION get_current_workflow_version(p_workflow_id UUID)
RETURNS TABLE(
    id UUID,
    workflow_id UUID,
    version_number INTEGER,
    name TEXT,
    description TEXT,
    definition JSONB,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        wv.id,
        wv.workflow_id,
        wv.version_number,
        wv.name,
        wv.description,
        wv.definition,
        wv.created_at
    FROM workflow_versions wv
    WHERE wv.workflow_id = p_workflow_id 
      AND wv.is_current = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to create a new workflow version
CREATE OR REPLACE FUNCTION create_workflow_version(
    p_workflow_id UUID,
    p_name TEXT,
    p_description TEXT DEFAULT NULL,
    p_definition JSONB DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_version_number INTEGER;
    v_version_id UUID;
    v_definition JSONB;
BEGIN
    -- Get the next version number
    SELECT COALESCE(MAX(version_number), 0) + 1 
    INTO v_version_number
    FROM workflow_versions 
    WHERE workflow_id = p_workflow_id;
    
    -- Use provided definition or get from draft
    IF p_definition IS NOT NULL THEN
        v_definition := p_definition;
    ELSE
        SELECT definition INTO v_definition
        FROM workflow_drafts 
        WHERE workflow_id = p_workflow_id;
        
        -- If no draft exists, use empty workflow
        IF v_definition IS NULL THEN
            v_definition := '{"nodes": [], "edges": []}';
        END IF;
    END IF;
    
    -- Mark all other versions as not current
    UPDATE workflow_versions 
    SET is_current = false 
    WHERE workflow_id = p_workflow_id;
    
    -- Create new version
    INSERT INTO workflow_versions (
        workflow_id, 
        version_number, 
        name, 
        description, 
        definition, 
        is_current
    ) VALUES (
        p_workflow_id, 
        v_version_number, 
        p_name, 
        p_description, 
        v_definition, 
        true
    ) RETURNING id INTO v_version_id;
    
    RETURN v_version_id;
END;
$$ LANGUAGE plpgsql;

-- Function to deploy a specific version (make it current)
CREATE OR REPLACE FUNCTION deploy_workflow_version(
    p_workflow_id UUID,
    p_version_number INTEGER
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Mark all versions as not current
    UPDATE workflow_versions 
    SET is_current = false 
    WHERE workflow_id = p_workflow_id;
    
    -- Mark the specified version as current
    UPDATE workflow_versions 
    SET is_current = true 
    WHERE workflow_id = p_workflow_id 
      AND version_number = p_version_number;
    
    -- Return success if a row was updated
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;