-- Normalize workflow storage for individual node and edge CRUD operations
-- Extends existing agents table to support proper workflow management

-- Workflow nodes table - stores individual nodes for live editing
CREATE TABLE IF NOT EXISTS workflow_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    node_id TEXT NOT NULL, -- client-generated node ID for referencing in edges
    node_type TEXT NOT NULL, -- corresponds to node definition types
    position_x FLOAT DEFAULT 0,
    position_y FLOAT DEFAULT 0,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(agent_id, node_id)
);

-- Workflow edges table - stores connections between nodes
CREATE TABLE IF NOT EXISTS workflow_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    edge_id TEXT NOT NULL, -- client-generated edge ID
    source_node_id TEXT NOT NULL,
    target_node_id TEXT NOT NULL,
    source_handle TEXT, -- for multi-output nodes
    target_handle TEXT, -- for multi-input nodes
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(agent_id, edge_id),
    FOREIGN KEY (agent_id, source_node_id) REFERENCES workflow_nodes(agent_id, node_id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id, target_node_id) REFERENCES workflow_nodes(agent_id, node_id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_workflow_nodes_agent_id ON workflow_nodes(agent_id);
CREATE INDEX IF NOT EXISTS idx_workflow_nodes_type ON workflow_nodes(node_type);
CREATE INDEX IF NOT EXISTS idx_workflow_edges_agent_id ON workflow_edges(agent_id);
CREATE INDEX IF NOT EXISTS idx_workflow_edges_source ON workflow_edges(agent_id, source_node_id);
CREATE INDEX IF NOT EXISTS idx_workflow_edges_target ON workflow_edges(agent_id, target_node_id);

-- Update function for updating updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updating updated_at
CREATE TRIGGER update_workflow_nodes_updated_at BEFORE UPDATE ON workflow_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();