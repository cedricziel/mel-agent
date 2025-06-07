import { useState, useCallback, useEffect, useRef } from 'react';
import workflowClient from '../api/workflowClient';
import { DraftAPI, useAutoSaver } from '../api/draftClient';

export function useWorkflowState(workflowId, enableAutoPersistence = true) {
  const [workflow, setWorkflow] = useState(null);
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isDirty, setIsDirty] = useState(false);
  const [isDraft, setIsDraft] = useState(true); // New: track if we're in draft mode
  
  // Keep track of pending operations to avoid conflicts
  const pendingOperations = useRef(new Set());
  
  // Auto-persistence for drafts
  const { 
    scheduleSave, 
    saveNow, 
    isSaving, 
    lastSaved, 
    saveError 
  } = useAutoSaver(enableAutoPersistence ? workflowId : null);

  // Helper to trigger auto-save of draft
  const triggerAutoSave = useCallback(() => {
    if (enableAutoPersistence && isDraft && scheduleSave) {
      const draftData = {
        nodes: nodes.map(node => ({
          id: node.id,
          type: node.type,
          position: node.position,
          data: node.data
        })),
        edges: edges.map(edge => ({
          id: edge.id,
          source: edge.source,
          target: edge.target,
          sourceHandle: edge.sourceHandle,
          targetHandle: edge.targetHandle
        }))
      };
      scheduleSave(draftData);
    }
  }, [enableAutoPersistence, isDraft, scheduleSave, nodes, edges]);

  // Load initial workflow data - prioritize draft over latest version
  const loadWorkflow = useCallback(async () => {
    if (!workflowId) return;
    
    try {
      setLoading(true);
      setError(null);
      
      let workflowData = null;
      let isDraftMode = true;
      
      // Try to load draft first
      if (enableAutoPersistence) {
        try {
          const draft = await DraftAPI.getDraft(workflowId);
          if (draft && (draft.nodes.length > 0 || draft.edges.length > 0)) {
            // Ensure draft nodes have proper ReactFlow format
            const formattedNodes = draft.nodes.map(node => ({
              ...node,
              position: node.position || { x: 100, y: 100 }, // Default position if missing
              data: node.data || {}
            }));
            
            workflowData = {
              workflow: { id: workflowId, name: 'Draft' },
              nodes: formattedNodes,
              edges: draft.edges
            };
            console.log('✅ Loaded draft with', draft.nodes.length, 'nodes');
          }
        } catch (draftErr) {
          console.log('No draft found, loading latest version:', draftErr.message);
        }
      }
      
      // Fall back to latest version if no draft
      if (!workflowData) {
        workflowData = await workflowClient.loadWorkflowData(workflowId);
        isDraftMode = false;
      }
      
      setWorkflow(workflowData.workflow);
      setNodes(workflowData.nodes);
      setEdges(workflowData.edges);
      setIsDraft(isDraftMode);
      setIsDirty(false);
      
      console.log(isDraftMode ? '📝 Draft mode active' : '🚀 Production mode active');
    } catch (err) {
      console.error('Failed to load workflow:', err);
      setError(err.message || 'Failed to load workflow');
    } finally {
      setLoading(false);
    }
  }, [workflowId, enableAutoPersistence]);

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
    
    // Trigger auto-save after successful creation
    triggerAutoSave();
  }, [workflowId, withOptimisticUpdate, triggerAutoSave]);

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
    
    // Trigger auto-save after successful update
    triggerAutoSave();
  }, [workflowId, withOptimisticUpdate, triggerAutoSave]);

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
    
    // Trigger auto-save after successful deletion
    triggerAutoSave();
  }, [workflowId, withOptimisticUpdate, triggerAutoSave]);

  // Edge operations
  const createEdge = useCallback(async (edgeData) => {
    const apiEdgeData = workflowClient.toApiEdge(edgeData);
    
    await withOptimisticUpdate(
      () => setEdges(prev => [...prev, edgeData]),
      () => workflowClient.createEdge(workflowId, apiEdgeData),
      () => setEdges(prev => prev.filter(e => e.id !== edgeData.id))
    );
    
    // Trigger auto-save after successful edge creation
    triggerAutoSave();
  }, [workflowId, withOptimisticUpdate, triggerAutoSave]);

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
    
    // Trigger auto-save after successful edge deletion
    triggerAutoSave();
  }, [workflowId, withOptimisticUpdate, triggerAutoSave]);

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

  // Test draft node functionality
  const testDraftNode = useCallback(async (nodeId, testData = {}) => {
    if (!isDraft) {
      throw new Error('Node testing is only available in draft mode');
    }
    
    try {
      const result = await DraftAPI.testDraftNode(workflowId, nodeId, testData);
      return result;
    } catch (err) {
      console.error('Failed to test draft node:', err);
      setError(err.message || 'Failed to test draft node');
      throw err;
    }
  }, [workflowId, isDraft]);

  // Deploy current draft as a new version
  const deployDraft = useCallback(async (notes = '') => {
    if (!isDraft) {
      throw new Error('Can only deploy from draft mode');
    }
    
    try {
      // First save current state as a version
      await saveVersion();
      
      // Then deploy that version
      const result = await DraftAPI.deployVersion(workflowId, 1, notes); // Assume version 1 for now
      
      setIsDraft(false);
      console.log('🚀 Draft deployed successfully');
      return result;
    } catch (err) {
      console.error('Failed to deploy draft:', err);
      setError(err.message || 'Failed to deploy draft');
      throw err;
    }
  }, [isDraft, workflowId, saveVersion]);

  return {
    // State
    workflow,
    nodes,
    edges,
    loading,
    error,
    isDirty,
    isDraft,
    
    // Auto-persistence state
    isSaving,
    lastSaved,
    saveError,
    
    // Operations
    loadWorkflow,
    createNode,
    updateNode,
    deleteNode,
    createEdge,
    deleteEdge,
    updateWorkflow,
    autoLayout,
    
    // Draft operations
    testDraftNode,
    deployDraft,
    saveNow,
    
    // ReactFlow compatibility
    applyNodeChanges,
    applyEdgeChanges,
    
    // Legacy
    saveVersion,
    
    // Utilities
    clearError: () => setError(null)
  };
}