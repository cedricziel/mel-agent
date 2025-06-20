-- Migration 016: Create webhook_events table for tracking webhook calls
-- This table stores all incoming webhook requests for auditing and debugging

CREATE TABLE IF NOT EXISTS webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_id UUID NOT NULL REFERENCES triggers(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    headers JSONB,
    source_ip TEXT,
    user_agent TEXT,
    status TEXT NOT NULL DEFAULT 'received',
    processed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_webhook_events_trigger_id ON webhook_events(trigger_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status);

-- Add constraint for status values
ALTER TABLE webhook_events
  ADD CONSTRAINT check_webhook_event_status CHECK (status IN ('received', 'processed', 'failed', 'ignored'));

-- Add comments for documentation
COMMENT ON TABLE webhook_events IS 'Stores incoming webhook requests for auditing and debugging';
COMMENT ON COLUMN webhook_events.trigger_id IS 'Reference to the trigger that received this webhook';
COMMENT ON COLUMN webhook_events.payload IS 'The JSON payload received from the webhook';
COMMENT ON COLUMN webhook_events.headers IS 'HTTP headers from the webhook request';
COMMENT ON COLUMN webhook_events.status IS 'Processing status: received, processed, failed, ignored';