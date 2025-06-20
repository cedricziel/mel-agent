-- Migration 017: Add workflow_id to triggers table for OpenAPI compatibility
-- This adds workflow_id column to support both legacy agent_id and new workflow_id

-- Add workflow_id column to triggers table
ALTER TABLE triggers 
ADD COLUMN IF NOT EXISTS workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_triggers_workflow_id ON triggers(workflow_id);

-- For OpenAPI compatibility, allow both agent_id and workflow_id to be nullable
-- since triggers can now reference either agents (legacy) or workflows (new)
ALTER TABLE triggers ALTER COLUMN agent_id DROP NOT NULL;