-- Migration 013: Update triggers table to match OpenAPI spec
-- Adds name and type columns and aligns with OpenAPI handler expectations

-- Add name and type columns for OpenAPI compatibility
ALTER TABLE triggers 
  ADD COLUMN IF NOT EXISTS name TEXT,
  ADD COLUMN IF NOT EXISTS type TEXT;

-- Update existing rows to have reasonable defaults if they exist
UPDATE triggers 
SET 
  name = COALESCE(name, 'Legacy Trigger'),
  type = COALESCE(type, 'schedule')
WHERE name IS NULL OR type IS NULL;

-- Make name and type NOT NULL after setting defaults
ALTER TABLE triggers 
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN type SET NOT NULL;

-- Add constraint to ensure type is valid
ALTER TABLE triggers
  ADD CONSTRAINT check_trigger_type CHECK (type IN ('schedule', 'webhook'));

-- Update the index for better performance
CREATE INDEX IF NOT EXISTS idx_triggers_type ON triggers(type);
CREATE INDEX IF NOT EXISTS idx_triggers_name ON triggers(name);