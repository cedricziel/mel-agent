-- Ensure pgcrypto extension is available for UUID generation
-- This migration serves as a safety net to ensure the pgcrypto extension
-- is available, even in scenarios where the initial migration might have been
-- skipped or in partial database setups. Required for gen_random_uuid() function.

CREATE EXTENSION IF NOT EXISTS pgcrypto;