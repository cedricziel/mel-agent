import { useState, useCallback, useEffect, useRef } from 'react';
import workflowClient from '../api/workflowClient';

export function useWorkflowState(workflowId) {
  const [workflow, setWorkflow] = useState(null);
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isDirty, setIsDirty] = useState(false);
  
  // Keep track of pending operations to avoid conflicts
  const pendingOperations = useRef(new Set());

  // Load initial workflow data
  const loadWorkflow = useCallback(async () => {
    if (!workflowId) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const { workflow, nodes, edges } = await workflowClient.loadWorkflowData(workflowId);
      
      setWorkflow(workflow);
      setNodes(nodes);
      setEdges(edges);
      setIsDirty(false);
    } catch (err) {
      console.error('Failed to load workflow:', err);
      setError(err.message || 'Failed to load workflow');
    } finally {
      setLoading(false);
    }
  }, [workflowId]);

  useEffect(() => {
    loadWorkflow();
  }, [loadWorkflow]);

  // Optimistic update helper
  const withOptimisticUpdate = useCallback(async (
    optimisticUpdate,
    apiCall,
    rollbackUpdate
  ) => {
    const operationId = Math.random().toString(36);
    pendingOperations.current.add(operationId);
    
    try {
      // Apply optimistic update immediately
      optimisticUpdate();
      setIsDirty(true);
      
      // Execute API call
      await apiCall();
      
      // Operation succeeded, no rollback needed
    } catch (err) {
      console.error('Operation failed, rolling back:', err);
      
      // Rollback optimistic update
      if (rollbackUpdate) {
        rollbackUpdate();
      }
      
      setError(err.message || 'Operation failed');
      throw err;
    } finally {
      pendingOperations.current.delete(operationId);
    }
  }, []);

  // Node operations
  const createNode = useCallback(async (nodeData) => {
    const apiNodeData = workflowClient.toApiNode(nodeData);
    
    await withOptimisticUpdate(
      () => setNodes(prev => [...prev, nodeData]),
      () => workflowClient.createNode(workflowId, apiNodeData),
      () => setNodes(prev => prev.filter(n => n.id !== nodeData.id))
    );
  }, [workflowId, withOptimisticUpdate]);

  const updateNode = useCallback(async (nodeId, updates) => {
    let originalNode = null;
    
    await withOptimisticUpdate(
      () => setNodes(prev => {
        const index = prev.findIndex(n => n.id === nodeId);
        if (index === -1) return prev;
        
        originalNode = prev[index];
        const newNodes = [...prev];
        newNodes[index] = { ...originalNode, ...updates };
        return newNodes;
      }),
      () => {
        // Convert updates to API format
        const apiUpdates = {};
        if (updates.position) {
          apiUpdates.position_x = updates.position.x;
          apiUpdates.position_y = updates.position.y;
        }
        if (updates.data) {
          apiUpdates.config = updates.data;
        }
        if (updates.type) {
          apiUpdates.node_type = updates.type;
        }
        
        return workflowClient.updateNode(workflowId, nodeId, apiUpdates);
      },
      () => {
        if (originalNode) {
          setNodes(prev => {
            const index = prev.findIndex(n => n.id === nodeId);
            if (index === -1) return prev;
            const newNodes = [...prev];
            newNodes[index] = originalNode;
            return newNodes;
          });
        }
      }
    );
  }, [workflowId, withOptimisticUpdate]);

  const deleteNode = useCallback(async (nodeId) => {
    let originalNode = null;
    
    await withOptimisticUpdate(
      () => setNodes(prev => {
        originalNode = prev.find(n => n.id === nodeId);
        return prev.filter(n => n.id !== nodeId);
      }),
      () => workflowClient.deleteNode(workflowId, nodeId),
      () => {
        if (originalNode) {
          setNodes(prev => [...prev, originalNode]);
        }
      }
    );
  }, [workflowId, withOptimisticUpdate]);

  // Edge operations
  const createEdge = useCallback(async (edgeData) => {
    const apiEdgeData = workflowClient.toApiEdge(edgeData);
    
    await withOptimisticUpdate(
      () => setEdges(prev => [...prev, edgeData]),
      () => workflowClient.createEdge(workflowId, apiEdgeData),
      () => setEdges(prev => prev.filter(e => e.id !== edgeData.id))
    );
  }, [workflowId, withOptimisticUpdate]);

  const deleteEdge = useCallback(async (edgeId) => {
    let originalEdge = null;
    
    await withOptimisticUpdate(
      () => setEdges(prev => {
        originalEdge = prev.find(e => e.id === edgeId);
        return prev.filter(e => e.id !== edgeId);
      }),
      () => workflowClient.deleteEdge(workflowId, edgeId),
      () => {
        if (originalEdge) {
          setEdges(prev => [...prev, originalEdge]);
        }
      }
    );
  }, [workflowId, withOptimisticUpdate]);

  // Workflow operations
  const updateWorkflow = useCallback(async (updates) => {
    let originalWorkflow = null;
    
    await withOptimisticUpdate(
      () => setWorkflow(prev => {
        originalWorkflow = prev;
        return { ...prev, ...updates };
      }),
      () => workflowClient.updateWorkflow(workflowId, updates),
      () => {
        if (originalWorkflow) {
          setWorkflow(originalWorkflow);
        }
      }
    );
  }, [workflowId, withOptimisticUpdate]);

  // Auto-layout
  const autoLayout = useCallback(async () => {
    try {
      await workflowClient.autoLayout(workflowId);
      // Reload nodes to get updated positions
      const newNodes = await workflowClient.getNodes(workflowId);
      setNodes(newNodes.map(node => workflowClient.toReactFlowNode(node)));
    } catch (err) {
      console.error('Auto-layout failed:', err);
      setError(err.message || 'Auto-layout failed');
    }
  }, [workflowId]);

  // Batch operations for ReactFlow compatibility
  const applyNodeChanges = useCallback((changes) => {
    changes.forEach(change => {
      switch (change.type) {
        case 'position':
          if (change.position) {
            updateNode(change.id, { position: change.position });
          }
          break;
        case 'remove':
          deleteNode(change.id);
          break;
        default:
          console.warn('Unhandled node change:', change);
      }
    });
  }, [updateNode, deleteNode]);

  const applyEdgeChanges = useCallback((changes) => {
    changes.forEach(change => {
      switch (change.type) {
        case 'remove':
          deleteEdge(change.id);
          break;
        default:
          console.warn('Unhandled edge change:', change);
      }
    });
  }, [deleteEdge]);

  // Save as version (backward compatibility)
  const saveVersion = useCallback(async () => {
    try {
      const graph = { nodes, edges };
      await workflowClient.saveWorkflowVersion(workflowId, graph);
      setIsDirty(false);
    } catch (err) {
      console.error('Failed to save version:', err);
      setError(err.message || 'Failed to save version');
      throw err;
    }
  }, [workflowId, nodes, edges]);

  return {
    // State
    workflow,
    nodes,
    edges,
    loading,
    error,
    isDirty,
    
    // Operations
    loadWorkflow,
    createNode,
    updateNode,
    deleteNode,
    createEdge,
    deleteEdge,
    updateWorkflow,
    autoLayout,
    
    // ReactFlow compatibility
    applyNodeChanges,
    applyEdgeChanges,
    
    // Legacy
    saveVersion,
    
    // Utilities
    clearError: () => setError(null)
  };
}