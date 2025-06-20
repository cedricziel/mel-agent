-- Migration 015: Add OpenAPI-compatible columns to workflow_runs table
-- This adds columns expected by the OpenAPI handlers while maintaining backward compatibility

-- Make agent_id and version_id nullable for new OpenAPI workflow runs that use workflow_id instead
ALTER TABLE workflow_runs ALTER COLUMN agent_id DROP NOT NULL;
ALTER TABLE workflow_runs ALTER COLUMN version_id DROP NOT NULL;

-- Add workflow_id column to map to the workflows table created in migration 014
ALTER TABLE workflow_runs 
ADD COLUMN IF NOT EXISTS workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL;

-- Add context column to replace/supplement variables
ALTER TABLE workflow_runs 
ADD COLUMN IF NOT EXISTS context JSONB;

-- Add error column to replace/supplement error_data
ALTER TABLE workflow_runs 
ADD COLUMN IF NOT EXISTS error TEXT;

-- Update existing workflow_runs to have proper context and error fields
-- Copy variables to context for backward compatibility
UPDATE workflow_runs 
SET context = variables 
WHERE context IS NULL AND variables IS NOT NULL;

-- Copy error_data to error field (extract message if it's JSON)
UPDATE workflow_runs 
SET error = CASE 
    WHEN error_data IS NOT NULL AND jsonb_typeof(error_data) = 'object' AND error_data ? 'message' 
    THEN error_data ->> 'message'
    WHEN error_data IS NOT NULL 
    THEN error_data::text
    ELSE NULL
END
WHERE error IS NULL AND error_data IS NOT NULL;

-- Create an index for the new workflow_id column
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id ON workflow_runs(workflow_id);

-- Update column mappings for the OpenAPI handlers
-- The handlers expect: workflow_id, context, error
-- But existing schema has: agent_id, variables, error_data
-- We'll support both for compatibility