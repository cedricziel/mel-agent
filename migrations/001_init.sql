-- Enable extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    name        TEXT,
    plan        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Integrations (generic external systems)
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    category TEXT NOT NULL,
    auth_type TEXT NOT NULL,
    base_url TEXT,
    hosted_available BOOLEAN DEFAULT FALSE,
    hosted_pricing JSONB,
    capabilities TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Connections (user-scoped credential instances)
CREATE TABLE IF NOT EXISTS connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    integration_id UUID REFERENCES integrations(id),
    name TEXT NOT NULL,
    secret JSONB,
    config JSONB,
    usage_limit_month INTEGER,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_validated TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'valid'
);

-- Agents
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    latest_version_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Agent versions
CREATE TABLE IF NOT EXISTS agent_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    semantic_version TEXT NOT NULL,
    graph JSONB NOT NULL,
    default_params JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed a default user for early dev (will be removed once auth lands)
INSERT INTO users (id, email, name, plan)
VALUES ('00000000-0000-0000-0000-000000000001', 'demo@example.com', 'Demo User', 'free')
ON CONFLICT (id) DO NOTHING;
