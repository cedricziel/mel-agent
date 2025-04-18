-- Migration 005: Add agent_id and node_id to triggers table
-- Adds references to agent and the graph node identifier for trigger sync
ALTER TABLE triggers
  ADD COLUMN IF NOT EXISTS agent_id UUID REFERENCES agents(id) ON DELETE CASCADE;
ALTER TABLE triggers
  ADD COLUMN IF NOT EXISTS node_id TEXT;
-- Index for quick lookup by agent and node
CREATE INDEX IF NOT EXISTS idx_triggers_agent_node ON triggers(agent_id, node_id);