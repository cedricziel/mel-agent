-- Migration 018: Add workflow versions and drafts functionality
-- Drop existing agent-based tables and create new workflow-based ones

-- Drop existing agent-based tables
DROP TABLE IF EXISTS workflow_versions CASCADE;
DROP TABLE IF EXISTS workflow_drafts CASCADE;

-- Create new workflow_versions table for workflows (not agents)
CREATE TABLE workflow_versions (
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

-- Create new workflow_drafts table for workflows (not agents)
CREATE TABLE workflow_drafts (
    workflow_id UUID PRIMARY KEY REFERENCES workflows(id) ON DELETE CASCADE,
    definition JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_workflow_versions_workflow_id ON workflow_versions(workflow_id);
CREATE INDEX idx_workflow_versions_current ON workflow_versions(workflow_id, is_current);
CREATE INDEX idx_workflow_versions_created_at ON workflow_versions(created_at);
CREATE INDEX idx_workflow_drafts_updated_at ON workflow_drafts(updated_at);