import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useRunsData } from '../useRunsData';
import { nodeTypesApi, workflowRunsApi } from '../../api/client';

// Mock the API clients
vi.mock('../../api/client', () => ({
  nodeTypesApi: {
    listNodeTypes: vi.fn(),
  },
  workflowRunsApi: {
    listWorkflowRuns: vi.fn(),
    getWorkflowRun: vi.fn(),
  },
}));

describe('useRunsData', () => {
  const mockAgentId = 'agent-123';
  const mockRuns = [
    { id: 'run-1', created_at: '2024-01-01T10:00:00Z' },
    { id: 'run-2', created_at: '2024-01-02T10:00:00Z' },
  ];
  const mockRunDetails = {
    id: 'run-1',
    graph: {
      nodes: [
        { id: 'node-1', type: 'webhook', position: { x: 0, y: 0 } },
        { id: 'node-2', type: 'agent', position: { x: 200, y: 0 } },
      ],
      edges: [{ id: 'edge-1', source: 'node-1', target: 'node-2' }],
    },
    trace: [
      { nodeId: 'node-1', status: 'completed', output: 'test output' },
      { nodeId: 'node-2', status: 'running' },
    ],
  };
  const mockNodeDefs = [
    { type: 'webhook', label: 'Webhook', entry_point: true },
    { type: 'agent', label: 'Agent', entry_point: false },
  ];

  beforeEach(() => {
    vi.clearAllMocks();

    // Mock nodeTypesApi
    nodeTypesApi.listNodeTypes.mockResolvedValue({ data: mockNodeDefs });

    // Mock workflowRunsApi
    workflowRunsApi.listWorkflowRuns.mockImplementation((workflowId) => {
      if (workflowId === mockAgentId) {
        return Promise.resolve({ data: { runs: mockRuns } });
      }
      return Promise.resolve({ data: { runs: [] } });
    });

    workflowRunsApi.getWorkflowRun.mockImplementation((runId) => {
      if (runId === 'run-1') {
        return Promise.resolve({ data: mockRunDetails });
      }
      return Promise.reject(new Error('Run not found'));
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with default values', () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    expect(result.current.runs).toEqual([]);
    expect(result.current.runDetails).toBeNull();
    expect(result.current.nodeDefs).toEqual([]);
    expect(result.current.selectedRunID).toBeNull();
    expect(result.current.selectedRunNodeID).toBeNull();
    expect(result.current.selectedRunStep).toBeUndefined();
    expect(result.current.selectedRunNodeDef).toBeUndefined();
    expect(result.current.rfNodes).toEqual([]);
    expect(result.current.rfEdges).toEqual([]);
  });

  it('should fetch node definitions on mount', async () => {
    renderHook(() => useRunsData(mockAgentId));

    await waitFor(() => {
      expect(nodeTypesApi.listNodeTypes).toHaveBeenCalled();
    });
  });

  it('should fetch runs when agentId is provided', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    await waitFor(() => {
      expect(workflowRunsApi.listWorkflowRuns).toHaveBeenCalledWith(
        mockAgentId
      );
      expect(result.current.runs).toEqual(mockRuns);
    });
  });

  it('should not fetch runs when agentId is not provided', () => {
    renderHook(() => useRunsData(null));

    expect(workflowRunsApi.listWorkflowRuns).not.toHaveBeenCalled();
  });

  it('should fetch run details when run is selected', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    await waitFor(() => {
      expect(result.current.runs).toEqual(mockRuns);
    });

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(workflowRunsApi.getWorkflowRun).toHaveBeenCalledWith('run-1');
      expect(result.current.selectedRunID).toBe('run-1');
      expect(result.current.runDetails).toEqual(mockRunDetails);
    });
  });

  it('should derive rfNodes and rfEdges from run details', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.rfNodes).toEqual(mockRunDetails.graph.nodes);
      expect(result.current.rfEdges).toEqual(mockRunDetails.graph.edges);
    });
  });

  it('should handle node selection', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.runDetails).toEqual(mockRunDetails);
    });

    act(() => {
      result.current.handleNodeSelect('node-1');
    });

    expect(result.current.selectedRunNodeID).toBe('node-1');
    expect(result.current.selectedRunStep).toEqual(mockRunDetails.trace[0]);
  });

  it('should derive selectedRunNodeDef correctly', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    // Wait for node defs and run details to load
    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
    });

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.runDetails).toEqual(mockRunDetails);
    });

    act(() => {
      result.current.handleNodeSelect('node-1');
    });

    expect(result.current.selectedRunNodeDef).toEqual(mockNodeDefs[0]);
  });

  it('should reset selected node when run details change', async () => {
    // Add mock for run-2
    workflowRunsApi.getWorkflowRun.mockImplementation((runId) => {
      if (runId === 'run-1') {
        return Promise.resolve({ data: mockRunDetails });
      }
      if (runId === 'run-2') {
        return Promise.resolve({ data: { ...mockRunDetails, id: 'run-2' } });
      }
      return Promise.reject(new Error('Run not found'));
    });

    const { result } = renderHook(() => useRunsData(mockAgentId));

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.runDetails).toEqual(mockRunDetails);
    });

    act(() => {
      result.current.handleNodeSelect('node-1');
    });

    expect(result.current.selectedRunNodeID).toBe('node-1');

    // Select a different run
    act(() => {
      result.current.handleRunSelect('run-2');
    });

    await waitFor(() => {
      expect(result.current.selectedRunNodeID).toBeNull();
    });
  });

  it('should handle pane click to deselect node', async () => {
    const { result } = renderHook(() => useRunsData(mockAgentId));

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.runDetails).toEqual(mockRunDetails);
    });

    act(() => {
      result.current.handleNodeSelect('node-1');
    });

    expect(result.current.selectedRunNodeID).toBe('node-1');

    act(() => {
      result.current.handlePaneClick();
    });

    expect(result.current.selectedRunNodeID).toBeNull();
  });

  it('should handle API errors gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    nodeTypesApi.listNodeTypes.mockRejectedValue(new Error('API Error'));
    workflowRunsApi.listWorkflowRuns.mockRejectedValue(new Error('API Error'));

    renderHook(() => useRunsData(mockAgentId));

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith(
        'fetch node-types failed',
        expect.any(Error)
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        'fetch runs list failed',
        expect.any(Error)
      );
    });

    consoleSpy.mockRestore();
  });

  it('should handle empty run details gracefully', async () => {
    // Override mock to return null for run details
    workflowRunsApi.getWorkflowRun.mockResolvedValue({ data: null });

    const { result } = renderHook(() => useRunsData(mockAgentId));

    act(() => {
      result.current.handleRunSelect('run-1');
    });

    await waitFor(() => {
      expect(result.current.runDetails).toBeNull();
      expect(result.current.rfNodes).toEqual([]);
      expect(result.current.rfEdges).toEqual([]);
    });
  });
});
