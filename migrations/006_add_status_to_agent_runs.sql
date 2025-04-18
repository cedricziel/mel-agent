-- Add status column to agent_runs for runner
ALTER TABLE agent_runs
    ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';