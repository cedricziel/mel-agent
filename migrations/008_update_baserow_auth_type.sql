-- Update Baserow integration to support custom authentication (JWT with username/password)
UPDATE integrations 
SET auth_type = 'custom' 
WHERE name = 'baserow';