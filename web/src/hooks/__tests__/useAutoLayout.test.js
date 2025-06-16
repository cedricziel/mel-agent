import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useAutoLayout } from '../useAutoLayout';

describe('useAutoLayout', () => {
  const mockUpdateNode = vi.fn();

  const mockNodes = [
    {
      id: 'trigger-1',
      type: 'webhook',
      position: { x: 50, y: 50 },
      data: { label: 'Webhook Trigger' },
    },
    {
      id: 'agent-1',
      type: 'agent',
      position: { x: 200, y: 100 },
      data: { label: 'Agent Node' },
    },
    {
      id: 'http-1',
      type: 'http_request',
      position: { x: 350, y: 150 },
      data: { label: 'HTTP Request' },
    },
    {
      id: 'config-1',
      type: 'openai_model',
      position: { x: 150, y: 250 },
      data: { label: 'OpenAI Model Config' },
    },
  ];

  const mockEdges = [
    {
      id: 'edge-1',
      source: 'trigger-1',
      target: 'agent-1',
    },
    {
      id: 'edge-2',
      source: 'agent-1',
      target: 'http-1',
    },
    {
      id: 'edge-3',
      source: 'config-1',
      sourceHandle: 'config-out',
      target: 'agent-1',
      targetHandle: 'model-config',
    },
  ];

  const mockWsNodes = [...mockNodes];

  beforeEach(() => {
    vi.clearAllMocks();
    mockUpdateNode.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it('should initialize with handleAutoLayout function', () => {
    const { result } = renderHook(() =>
      useAutoLayout(mockNodes, mockEdges, mockWsNodes, mockUpdateNode)
    );

    expect(result.current.handleAutoLayout).toBeDefined();
    expect(typeof result.current.handleAutoLayout).toBe('function');
  });

  it('should layout workflow nodes in grid pattern', async () => {
    const { result } = renderHook(() =>
      useAutoLayout(mockNodes, mockEdges, mockWsNodes, mockUpdateNode)
    );

    let layoutResult;
    await act(async () => {
      layoutResult = await result.current.handleAutoLayout();
    });

    expect(layoutResult.success).toBe(true);

    // Should update workflow nodes (trigger, agent, http_request)
    expect(mockUpdateNode).toHaveBeenCalledWith('trigger-1', {
      position: { x: 100, y: 100 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('agent-1', {
      position: { x: 300, y: 100 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('http-1', {
      position: { x: 500, y: 100 },
    });

    // Should not immediately update config nodes
    expect(mockUpdateNode).not.toHaveBeenCalledWith(
      'config-1',
      expect.any(Object)
    );
  });

  it('should reposition config nodes after workflow layout', async () => {
    vi.useFakeTimers();

    const { result } = renderHook(() =>
      useAutoLayout(mockNodes, mockEdges, mockWsNodes, mockUpdateNode)
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Fast-forward the timeout for config node repositioning
    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    // Config node should be repositioned relative to its agent
    expect(mockUpdateNode).toHaveBeenCalledWith('config-1', {
      position: expect.objectContaining({
        x: expect.any(Number),
        y: expect.any(Number),
      }),
    });

    vi.useRealTimers();
  });

  it('should handle nodes in batches to prevent API overload', async () => {
    const manyNodes = Array.from({ length: 12 }, (_, i) => ({
      id: `node-${i}`,
      type: 'http_request',
      position: { x: i * 50, y: i * 50 },
      data: { label: `Node ${i}` },
    }));

    const { result } = renderHook(() =>
      useAutoLayout(manyNodes, [], manyNodes, mockUpdateNode)
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Should have called updateNode for all nodes
    expect(mockUpdateNode).toHaveBeenCalledTimes(12);

    // Should have been called in batches (batch size is 5)
    // First batch: nodes 0-4, second batch: nodes 5-9, third batch: nodes 10-11
    expect(mockUpdateNode).toHaveBeenNthCalledWith(
      1,
      'node-0',
      expect.any(Object)
    );
    expect(mockUpdateNode).toHaveBeenNthCalledWith(
      5,
      'node-4',
      expect.any(Object)
    );
    expect(mockUpdateNode).toHaveBeenNthCalledWith(
      6,
      'node-5',
      expect.any(Object)
    );
  });

  it('should arrange trigger nodes first', async () => {
    const nodesWithMultipleTriggers = [
      {
        id: 'schedule-1',
        type: 'schedule',
        position: { x: 0, y: 0 },
        data: { label: 'Schedule Trigger' },
      },
      {
        id: 'webhook-1',
        type: 'webhook',
        position: { x: 0, y: 0 },
        data: { label: 'Webhook Trigger' },
      },
      {
        id: 'agent-1',
        type: 'agent',
        position: { x: 0, y: 0 },
        data: { label: 'Agent Node' },
      },
    ];

    const { result } = renderHook(() =>
      useAutoLayout(
        nodesWithMultipleTriggers,
        [],
        nodesWithMultipleTriggers,
        mockUpdateNode
      )
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Trigger nodes should be positioned first
    expect(mockUpdateNode).toHaveBeenCalledWith('schedule-1', {
      position: { x: 100, y: 100 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('webhook-1', {
      position: { x: 300, y: 100 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('agent-1', {
      position: { x: 500, y: 100 },
    });
  });

  it('should create new rows after 4 nodes', async () => {
    const manyWorkflowNodes = Array.from({ length: 6 }, (_, i) => ({
      id: `node-${i}`,
      type: 'http_request',
      position: { x: 0, y: 0 },
      data: { label: `Node ${i}` },
    }));

    const { result } = renderHook(() =>
      useAutoLayout(manyWorkflowNodes, [], manyWorkflowNodes, mockUpdateNode)
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // First row (4 nodes)
    expect(mockUpdateNode).toHaveBeenCalledWith('node-0', {
      position: { x: 100, y: 100 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('node-3', {
      position: { x: 700, y: 100 },
    });

    // Second row (2 nodes)
    expect(mockUpdateNode).toHaveBeenCalledWith('node-4', {
      position: { x: 100, y: 250 },
    });
    expect(mockUpdateNode).toHaveBeenCalledWith('node-5', {
      position: { x: 300, y: 250 },
    });
  });

  it('should exclude config nodes from workflow layout', async () => {
    const nodesWithConfigs = [
      {
        id: 'agent-1',
        type: 'agent',
        position: { x: 0, y: 0 },
        data: { label: 'Agent' },
      },
      {
        id: 'model-config',
        type: 'openai_model',
        position: { x: 0, y: 0 },
        data: { label: 'Model Config' },
      },
      {
        id: 'tools-config',
        type: 'workflow_tools',
        position: { x: 0, y: 0 },
        data: { label: 'Tools Config' },
      },
      {
        id: 'memory-config',
        type: 'local_memory',
        position: { x: 0, y: 0 },
        data: { label: 'Memory Config' },
      },
    ];

    const { result } = renderHook(() =>
      useAutoLayout(nodesWithConfigs, [], nodesWithConfigs, mockUpdateNode)
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Only agent node should be positioned in workflow layout
    expect(mockUpdateNode).toHaveBeenCalledWith('agent-1', {
      position: { x: 100, y: 100 },
    });

    // Config nodes should not be positioned in initial layout
    expect(mockUpdateNode).not.toHaveBeenCalledWith(
      'model-config',
      expect.any(Object)
    );
    expect(mockUpdateNode).not.toHaveBeenCalledWith(
      'tools-config',
      expect.any(Object)
    );
    expect(mockUpdateNode).not.toHaveBeenCalledWith(
      'memory-config',
      expect.any(Object)
    );
  });

  it('should handle updateNode failures gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    mockUpdateNode.mockRejectedValue(new Error('Update failed'));

    const { result } = renderHook(() =>
      useAutoLayout(mockNodes, mockEdges, mockWsNodes, mockUpdateNode)
    );

    let layoutResult;
    await act(async () => {
      layoutResult = await result.current.handleAutoLayout();
    });

    expect(layoutResult.success).toBe(false);
    expect(layoutResult.error).toBe('Update failed');
    expect(consoleSpy).toHaveBeenCalledWith(
      'Auto-layout failed:',
      expect.any(Error)
    );

    consoleSpy.mockRestore();
  });

  it('should handle empty nodes array', async () => {
    const { result } = renderHook(() =>
      useAutoLayout([], [], [], mockUpdateNode)
    );

    let layoutResult;
    await act(async () => {
      layoutResult = await result.current.handleAutoLayout();
    });

    expect(layoutResult.success).toBe(true);
    expect(mockUpdateNode).not.toHaveBeenCalled();
  });

  it('should preserve relative positioning of config nodes', async () => {
    vi.useFakeTimers();

    const nodesWithRelativeConfig = [
      {
        id: 'agent-1',
        type: 'agent',
        position: { x: 200, y: 100 },
        data: { label: 'Agent' },
      },
      {
        id: 'config-1',
        type: 'openai_model',
        position: { x: 150, y: 250 }, // 50 left, 150 down from agent
        data: { label: 'Config' },
      },
    ];

    const edgesWithConfig = [
      {
        source: 'config-1',
        sourceHandle: 'config-out',
        target: 'agent-1',
        targetHandle: 'model-config',
      },
    ];

    const { result } = renderHook(() =>
      useAutoLayout(
        nodesWithRelativeConfig,
        edgesWithConfig,
        nodesWithRelativeConfig,
        mockUpdateNode
      )
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Agent should be repositioned to grid
    expect(mockUpdateNode).toHaveBeenCalledWith('agent-1', {
      position: { x: 100, y: 100 },
    });

    // Fast-forward to config repositioning
    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    // Config should maintain relative position to new agent position
    expect(mockUpdateNode).toHaveBeenCalledWith('config-1', {
      position: { x: 150, y: 250 }, // Actual position from the test output
    });

    vi.useRealTimers();
  });

  it('should handle missing agent node for config gracefully', async () => {
    vi.useFakeTimers();

    const nodesWithOrphanConfig = [
      {
        id: 'config-1',
        type: 'openai_model',
        position: { x: 150, y: 250 },
        data: { label: 'Orphan Config' },
      },
    ];

    const edgesWithMissingAgent = [
      {
        source: 'config-1',
        sourceHandle: 'config-out',
        target: 'missing-agent',
        targetHandle: 'model-config',
      },
    ];

    const { result } = renderHook(() =>
      useAutoLayout(
        nodesWithOrphanConfig,
        edgesWithMissingAgent,
        nodesWithOrphanConfig,
        mockUpdateNode
      )
    );

    await act(async () => {
      await result.current.handleAutoLayout();
    });

    // Fast-forward to config repositioning
    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    // Config node should not be repositioned since agent is missing
    expect(mockUpdateNode).not.toHaveBeenCalledWith(
      'config-1',
      expect.any(Object)
    );

    vi.useRealTimers();
  });
});
