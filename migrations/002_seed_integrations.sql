INSERT INTO integrations (name, category, auth_type, capabilities, hosted_available)
VALUES
  ('openai', 'llm_provider', 'api_key', ARRAY['chat','embeddings']::text[], TRUE),
  ('gmail', 'communication', 'oauth2', ARRAY['send_email']::text[], FALSE),
  ('airtable', 'database', 'api_key', ARRAY['read_rows','write_rows']::text[], FALSE)
ON CONFLICT (name) DO NOTHING;
