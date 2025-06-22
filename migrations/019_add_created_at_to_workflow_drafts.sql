-- Migration 019: Add created_at column to workflow_drafts table
-- This ensures consistency with the handler expectations

ALTER TABLE workflow_drafts 
ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Update existing records to have a created_at value
UPDATE workflow_drafts 
SET created_at = updated_at 
WHERE created_at IS NULL;