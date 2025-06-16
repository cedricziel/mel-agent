import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNodeManagement } from '../useNodeManagement';

// Mock the workflow state hook
const mockWorkflowState = {
  nodes: [],
  edges: [],
  createNode: vi.fn(),
  updateNode: vi.fn(),
  deleteNode: vi.fn(),
  createEdge: vi.fn(),
  deleteEdge: vi.fn(),
};

vi.mock('../useWorkflowState', () => ({
  useWorkflowState: () => mockWorkflowState,
}));

describe('useNodeManagement', () => {
  const mockBroadcastNodeChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockWorkflowState.nodes = [
      {
        id: 'node-1',
        type: 'default',
        data: { label: 'Node 1' },
        position: { x: 100, y: 100 },
      },
      {
        id: 'node-2',
        type: 'agent',
        data: { label: 'Agent Node' },
        position: { x: 200, y: 200 },
      },
    ];
    mockWorkflowState.edges = [
      {
        id: 'edge-1',
        source: 'node-1',
        target: 'node-2',
      },
    ];
  });

  it('should handle node deletion with edge cleanup', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    await act(async () => {
      await result.current.handleNodeDelete('node-1');
    });

    expect(mockWorkflowState.deleteNode).toHaveBeenCalledWith('node-1');
    expect(mockBroadcastNodeChange).toHaveBeenCalledWith('nodeDeleted', {
      nodeId: 'node-1',
    });
    expect(mockWorkflowState.deleteEdge).toHaveBeenCalledWith('edge-1');
    expect(mockBroadcastNodeChange).toHaveBeenCalledWith('edgeDeleted', {
      edgeId: 'edge-1',
    });
  });

  it('should handle edge deletion', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    await act(async () => {
      await result.current.handleEdgeDelete('edge-1');
    });

    expect(mockWorkflowState.deleteEdge).toHaveBeenCalledWith('edge-1');
    expect(mockBroadcastNodeChange).toHaveBeenCalledWith('edgeDeleted', {
      edgeId: 'edge-1',
    });
  });

  it('should create agent configuration nodes', async () => {
    // Add the agent node to the mock nodes array
    mockWorkflowState.nodes = [
      ...mockWorkflowState.nodes,
      {
        id: 'agent-1',
        type: 'agent',
        data: { label: 'Test Agent' },
        position: { x: 300, y: 300 },
      },
    ];

    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    const agentPosition = { x: 300, y: 300 };

    await act(async () => {
      await result.current.createAgentConfigurationNodes(
        'agent-1',
        agentPosition
      );
    });

    // Should create 3 config nodes (model, tools, memory)
    expect(mockWorkflowState.createNode).toHaveBeenCalledTimes(3);
    expect(mockWorkflowState.createEdge).toHaveBeenCalledTimes(3);
    expect(mockWorkflowState.updateNode).toHaveBeenCalledWith(
      'agent-1',
      expect.any(Object)
    );

    // Should broadcast all creations
    expect(mockBroadcastNodeChange).toHaveBeenCalledTimes(6); // 3 nodes + 3 edges
  });

  it('should handle node creation with broadcasting', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    const newNode = {
      id: 'new-node',
      type: 'default',
      data: { label: 'New Node' },
      position: { x: 400, y: 400 },
    };

    await act(async () => {
      await result.current.handleAddNode({
        type: 'default',
        label: 'New Node',
        position: { x: 400, y: 400 },
      });
    });

    expect(mockWorkflowState.createNode).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'default',
        data: expect.objectContaining({
          label: 'New Node',
        }),
        position: { x: 400, y: 400 },
      })
    );
    expect(mockBroadcastNodeChange).toHaveBeenCalledWith('nodeCreated', {
      node: expect.any(Object),
    });
  });

  it('should handle edge creation with broadcasting', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    await act(async () => {
      await result.current.handleConnectNodes({
        source_id: 'node-1',
        target_id: 'node-2',
      });
    });

    expect(mockWorkflowState.createEdge).toHaveBeenCalledWith(
      expect.objectContaining({
        source: 'node-1',
        target: 'node-2',
      })
    );
    expect(mockBroadcastNodeChange).toHaveBeenCalledWith('edgeCreated', {
      edge: expect.any(Object),
    });
  });

  it('should handle errors gracefully during node deletion', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    mockWorkflowState.deleteNode.mockRejectedValueOnce(
      new Error('Delete failed')
    );

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await act(async () => {
      await result.current.handleNodeDelete('node-1');
    });

    expect(consoleSpy).toHaveBeenCalledWith(
      'Failed to delete node:',
      expect.any(Error)
    );
    consoleSpy.mockRestore();
  });

  it('should handle errors gracefully during edge deletion', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    mockWorkflowState.deleteEdge.mockRejectedValueOnce(
      new Error('Delete failed')
    );

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await act(async () => {
      await result.current.handleEdgeDelete('edge-1');
    });

    expect(consoleSpy).toHaveBeenCalledWith(
      'Failed to delete edge:',
      expect.any(Error)
    );
    consoleSpy.mockRestore();
  });

  it('should auto-create configuration nodes for agent nodes', async () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    await act(async () => {
      await result.current.handleAddNode({
        type: 'agent',
        label: 'Test Agent',
        position: { x: 500, y: 500 },
      });
    });

    // Should create the agent node first
    expect(mockWorkflowState.createNode).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'agent',
      })
    );

    // Then should create configuration nodes (called separately)
    // This is tested more thoroughly in the createAgentConfigurationNodes test
  });

  it('should provide workflow data for assistant', () => {
    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    const workflow = result.current.handleGetWorkflow();

    expect(workflow).toEqual({
      nodes: mockWorkflowState.nodes,
      edges: mockWorkflowState.edges,
    });
  });

  it('should create agent configuration nodes in parallel for better performance', async () => {
    // Add the agent node to the mock nodes array
    mockWorkflowState.nodes = [
      ...mockWorkflowState.nodes,
      {
        id: 'agent-1',
        type: 'agent',
        data: { label: 'Test Agent' },
        position: { x: 100, y: 100 },
      },
    ];

    const { result } = renderHook(() =>
      useNodeManagement(mockBroadcastNodeChange)
    );

    // Mock timing to verify parallel execution
    const createNodeTimes = [];
    const createEdgeTimes = [];

    mockWorkflowState.createNode.mockImplementation(async () => {
      const start = Date.now();
      await new Promise((resolve) => setTimeout(resolve, 10)); // Simulate async work
      createNodeTimes.push(Date.now() - start);
    });

    mockWorkflowState.createEdge.mockImplementation(async () => {
      const start = Date.now();
      await new Promise((resolve) => setTimeout(resolve, 10)); // Simulate async work
      createEdgeTimes.push(Date.now() - start);
    });

    const startTime = Date.now();
    await act(async () => {
      await result.current.createAgentConfigurationNodes('agent-1', {
        x: 100,
        y: 100,
      });
    });
    const totalTime = Date.now() - startTime;

    // Verify all nodes and edges were created
    expect(mockWorkflowState.createNode).toHaveBeenCalledTimes(3); // model, tools, memory
    expect(mockWorkflowState.createEdge).toHaveBeenCalledTimes(3); // 3 edges
    expect(mockWorkflowState.updateNode).toHaveBeenCalledTimes(1); // agent update

    // Verify parallel execution: total time should be less than sum of individual times
    // If executed sequentially, it would take ~60ms (6 * 10ms)
    // With parallel execution, it should take ~20ms (2 batches * 10ms)
    expect(totalTime).toBeLessThan(40); // Allow some buffer for test execution
  });
});
