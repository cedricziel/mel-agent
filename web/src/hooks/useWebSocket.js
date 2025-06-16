import { useEffect, useRef, useCallback, useState, useMemo } from 'react';

/**
 * Custom hook for managing WebSocket connections in the workflow builder
 * Handles collaborative editing, real-time updates, and node execution status
 */
export function useWebSocket(agentId, clientId, callbacks = {}) {
  const {
    onNodeUpdated,
    onNodeCreated,
    onNodeDeleted,
    onEdgeCreated,
    onEdgeDeleted,
    onNodeExecution,
  } = callbacks;

  const wsRef = useRef(null);
  const execTimersRef = useRef({});
  const [isConnected, setIsConnected] = useState(false);

  // Generate client ID if not provided
  const finalClientId = useMemo(
    () => clientId || crypto.randomUUID(),
    [clientId]
  );

  // Broadcast node changes to other clients
  const broadcastNodeChange = useCallback(
    (type, data) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            clientId: finalClientId,
            type,
            workflowId: agentId,
            ...data,
          })
        );
      }
    },
    [finalClientId, agentId]
  );

  // WebSocket setup for collaborative editing
  useEffect(() => {
    if (!agentId) return;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/ws/agents/${agentId}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);

        // Ignore messages from the same client
        if (msg.clientId === finalClientId) return;

        switch (msg.type) {
          case 'nodeUpdated':
            // Handle individual node updates from other clients
            if (msg.workflowId === agentId && onNodeUpdated) {
              onNodeUpdated(msg.nodeId, msg.data);
            }
            break;

          case 'nodeCreated':
            if (msg.workflowId === agentId && onNodeCreated) {
              onNodeCreated(msg.node);
            }
            break;

          case 'nodeDeleted':
            if (msg.workflowId === agentId && onNodeDeleted) {
              onNodeDeleted(msg.nodeId);
            }
            break;

          case 'edgeCreated':
            if (msg.workflowId === agentId && onEdgeCreated) {
              onEdgeCreated(msg.edge);
            }
            break;

          case 'edgeDeleted':
            if (msg.workflowId === agentId && onEdgeDeleted) {
              onEdgeDeleted(msg.edgeId);
            }
            break;

          case 'nodeExecution': {
            // Runtime status updates for Live mode
            const { nodeId, phase } = msg;
            if (onNodeExecution) {
              onNodeExecution(nodeId, phase);
            }

            // Handle execution timing for visual feedback
            if (phase === 'start') {
              execTimersRef.current[nodeId] = {
                start: Date.now(),
                timeoutId: null,
              };
            } else if (phase === 'end') {
              const timer = execTimersRef.current[nodeId];
              const now = Date.now();

              if (timer) {
                const elapsed = now - timer.start;
                const remaining = 500 - elapsed; // Minimum 500ms display time

                if (timer.timeoutId) clearTimeout(timer.timeoutId);

                if (remaining <= 0) {
                  delete execTimersRef.current[nodeId];
                } else {
                  const tid = setTimeout(() => {
                    delete execTimersRef.current[nodeId];
                  }, remaining);
                  execTimersRef.current[nodeId].timeoutId = tid;
                }
              }
            }
            break;
          }

          default:
            console.warn('Unknown WebSocket message type:', msg.type);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      wsRef.current = null;
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    // Cleanup function
    return () => {
      // Clear all execution timers
      Object.values(execTimersRef.current).forEach((timer) => {
        if (timer.timeoutId) clearTimeout(timer.timeoutId);
      });
      execTimersRef.current = {};

      // Close WebSocket connection
      if (
        ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING
      ) {
        ws.close();
      }
      wsRef.current = null;
      setIsConnected(false);
    };
  }, [
    agentId,
    finalClientId,
    onNodeUpdated,
    onNodeCreated,
    onNodeDeleted,
    onEdgeCreated,
    onEdgeDeleted,
    onNodeExecution,
  ]);

  return {
    isConnected,
    broadcastNodeChange,
    execTimers: execTimersRef.current,
  };
}
