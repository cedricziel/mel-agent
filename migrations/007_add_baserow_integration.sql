INSERT INTO integrations (name, category, auth_type, capabilities, hosted_available)
VALUES
  ('baserow', 'database', 'token', ARRAY['read_rows','write_rows','list_databases','list_tables']::text[], FALSE)
ON CONFLICT (name) DO NOTHING;