import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useWorkflowState } from '../useWorkflowState'
import * as workflowClient from '../../api/workflowClient'
import * as draftClient from '../../api/draftClient'

// Mock the API modules
vi.mock('../../api/workflowClient', () => ({
  default: {
    loadWorkflowData: vi.fn(),
    createNode: vi.fn(),
    updateNode: vi.fn(),
    deleteNode: vi.fn(),
    createEdge: vi.fn(),
    deleteEdge: vi.fn(),
    toApiNode: vi.fn((node) => node),
    toApiEdge: vi.fn((edge) => edge),
    saveWorkflowVersion: vi.fn(),
  }
}))

vi.mock('../../api/draftClient', () => ({
  DraftAPI: {
    getDraft: vi.fn(),
    testDraftNode: vi.fn(),
    deployVersion: vi.fn(),
  },
  useAutoSaver: vi.fn(() => ({
    scheduleSave: vi.fn(),
    saveNow: vi.fn(),
    isSaving: false,
    lastSaved: null,
    saveError: null,
  }))
}))

describe('useWorkflowState', () => {
  const mockWorkflowId = 'test-workflow-id'
  
  beforeEach(() => {
    vi.clearAllMocks()
    
    // Default mock responses
    workflowClient.default.loadWorkflowData.mockResolvedValue({
      workflow: { id: mockWorkflowId, name: 'Test Workflow' },
      nodes: [],
      edges: []
    })
    
    draftClient.DraftAPI.getDraft.mockRejectedValue(new Error('No draft found'))
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should initialize with loading state', () => {
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    expect(result.current.loading).toBe(true)
    expect(result.current.nodes).toEqual([])
    expect(result.current.edges).toEqual([])
    expect(result.current.error).toBe(null)
  })

  it('should load workflow data successfully', async () => {
    const mockWorkflowData = {
      workflow: { id: mockWorkflowId, name: 'Test Workflow' },
      nodes: [{ id: '1', type: 'test', data: {} }],
      edges: [{ id: 'e1', source: '1', target: '2' }]
    }
    
    workflowClient.default.loadWorkflowData.mockResolvedValue(mockWorkflowData)
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    expect(result.current.workflow).toEqual(mockWorkflowData.workflow)
    expect(result.current.nodes).toEqual(mockWorkflowData.nodes)
    expect(result.current.edges).toEqual(mockWorkflowData.edges)
    expect(result.current.error).toBe(null)
  })

  it('should prefer draft over production version when available', async () => {
    const mockDraft = {
      nodes: [{ id: 'draft-1', type: 'test', data: {}, position: { x: 100, y: 100 } }],
      edges: [{ id: 'draft-e1', source: 'draft-1', target: 'draft-2' }]
    }
    
    draftClient.DraftAPI.getDraft.mockResolvedValue(mockDraft)
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    expect(result.current.isDraft).toBe(true)
    expect(result.current.nodes).toEqual(mockDraft.nodes)
    expect(result.current.edges).toEqual(mockDraft.edges)
    expect(workflowClient.default.loadWorkflowData).not.toHaveBeenCalled()
  })

  it('should create a node optimistically', async () => {
    workflowClient.default.createNode.mockResolvedValue({})
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    const newNode = { id: 'new-node', type: 'test', data: { label: 'Test' } }
    
    await act(async () => {
      await result.current.createNode(newNode)
    })
    
    expect(result.current.nodes).toContainEqual(newNode)
    expect(result.current.isDirty).toBe(true)
    expect(workflowClient.default.createNode).toHaveBeenCalledWith(
      mockWorkflowId,
      newNode
    )
  })

  it('should update a node optimistically', async () => {
    const initialNode = { id: 'node-1', type: 'test', data: { label: 'Initial' } }
    
    workflowClient.default.loadWorkflowData.mockResolvedValue({
      workflow: { id: mockWorkflowId, name: 'Test' },
      nodes: [initialNode],
      edges: []
    })
    
    workflowClient.default.updateNode.mockResolvedValue({})
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    const updates = { data: { label: 'Updated' } }
    
    await act(async () => {
      await result.current.updateNode('node-1', updates)
    })
    
    const updatedNode = result.current.nodes.find(n => n.id === 'node-1')
    expect(updatedNode.data.label).toBe('Updated')
    expect(result.current.isDirty).toBe(true)
  })

  it('should test a draft node when in draft mode', async () => {
    const mockDraft = {
      nodes: [{ id: 'node-1', type: 'test', data: {}, position: { x: 100, y: 100 } }],
      edges: []
    }
    const mockResult = { success: true, result: { data: 'test output' } }
    
    draftClient.DraftAPI.getDraft.mockResolvedValue(mockDraft)
    draftClient.DraftAPI.testDraftNode.mockResolvedValue(mockResult)
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    // Should already be in draft mode since we loaded a draft
    expect(result.current.isDraft).toBe(true)
    
    const testResult = await act(async () => {
      return await result.current.testDraftNode('node-1', { input: 'test' })
    })
    
    expect(testResult).toEqual(mockResult)
    expect(draftClient.DraftAPI.testDraftNode).toHaveBeenCalledWith(
      mockWorkflowId,
      'node-1',
      { input: 'test' }
    )
  })

  it('should handle errors gracefully', async () => {
    const error = new Error('Network error')
    workflowClient.default.loadWorkflowData.mockRejectedValue(error)
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    expect(result.current.error).toBe('Network error')
  })

  it('should rollback optimistic updates on failure', async () => {
    const initialNodes = [{ id: 'node-1', type: 'test', data: { label: 'Initial' } }]
    
    workflowClient.default.loadWorkflowData.mockResolvedValue({
      workflow: { id: mockWorkflowId, name: 'Test' },
      nodes: initialNodes,
      edges: []
    })
    
    workflowClient.default.createNode.mockRejectedValue(new Error('Create failed'))
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    const newNode = { id: 'new-node', type: 'test', data: { label: 'Test' } }
    
    await act(async () => {
      try {
        await result.current.createNode(newNode)
      } catch (err) {
        // Expected to fail
      }
    })
    
    // Should rollback to original state
    expect(result.current.nodes).toEqual(initialNodes)
    expect(result.current.error).toBe('Create failed')
  })

  it('should deploy draft successfully when in draft mode', async () => {
    const mockDraft = {
      nodes: [{ id: 'node-1', type: 'test', data: {}, position: { x: 100, y: 100 } }],
      edges: []
    }
    
    draftClient.DraftAPI.getDraft.mockResolvedValue(mockDraft)
    draftClient.DraftAPI.deployVersion.mockResolvedValue({ success: true })
    workflowClient.default.saveWorkflowVersion.mockResolvedValue({})
    
    const { result } = renderHook(() => useWorkflowState(mockWorkflowId))
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    
    // Should already be in draft mode since we loaded a draft
    expect(result.current.isDraft).toBe(true)
    
    await act(async () => {
      await result.current.deployDraft('Test deployment')
    })
    
    expect(result.current.isDraft).toBe(false)
    expect(workflowClient.default.saveWorkflowVersion).toHaveBeenCalled()
    expect(draftClient.DraftAPI.deployVersion).toHaveBeenCalledWith(
      mockWorkflowId,
      1,
      'Test deployment'
    )
  })
})