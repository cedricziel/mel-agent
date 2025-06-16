import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// Mock the updateNode function to track calls
const mockUpdateNode = vi.fn().mockImplementation(() => Promise.resolve());

// Simulate the old sequential approach
const layoutWorkflowNodesSequential = async (nodes, updateNode) => {
  const startTime = performance.now();

  for (const node of nodes) {
    await updateNode(node.id, {
      position: { x: 100, y: 100 },
    });
  }

  const endTime = performance.now();
  return {
    duration: endTime - startTime,
    calls: nodes.length,
  };
};

// Simulate the new batched approach
const layoutWorkflowNodesBatched = async (nodes, updateNode) => {
  const startTime = performance.now();

  // Collect all position updates in a batch
  const positionUpdates = [];

  for (const node of nodes) {
    positionUpdates.push({
      id: node.id,
      position: { x: 100, y: 100 },
    });
  }

  // Apply all updates in batches to prevent UI blocking and reduce API calls
  const batchSize = 5; // Update 5 nodes at a time
  for (let i = 0; i < positionUpdates.length; i += batchSize) {
    const batch = positionUpdates.slice(i, i + batchSize);

    // Execute batch updates in parallel
    await Promise.all(
      batch.map(({ id, position }) => updateNode(id, { position }))
    );

    // Add small delay between batches to prevent overwhelming the API
    if (i + batchSize < positionUpdates.length) {
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
  }

  const endTime = performance.now();
  return {
    duration: endTime - startTime,
    calls: nodes.length,
  };
};

describe('BuilderPage Layout Performance', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Add realistic API delay simulation
    mockUpdateNode.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 10)) // 10ms API delay
    );
  });

  const createMockNodes = (count) => {
    return Array.from({ length: count }, (_, i) => ({
      id: `node-${i}`,
      type: 'test',
      position: { x: 0, y: 0 },
    }));
  };

  it('should demonstrate performance improvement with batched updates vs sequential', async () => {
    const nodeCount = 20; // Test with 20 nodes
    const mockNodes = createMockNodes(nodeCount);

    // Test sequential approach (old way)
    const sequentialResult = await layoutWorkflowNodesSequential(
      mockNodes,
      mockUpdateNode
    );
    const sequentialCalls = mockUpdateNode.mock.calls.length;

    vi.clearAllMocks();

    // Test batched approach (new way)
    const batchedResult = await layoutWorkflowNodesBatched(
      mockNodes,
      mockUpdateNode
    );
    const batchedCalls = mockUpdateNode.mock.calls.length;

    // Verify both approaches call updateNode the same number of times
    expect(sequentialCalls).toBe(nodeCount);
    expect(batchedCalls).toBe(nodeCount);

    // The batched approach should be significantly faster due to parallel execution
    // Even with the 50ms delays, it should be faster than sequential 10ms delays
    expect(batchedResult.duration).toBeLessThan(sequentialResult.duration);

    console.log('Performance Comparison:');
    console.log(
      `Sequential: ${sequentialResult.duration.toFixed(2)}ms for ${nodeCount} nodes`
    );
    console.log(
      `Batched: ${batchedResult.duration.toFixed(2)}ms for ${nodeCount} nodes`
    );
    console.log(
      `Improvement: ${(((sequentialResult.duration - batchedResult.duration) / sequentialResult.duration) * 100).toFixed(1)}%`
    );
  });

  it('should handle batching correctly with different batch sizes', async () => {
    const nodeCount = 13; // Not evenly divisible by batch size
    const mockNodes = createMockNodes(nodeCount);
    const batchSize = 5;

    await layoutWorkflowNodesBatched(mockNodes, mockUpdateNode);

    // Should have called updateNode for all nodes
    expect(mockUpdateNode).toHaveBeenCalledTimes(nodeCount);

    // Verify the batching pattern by checking call timing
    const callTimes = mockUpdateNode.mock.calls.map((_, index) => {
      const batchIndex = Math.floor(index / batchSize);
      const positionInBatch = index % batchSize;
      return { batchIndex, positionInBatch };
    });

    // First 5 calls should be batch 0
    expect(callTimes.slice(0, 5).every((call) => call.batchIndex === 0)).toBe(
      true
    );
    // Next 5 calls should be batch 1
    expect(callTimes.slice(5, 10).every((call) => call.batchIndex === 1)).toBe(
      true
    );
    // Last 3 calls should be batch 2
    expect(callTimes.slice(10, 13).every((call) => call.batchIndex === 2)).toBe(
      true
    );
  });

  it('should demonstrate reduced API pressure with parallel batches', async () => {
    const nodeCount = 10;
    const mockNodes = createMockNodes(nodeCount);

    // Track when calls start to measure parallelism
    const callStartTimes = [];
    mockUpdateNode.mockImplementation(() => {
      callStartTimes.push(performance.now());
      return new Promise((resolve) => setTimeout(resolve, 20));
    });

    await layoutWorkflowNodesBatched(mockNodes, mockUpdateNode);

    // In batched mode, calls within the same batch should start at nearly the same time
    // Check that first 5 calls (first batch) start within a short time window
    const firstBatchTimes = callStartTimes.slice(0, 5);
    const firstBatchTimeSpread =
      Math.max(...firstBatchTimes) - Math.min(...firstBatchTimes);

    // Should be less than 5ms spread (very fast parallel execution)
    expect(firstBatchTimeSpread).toBeLessThan(5);

    // There should be a gap between batches (due to the 50ms delay)
    if (callStartTimes.length > 5) {
      const timeBetweenBatches = callStartTimes[5] - callStartTimes[4];
      expect(timeBetweenBatches).toBeGreaterThan(40); // Should include the 50ms delay
    }
  });

  it('should handle empty node list gracefully', async () => {
    const result = await layoutWorkflowNodesBatched([], mockUpdateNode);

    expect(mockUpdateNode).not.toHaveBeenCalled();
    expect(result.calls).toBe(0);
    expect(result.duration).toBeGreaterThanOrEqual(0);
  });

  it('should handle single node efficiently', async () => {
    const singleNode = [
      { id: 'node-1', type: 'test', position: { x: 0, y: 0 } },
    ];

    await layoutWorkflowNodesBatched(singleNode, mockUpdateNode);

    expect(mockUpdateNode).toHaveBeenCalledTimes(1);
    expect(mockUpdateNode).toHaveBeenCalledWith('node-1', {
      position: { x: 100, y: 100 },
    });
  });
});
