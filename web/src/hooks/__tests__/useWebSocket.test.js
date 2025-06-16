import { renderHook, act } from '@testing-library/react';
import {
  describe,
  it,
  expect,
  vi,
  beforeEach,
  afterEach,
  afterAll,
} from 'vitest';
import { useWebSocket } from '../useWebSocket';

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    this.onopen = null;
    this.onclose = null;
    this.onmessage = null;
    this.onerror = null;
    this.sentMessages = [];

    // Simulate connection after a short delay
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      if (this.onopen) this.onopen();
    }, 10);
  }

  send(data) {
    this.sentMessages.push(data);
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) this.onclose();
  }

  // Helper method to simulate receiving messages
  simulateMessage(data) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) });
    }
  }

  // Helper method to simulate errors
  simulateError(error) {
    if (this.onerror) this.onerror(error);
  }
}

describe('useWebSocket', () => {
  // Save original WebSocket to restore later
  const originalWebSocket = global.WebSocket;
  let mockWebSocket;

  beforeEach(() => {
    // Mock global WebSocket for each test
    global.WebSocket = MockWebSocket;
    vi.clearAllMocks();
    // Store reference to created WebSocket instances
    const originalWebSocket = global.WebSocket;
    global.WebSocket = vi.fn().mockImplementation((url) => {
      mockWebSocket = new originalWebSocket(url);
      return mockWebSocket;
    });
  });

  afterEach(() => {
    if (mockWebSocket) {
      mockWebSocket.close();
    }
  });

  afterAll(() => {
    // Restore original WebSocket to prevent test leakage
    global.WebSocket = originalWebSocket;
  });

  it('should establish WebSocket connection', async () => {
    const { result } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id')
    );

    expect(global.WebSocket).toHaveBeenCalledWith(
      'ws://localhost:3000/api/ws/agents/test-agent-id'
    );

    // Wait for connection to be established
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    expect(result.current.isConnected).toBe(true);
  });

  it('should handle incoming messages', async () => {
    const onNodeUpdated = vi.fn();
    const onNodeCreated = vi.fn();
    const onNodeDeleted = vi.fn();

    const { result } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id', {
        onNodeUpdated,
        onNodeCreated,
        onNodeDeleted,
      })
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    // Simulate node updated message
    act(() => {
      mockWebSocket.simulateMessage({
        type: 'nodeUpdated',
        clientId: 'other-client',
        workflowId: 'test-agent-id',
        nodeId: 'node-1',
        data: { position: { x: 100, y: 200 } },
      });
    });

    expect(onNodeUpdated).toHaveBeenCalledWith('node-1', {
      position: { x: 100, y: 200 },
    });
  });

  it('should ignore messages from same client', async () => {
    const onNodeUpdated = vi.fn();

    renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id', {
        onNodeUpdated,
      })
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    // Simulate message from same client
    act(() => {
      mockWebSocket.simulateMessage({
        type: 'nodeUpdated',
        clientId: 'test-client-id', // Same as our client ID
        workflowId: 'test-agent-id',
        nodeId: 'node-1',
        data: { position: { x: 100, y: 200 } },
      });
    });

    expect(onNodeUpdated).not.toHaveBeenCalled();
  });

  it('should broadcast node changes', async () => {
    const { result } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id')
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    act(() => {
      result.current.broadcastNodeChange('nodeUpdated', {
        nodeId: 'node-1',
        data: { position: { x: 100, y: 200 } },
      });
    });

    expect(mockWebSocket.sentMessages).toHaveLength(1);
    const sentMessage = JSON.parse(mockWebSocket.sentMessages[0]);
    expect(sentMessage).toEqual({
      clientId: 'test-client-id',
      type: 'nodeUpdated',
      workflowId: 'test-agent-id',
      nodeId: 'node-1',
      data: { position: { x: 100, y: 200 } },
    });
  });

  it('should handle node execution messages', async () => {
    const onNodeExecution = vi.fn();

    renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id', {
        onNodeExecution,
      })
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    // Simulate node execution start
    act(() => {
      mockWebSocket.simulateMessage({
        type: 'nodeExecution',
        clientId: 'other-client',
        workflowId: 'test-agent-id',
        nodeId: 'node-1',
        phase: 'start',
      });
    });

    expect(onNodeExecution).toHaveBeenCalledWith('node-1', 'start');
  });

  it('should cleanup on unmount', async () => {
    const { unmount } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id')
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    const closeSpy = vi.spyOn(mockWebSocket, 'close');

    unmount();

    expect(closeSpy).toHaveBeenCalled();
  });

  it('should handle connection errors gracefully', async () => {
    const { result } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id')
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    act(() => {
      mockWebSocket.simulateError(new Error('Connection failed'));
    });

    expect(result.current.isConnected).toBe(true); // Should still be connected until close
  });

  it.skip('should not broadcast when WebSocket is not open', () => {
    // Create a mock that doesn't auto-connect and doesn't call send
    class NonConnectingMockWebSocket {
      constructor(url) {
        this.url = url;
        this.readyState = WebSocket.CONNECTING; // Stay in CONNECTING state
        this.onopen = null;
        this.onclose = null;
        this.onmessage = null;
        this.onerror = null;
        this.sentMessages = [];
        // Don't set up auto-connection timeout
      }

      send(data) {
        // This should never be called because the hook checks readyState
        this.sentMessages.push(data);
      }

      close() {
        this.readyState = WebSocket.CLOSED;
        if (this.onclose) this.onclose();
      }
    }

    global.WebSocket = vi.fn().mockImplementation((url) => {
      mockWebSocket = new NonConnectingMockWebSocket(url);
      return mockWebSocket;
    });

    const { result } = renderHook(() =>
      useWebSocket('test-agent-id', 'test-client-id')
    );

    // Immediately try to broadcast while still CONNECTING
    // The hook should check readyState and not call send()
    act(() => {
      result.current.broadcastNodeChange('nodeUpdated', {
        nodeId: 'node-1',
        data: { position: { x: 100, y: 200 } },
      });
    });

    // Since readyState is CONNECTING (not OPEN), send should not be called
    expect(mockWebSocket.sentMessages).toHaveLength(0);
  });
});
