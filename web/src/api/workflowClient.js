import { workflowsApi } from './client';

class WorkflowClient {
  constructor() {
    // Use generated API clients
    this.workflowsApi = workflowsApi;
  }

  // Workflow management
  async getWorkflow(workflowId) {
    const response = await this.workflowsApi.getWorkflow(workflowId);
    return response.data;
  }

  async updateWorkflow(workflowId, updates) {
    const response = await this.workflowsApi.updateWorkflow(
      workflowId,
      updates
    );
    return response.data;
  }

  async deleteWorkflow(workflowId) {
    await this.workflowsApi.deleteWorkflow(workflowId);
  }

  // Node management
  async getNodes(workflowId) {
    const response = await this.workflowsApi.listWorkflowNodes(workflowId);
    return response.data;
  }

  async createNode(workflowId, nodeData) {
    const response = await this.workflowsApi.createWorkflowNode(
      workflowId,
      nodeData
    );
    return response.data;
  }

  async getNode(workflowId, nodeId) {
    const response = await this.workflowsApi.getWorkflowNode(
      workflowId,
      nodeId
    );
    return response.data;
  }

  async updateNode(workflowId, nodeId, updates) {
    const response = await this.workflowsApi.updateWorkflowNode(
      workflowId,
      nodeId,
      updates
    );
    return response.data;
  }

  async deleteNode(workflowId, nodeId) {
    await this.workflowsApi.deleteWorkflowNode(workflowId, nodeId);
  }

  // Edge management
  async getEdges(workflowId) {
    const response = await this.workflowsApi.listWorkflowEdges(workflowId);
    return response.data;
  }

  async createEdge(workflowId, edgeData) {
    const response = await this.workflowsApi.createWorkflowEdge(
      workflowId,
      edgeData
    );
    return response.data;
  }

  async deleteEdge(workflowId, edgeId) {
    await this.workflowsApi.deleteWorkflowEdge(workflowId, edgeId);
  }

  // Layout management
  async autoLayout(workflowId) {
    const response = await this.workflowsApi.autoLayoutWorkflow(workflowId);
    return response.data;
  }

  // Data format transformation utilities
  toReactFlowNode(apiNode) {
    return {
      id: apiNode.node_id,
      type: apiNode.node_type,
      position: {
        x: apiNode.position_x || 0,
        y: apiNode.position_y || 0,
      },
      data: {
        ...apiNode.config,
        label: apiNode.config?.label || apiNode.node_type,
      },
    };
  }

  toApiNode(rfNode) {
    return {
      node_id: rfNode.id,
      node_type: rfNode.type,
      position_x: rfNode.position.x,
      position_y: rfNode.position.y,
      config: {
        ...rfNode.data,
        // Remove ReactFlow-specific properties
        label: rfNode.data.label,
      },
    };
  }

  toReactFlowEdge(apiEdge) {
    return {
      id: apiEdge.edge_id,
      source: apiEdge.source_node_id,
      target: apiEdge.target_node_id,
      sourceHandle: apiEdge.source_handle || null,
      targetHandle: apiEdge.target_handle || null,
      type: 'default',
    };
  }

  toApiEdge(rfEdge) {
    return {
      edge_id: rfEdge.id,
      source_node_id: rfEdge.source,
      target_node_id: rfEdge.target,
      source_handle: rfEdge.sourceHandle || null,
      target_handle: rfEdge.targetHandle || null,
    };
  }

  // Load complete workflow with nodes and edges
  async loadWorkflowData(workflowId) {
    const [workflow, nodes, edges] = await Promise.all([
      this.getWorkflow(workflowId),
      this.getNodes(workflowId),
      this.getEdges(workflowId),
    ]);

    return {
      workflow,
      nodes: nodes.map((node) => this.toReactFlowNode(node)),
      edges: edges.map((edge) => this.toReactFlowEdge(edge)),
    };
  }

  // Save entire workflow as version
  async saveWorkflowVersion(workflowId, graph) {
    const response = await this.workflowsApi.createWorkflowVersion(workflowId, {
      name: '1.0.0',
      definition: { nodes: graph.nodes, edges: graph.edges },
    });
    return response.data;
  }

  // Load latest version
  async getLatestWorkflowVersion(workflowId) {
    const response =
      await this.workflowsApi.getLatestWorkflowVersion(workflowId);
    return response.data;
  }
}

// Create singleton instance
const workflowClient = new WorkflowClient();

export default workflowClient;
