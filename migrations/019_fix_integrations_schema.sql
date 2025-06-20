-- Migration 019: Fix integrations table schema to match API expectations
-- Add missing columns that the API handler expects

-- Add description column
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS description TEXT;

-- Add status column with default value
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active';

-- Add updated_at column
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Create trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_integrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_integrations_updated_at
    BEFORE UPDATE ON integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_integrations_updated_at();

-- Update existing rows to have sensible defaults
UPDATE integrations SET 
    description = 'Integration for ' || name,
    status = 'active',
    updated_at = created_at
WHERE description IS NULL;

-- Also need to handle the type vs category mismatch
-- The API expects "type" but the table has "category"
-- Let's add type as an alias for category
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS type TEXT;
UPDATE integrations SET type = category WHERE type IS NULL;