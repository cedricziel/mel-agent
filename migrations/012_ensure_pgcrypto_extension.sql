-- Ensure pgcrypto extension is available for UUID generation
-- This migration ensures that fresh databases have the pgcrypto extension
-- which is required for gen_random_uuid() function used throughout the schema

CREATE EXTENSION IF NOT EXISTS pgcrypto;