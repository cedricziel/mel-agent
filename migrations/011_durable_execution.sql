-- Durable Workflow Execution Schema
-- This migration adds support for step-by-step workflow execution persistence

-- Drop the old agent_runs table since we're fully replacing it
DROP TABLE IF EXISTS agent_runs CASCADE;

-- Enhanced workflow runs table
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    version_id UUID NOT NULL,
    trigger_id UUID REFERENCES triggers(id) ON DELETE SET NULL,
    
    -- Execution metadata
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Input/output data
    input_data JSONB,
    output_data JSONB,
    error_data JSONB,
    
    -- Execution context
    variables JSONB DEFAULT '{}',
    timeout_seconds INTEGER DEFAULT 3600,
    retry_policy JSONB DEFAULT '{"max_attempts": 3, "backoff_multiplier": 2}',
    
    -- Worker tracking
    assigned_worker_id TEXT,
    worker_heartbeat TIMESTAMP WITH TIME ZONE,
    
    -- Performance metrics
    total_steps INTEGER DEFAULT 0,
    completed_steps INTEGER DEFAULT 0,
    failed_steps INTEGER DEFAULT 0
);

-- Individual workflow step execution tracking
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    
    -- Step identification
    node_id TEXT NOT NULL,
    node_type TEXT NOT NULL,
    step_number INTEGER NOT NULL,
    
    -- Step state
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    
    -- Timing
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    
    -- Data
    input_envelope JSONB,
    output_envelope JSONB,
    node_config JSONB,
    error_details JSONB,
    
    -- Worker assignment
    assigned_worker_id TEXT,
    worker_heartbeat TIMESTAMP WITH TIME ZONE,
    
    -- Dependencies
    depends_on UUID[] DEFAULT '{}',
    
    UNIQUE(run_id, node_id)
);

-- Worker pool management
CREATE TABLE IF NOT EXISTS workflow_workers (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    process_id INTEGER,
    version TEXT,
    capabilities TEXT[] DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'idle',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    max_concurrent_steps INTEGER DEFAULT 10,
    current_step_count INTEGER DEFAULT 0,
    total_steps_executed INTEGER DEFAULT 0,
    total_execution_time_ms BIGINT DEFAULT 0
);

-- Workflow execution queue
CREATE TABLE IF NOT EXISTS workflow_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE CASCADE,
    queue_type TEXT NOT NULL,
    priority INTEGER DEFAULT 5,
    available_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    claimed_at TIMESTAMP WITH TIME ZONE,
    claimed_by TEXT REFERENCES workflow_workers(id) ON DELETE SET NULL,
    attempt_count INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    payload JSONB
);

-- Execution checkpoints for recovery
CREATE TABLE IF NOT EXISTS workflow_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    checkpoint_type TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    execution_context JSONB,
    variables JSONB,
    trace_data JSONB,
    next_steps UUID[]
);

-- Execution events for debugging and monitoring
CREATE TABLE IF NOT EXISTS workflow_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE CASCADE,
    worker_id TEXT REFERENCES workflow_workers(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    event_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs(status);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_agent_id ON workflow_runs(agent_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_worker ON workflow_runs(assigned_worker_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_created ON workflow_runs(created_at);

CREATE INDEX IF NOT EXISTS idx_workflow_steps_run_id ON workflow_steps(run_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_status ON workflow_steps(status);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_worker ON workflow_steps(assigned_worker_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_retry ON workflow_steps(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_dependencies ON workflow_steps USING GIN (depends_on);

CREATE INDEX IF NOT EXISTS idx_workflow_workers_status ON workflow_workers(status);
CREATE INDEX IF NOT EXISTS idx_workflow_workers_heartbeat ON workflow_workers(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_workflow_workers_capabilities ON workflow_workers USING GIN (capabilities);

CREATE INDEX IF NOT EXISTS idx_workflow_queue_available ON workflow_queue(available_at, priority);
CREATE INDEX IF NOT EXISTS idx_workflow_queue_claimed ON workflow_queue(claimed_by);
CREATE INDEX IF NOT EXISTS idx_workflow_queue_run_id ON workflow_queue(run_id);
CREATE INDEX IF NOT EXISTS idx_workflow_queue_type ON workflow_queue(queue_type);

CREATE INDEX IF NOT EXISTS idx_workflow_checkpoints_run_id ON workflow_checkpoints(run_id);
CREATE INDEX IF NOT EXISTS idx_workflow_checkpoints_step_id ON workflow_checkpoints(step_id);
CREATE INDEX IF NOT EXISTS idx_workflow_checkpoints_created ON workflow_checkpoints(created_at);

CREATE INDEX IF NOT EXISTS idx_workflow_events_run_id ON workflow_events(run_id);
CREATE INDEX IF NOT EXISTS idx_workflow_events_created ON workflow_events(created_at);
CREATE INDEX IF NOT EXISTS idx_workflow_events_type ON workflow_events(event_type);

-- Add comments for documentation
COMMENT ON TABLE workflow_runs IS 'Enhanced workflow execution runs with step tracking and worker assignment';
COMMENT ON TABLE workflow_steps IS 'Individual node execution tracking within workflow runs';
COMMENT ON TABLE workflow_checkpoints IS 'Execution state snapshots for recovery and debugging';
COMMENT ON TABLE workflow_workers IS 'Worker pool management and health tracking';
COMMENT ON TABLE workflow_queue IS 'Distributed work queue for workflow execution coordination';
COMMENT ON TABLE workflow_events IS 'Execution event log for monitoring and debugging';