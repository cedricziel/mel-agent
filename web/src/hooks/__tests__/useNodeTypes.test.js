import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import axios from 'axios';
import { useNodeTypes } from '../useNodeTypes.jsx';

// Mock axios
vi.mock('axios');

// Mock components
vi.mock('../components/DefaultNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/AgentNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/TriggerNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/ModelNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/ToolsNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/MemoryNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/OpenAIModelNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/AnthropicModelNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/LocalMemoryNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/IfNode', () => ({
  default: vi.fn(() => null),
}));

vi.mock('../components/HttpRequestNode', () => ({
  default: vi.fn(() => null),
}));

describe('useNodeTypes', () => {
  const mockAgentId = 'agent-1';
  const mockHandleNodeDelete = vi.fn();
  const mockOpenConfigDialog = vi.fn();
  const mockOnNodeClick = vi.fn();

  const mockNodeDefs = [
    {
      type: 'http_request',
      label: 'HTTP Request',
      category: 'Actions',
      description: 'Make HTTP requests',
    },
    {
      type: 'webhook',
      label: 'Webhook',
      category: 'Triggers',
      icon: 'webhook-icon',
    },
    {
      type: 'custom_node',
      label: 'Custom Node',
      category: 'Custom',
    },
  ];

  const mockTriggers = [
    {
      id: 'trigger-1',
      agent_id: 'agent-1',
      node_id: 'node-1',
      type: 'webhook',
    },
    {
      id: 'trigger-2',
      agent_id: 'agent-2',
      node_id: 'node-2',
      type: 'schedule',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    axios.get.mockImplementation((url) => {
      if (url === '/api/node-types') {
        return Promise.resolve({ data: mockNodeDefs });
      }
      if (url === '/api/triggers') {
        return Promise.resolve({ data: mockTriggers });
      }
      return Promise.reject(new Error('Unknown URL'));
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should load node definitions and triggers on mount', async () => {
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
      expect(result.current.triggers).toEqual(mockTriggers);
    });

    expect(axios.get).toHaveBeenCalledWith('/api/node-types');
    expect(axios.get).toHaveBeenCalledWith('/api/triggers');
  });

  it('should handle node definitions loading errors', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    axios.get.mockImplementation((url) => {
      if (url === '/api/node-types') {
        return Promise.reject(new Error('Network error'));
      }
      if (url === '/api/triggers') {
        return Promise.resolve({ data: mockTriggers });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith(
        'fetch node-types failed:',
        expect.any(Error)
      );
    });

    expect(result.current.nodeDefs).toEqual([]);
    consoleSpy.mockRestore();
  });

  it('should handle triggers loading errors', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    axios.get.mockImplementation((url) => {
      if (url === '/api/node-types') {
        return Promise.resolve({ data: mockNodeDefs });
      }
      if (url === '/api/triggers') {
        return Promise.reject(new Error('Network error'));
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith(
        'fetch triggers failed:',
        expect.any(Error)
      );
    });

    expect(result.current.triggers).toEqual([]);
    consoleSpy.mockRestore();
  });

  it('should create node types mapping correctly', async () => {
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
    });

    const nodeTypes = result.current.nodeTypes;

    // Should have default node types
    expect(nodeTypes.default).toBeDefined();
    expect(nodeTypes.agent).toBeDefined();
    expect(nodeTypes.if).toBeDefined();
    expect(nodeTypes.http_request).toBeDefined();

    // Should have trigger types
    expect(nodeTypes.webhook).toBeDefined();
    expect(nodeTypes.schedule).toBeDefined();

    // Should have custom node type
    expect(nodeTypes.custom_node).toBeDefined();
  });

  it('should create categories correctly', async () => {
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
    });

    const categories = result.current.categories;

    expect(categories).toHaveLength(3);
    expect(categories).toEqual([
      {
        category: 'Actions',
        types: [mockNodeDefs[0]],
      },
      {
        category: 'Triggers',
        types: [mockNodeDefs[1]],
      },
      {
        category: 'Custom',
        types: [mockNodeDefs[2]],
      },
    ]);
  });

  it('should create triggers map correctly', async () => {
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.triggers).toEqual(mockTriggers);
    });

    const triggersMap = result.current.triggersMap;

    // Should only include triggers for the current agent
    expect(triggersMap).toEqual({
      'node-1': mockTriggers[0],
    });
  });

  it('should refresh triggers successfully', async () => {
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.triggers).toEqual(mockTriggers);
    });

    // Mock updated triggers
    const updatedTriggers = [
      ...mockTriggers,
      {
        id: 'trigger-3',
        agent_id: 'agent-1',
        node_id: 'node-3',
        type: 'manual_trigger',
      },
    ];

    axios.get.mockImplementation((url) => {
      if (url === '/api/triggers') {
        return Promise.resolve({ data: updatedTriggers });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    await result.current.refreshTriggers();

    await waitFor(() => {
      expect(result.current.triggers).toEqual(updatedTriggers);
    });
  });

  it('should handle refresh triggers errors', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.triggers).toEqual(mockTriggers);
    });

    axios.get.mockImplementation((url) => {
      if (url === '/api/triggers') {
        return Promise.reject(new Error('Network error'));
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    await result.current.refreshTriggers();

    expect(consoleSpy).toHaveBeenCalledWith(
      'refresh triggers failed:',
      expect.any(Error)
    );
    consoleSpy.mockRestore();
  });

  it('should handle empty node definitions gracefully', async () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/node-types') {
        return Promise.resolve({ data: [] });
      }
      if (url === '/api/triggers') {
        return Promise.resolve({ data: mockTriggers });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual([]);
    });

    const categories = result.current.categories;
    expect(categories).toEqual([]);

    const nodeTypes = result.current.nodeTypes;
    expect(nodeTypes.default).toBeDefined();
    expect(nodeTypes.agent).toBeDefined();
  });

  it('should handle empty triggers gracefully', async () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/node-types') {
        return Promise.resolve({ data: mockNodeDefs });
      }
      if (url === '/api/triggers') {
        return Promise.resolve({ data: [] });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    const { result } = renderHook(() =>
      useNodeTypes(
        mockAgentId,
        mockHandleNodeDelete,
        mockOpenConfigDialog,
        mockOnNodeClick
      )
    );

    await waitFor(() => {
      expect(result.current.triggers).toEqual([]);
    });

    const triggersMap = result.current.triggersMap;
    expect(triggersMap).toEqual({});
  });

  it('should maintain stable nodeTypes reference when function dependencies are stable', async () => {
    const stableHandleNodeDelete = vi.fn();
    const stableOpenConfigDialog = vi.fn();
    const stableOnNodeClick = vi.fn();

    const { result, rerender } = renderHook(
      ({ agentId, handleNodeDelete, openConfigDialog, onNodeClick }) =>
        useNodeTypes(agentId, handleNodeDelete, openConfigDialog, onNodeClick),
      {
        initialProps: {
          agentId: mockAgentId,
          handleNodeDelete: stableHandleNodeDelete,
          openConfigDialog: stableOpenConfigDialog,
          onNodeClick: stableOnNodeClick,
        },
      }
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
    });

    const firstNodeTypes = result.current.nodeTypes;

    // Re-render with the same stable functions
    rerender({
      agentId: mockAgentId,
      handleNodeDelete: stableHandleNodeDelete,
      openConfigDialog: stableOpenConfigDialog,
      onNodeClick: stableOnNodeClick,
    });

    // nodeTypes should be the same reference (memoized)
    expect(result.current.nodeTypes).toBe(firstNodeTypes);
  });

  it('should update nodeTypes reference when function dependencies change', async () => {
    const stableHandleNodeDelete = vi.fn();
    const stableOpenConfigDialog = vi.fn();
    const stableOnNodeClick = vi.fn();

    const { result, rerender } = renderHook(
      ({ agentId, handleNodeDelete, openConfigDialog, onNodeClick }) =>
        useNodeTypes(agentId, handleNodeDelete, openConfigDialog, onNodeClick),
      {
        initialProps: {
          agentId: mockAgentId,
          handleNodeDelete: stableHandleNodeDelete,
          openConfigDialog: stableOpenConfigDialog,
          onNodeClick: stableOnNodeClick,
        },
      }
    );

    await waitFor(() => {
      expect(result.current.nodeDefs).toEqual(mockNodeDefs);
    });

    const firstNodeTypes = result.current.nodeTypes;

    // Re-render with new function instances
    const newHandleNodeDelete = vi.fn();
    const newOpenConfigDialog = vi.fn();
    const newOnNodeClick = vi.fn();

    rerender({
      agentId: mockAgentId,
      handleNodeDelete: newHandleNodeDelete,
      openConfigDialog: newOpenConfigDialog,
      onNodeClick: newOnNodeClick,
    });

    // nodeTypes should be a new reference (re-memoized)
    expect(result.current.nodeTypes).not.toBe(firstNodeTypes);
  });
});
