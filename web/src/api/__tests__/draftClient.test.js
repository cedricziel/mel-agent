import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { DraftAPI, AutoSaver, useAutoSaver } from '../draftClient';

// Mock the generated client
vi.mock('../client', () => ({
  workflowsApi: {
    getWorkflowDraft: vi.fn(),
    updateWorkflowDraft: vi.fn(),
    testWorkflowDraftNode: vi.fn(),
    deployWorkflowVersion: vi.fn(),
  },
}));

import { workflowsApi } from '../client';

describe('DraftAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getDraft', () => {
    it('should fetch draft successfully', async () => {
      const mockDraft = {
        nodes: [{ id: '1', type: 'test' }],
        edges: [{ id: 'e1', source: '1', target: '2' }],
      };

      workflowsApi.getWorkflowDraft.mockResolvedValueOnce({
        data: mockDraft,
      });

      const result = await DraftAPI.getDraft('workflow-123');

      expect(workflowsApi.getWorkflowDraft).toHaveBeenCalledWith(
        'workflow-123'
      );
      expect(result).toEqual(mockDraft);
    });

    it('should handle API errors', async () => {
      const consoleSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      workflowsApi.getWorkflowDraft.mockRejectedValueOnce(
        new Error('Not Found')
      );

      await expect(DraftAPI.getDraft('workflow-123')).rejects.toThrow(
        'Not Found'
      );

      consoleSpy.mockRestore();
    });
  });

  describe('updateDraft', () => {
    it('should update draft successfully', async () => {
      const draftData = {
        nodes: [{ id: '1', type: 'test' }],
        edges: [],
      };

      const mockResponse = { success: true };
      workflowsApi.updateWorkflowDraft.mockResolvedValueOnce({
        data: mockResponse,
      });

      const result = await DraftAPI.updateDraft('workflow-123', draftData);

      expect(workflowsApi.updateWorkflowDraft).toHaveBeenCalledWith(
        'workflow-123',
        { definition: draftData }
      );
      expect(result).toEqual(mockResponse);
    });

    it('should handle update errors', async () => {
      const consoleSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      workflowsApi.updateWorkflowDraft.mockRejectedValueOnce(
        new Error('Internal Server Error')
      );

      await expect(
        DraftAPI.updateDraft('workflow-123', { nodes: [], edges: [] })
      ).rejects.toThrow('Internal Server Error');

      consoleSpy.mockRestore();
    });
  });

  describe('testDraftNode', () => {
    it('should test node successfully', async () => {
      const mockResult = {
        success: true,
        result: { data: 'test output' },
        node_id: 'node-1',
      };

      workflowsApi.testWorkflowDraftNode.mockResolvedValueOnce({
        data: mockResult,
      });

      const result = await DraftAPI.testDraftNode('workflow-123', 'node-1', {
        input: 'test',
      });

      expect(workflowsApi.testWorkflowDraftNode).toHaveBeenCalledWith(
        'workflow-123',
        'node-1',
        { input: { input: 'test' } }
      );
      expect(result).toEqual(mockResult);
    });

    it('should handle test errors', async () => {
      const consoleSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      workflowsApi.testWorkflowDraftNode.mockRejectedValueOnce(
        new Error('Bad Request')
      );

      await expect(
        DraftAPI.testDraftNode('workflow-123', 'node-1', {})
      ).rejects.toThrow('Bad Request');

      consoleSpy.mockRestore();
    });
  });

  describe('deployVersion', () => {
    it('should deploy version successfully', async () => {
      const mockResponse = { success: true };
      workflowsApi.deployWorkflowVersion.mockResolvedValueOnce({
        data: mockResponse,
      });

      const result = await DraftAPI.deployVersion(
        'workflow-123',
        1,
        'Deploy notes'
      );

      expect(workflowsApi.deployWorkflowVersion).toHaveBeenCalledWith(
        'workflow-123',
        1
      );
      expect(result).toEqual(mockResponse);
    });
  });
});

describe('AutoSaver', () => {
  let autoSaver;

  beforeEach(() => {
    vi.useFakeTimers();
    autoSaver = new AutoSaver('workflow-123');
  });

  afterEach(() => {
    autoSaver.destroy();
    vi.useRealTimers();
  });

  it('should initialize with correct properties', () => {
    expect(autoSaver.workflowId).toBe('workflow-123');
    expect(autoSaver.saveDelay).toBe(1000);
    expect(autoSaver.isSaving).toBe(false);
    expect(autoSaver.pendingChanges).toBe(null);
    expect(autoSaver.saveTimeout).toBe(null);
  });

  it('should schedule save with debouncing', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    const draftData = { nodes: [], edges: [] };

    // Mock the workflowsApi for the DraftAPI call
    workflowsApi.updateWorkflowDraft.mockResolvedValue({
      data: {},
    });

    autoSaver.scheduleSave(draftData);
    autoSaver.scheduleSave(draftData); // Should debounce

    expect(autoSaver.isSaving).toBe(false);
    expect(autoSaver.pendingChanges).toEqual(draftData);

    // Reset workflowsApi call count before advancing timers
    workflowsApi.updateWorkflowDraft.mockClear();

    // Fast-forward past the delay
    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(workflowsApi.updateWorkflowDraft).toHaveBeenCalledTimes(1);
    expect(workflowsApi.updateWorkflowDraft).toHaveBeenCalledWith(
      'workflow-123',
      { definition: draftData }
    );

    consoleSpy.mockRestore();
  });

  it('should save immediately when saveNow is called', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    const draftData = { nodes: [], edges: [] };

    workflowsApi.updateWorkflowDraft.mockResolvedValue({
      data: {},
    });

    // Clear any previous API calls
    workflowsApi.updateWorkflowDraft.mockClear();

    autoSaver.pendingChanges = draftData;
    await autoSaver.saveNow();

    expect(workflowsApi.updateWorkflowDraft).toHaveBeenCalledTimes(1);
    expect(workflowsApi.updateWorkflowDraft).toHaveBeenCalledWith(
      'workflow-123',
      { definition: draftData }
    );

    consoleSpy.mockRestore();
  });

  it('should handle save errors', async () => {
    const consoleErrorSpy = vi
      .spyOn(console, 'error')
      .mockImplementation(() => {});
    const draftData = { nodes: [], edges: [] };
    const onErrorSpy = vi.fn();

    workflowsApi.updateWorkflowDraft.mockRejectedValueOnce(
      new Error('Save failed')
    );

    autoSaver = new AutoSaver('workflow-123', null, onErrorSpy);
    autoSaver.scheduleSave(draftData);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(onErrorSpy).toHaveBeenCalledWith(expect.any(Error));

    consoleErrorSpy.mockRestore();
  });

  it('should update save status correctly', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    const draftData = { nodes: [], edges: [] };
    const onSaveSpy = vi.fn();

    workflowsApi.updateWorkflowDraft.mockResolvedValueOnce({
      data: {},
    });

    autoSaver = new AutoSaver('workflow-123', onSaveSpy);
    autoSaver.scheduleSave(draftData);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(autoSaver.isSaving).toBe(false);
    expect(onSaveSpy).toHaveBeenCalled();

    consoleSpy.mockRestore();
  });
});

describe('useAutoSaver', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should return no-op functions when agentId is not provided', () => {
    const { result } = renderHook(() => useAutoSaver(null));

    expect(result.current.scheduleSave).toBeTypeOf('function');
    expect(result.current.saveNow).toBeTypeOf('function');
    expect(result.current.isSaving).toBe(false);
    expect(result.current.lastSaved).toBe(null);
    expect(result.current.saveError).toBe(null);
  });

  it('should create and manage AutoSaver instance', () => {
    const { result } = renderHook(() => useAutoSaver('workflow-123'));

    expect(result.current.scheduleSave).toBeTypeOf('function');
    expect(result.current.saveNow).toBeTypeOf('function');
    expect(result.current.isSaving).toBe(false);
    expect(result.current.lastSaved).toBe(null);
    expect(result.current.saveError).toBe(null);
  });

  it('should update state when AutoSaver state changes', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

    workflowsApi.updateWorkflowDraft.mockResolvedValueOnce({
      data: {},
    });

    const { result } = renderHook(() => useAutoSaver('workflow-123'));

    const draftData = { nodes: [], edges: [] };

    act(() => {
      result.current.scheduleSave(draftData);
    });

    expect(result.current.isSaving).toBe(true);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(result.current.lastSaved).toBeTruthy();
    expect(result.current.isSaving).toBe(false);

    consoleSpy.mockRestore();
  });

  it('should cleanup AutoSaver on unmount', () => {
    const { unmount } = renderHook(() => useAutoSaver('workflow-123'));

    // Should not throw when unmounting
    expect(() => unmount()).not.toThrow();
  });
});
