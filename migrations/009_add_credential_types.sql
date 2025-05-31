-- Add credential_type to integrations to specify which credential definition to use
ALTER TABLE integrations ADD COLUMN credential_type TEXT;

-- Update existing integrations to use appropriate credential types
UPDATE integrations SET credential_type = 'api_key' WHERE auth_type IN ('api_key', 'token');
UPDATE integrations SET credential_type = 'baserow_jwt' WHERE name = 'baserow' AND auth_type = 'custom';

-- Add credential_type to connections to track which credential definition was used
ALTER TABLE connections ADD COLUMN credential_type TEXT;