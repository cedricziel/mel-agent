import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { DraftAPI, AutoSaver, useAutoSaver } from '../draftClient';

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

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDraft),
      });

      const result = await DraftAPI.getDraft('workflow-123');

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/agents/workflow-123/draft'
      );
      expect(result).toEqual(mockDraft);
    });

    it('should handle API errors', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Not Found',
      });

      await expect(DraftAPI.getDraft('workflow-123')).rejects.toThrow(
        'Failed to get draft: Not Found'
      );
    });
  });

  describe('updateDraft', () => {
    it('should update draft successfully', async () => {
      const draftData = {
        nodes: [{ id: '1', type: 'test' }],
        edges: [],
      };

      const mockResponse = { success: true };
      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await DraftAPI.updateDraft('workflow-123', draftData);

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/agents/workflow-123/draft',
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(draftData),
        }
      );
      expect(result).toEqual(mockResponse);
    });

    it('should handle update errors', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Internal Server Error',
      });

      await expect(
        DraftAPI.updateDraft('workflow-123', { nodes: [], edges: [] })
      ).rejects.toThrow('Failed to update draft: Internal Server Error');
    });
  });

  describe('testDraftNode', () => {
    it('should test node successfully', async () => {
      const mockResult = {
        success: true,
        result: { data: 'test output' },
        node_id: 'node-1',
      };

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResult),
      });

      const result = await DraftAPI.testDraftNode('workflow-123', 'node-1', {
        input: 'test',
      });

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/agents/workflow-123/draft/nodes/node-1/test',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            node_id: 'node-1',
            test_data: { input: 'test' },
          }),
        }
      );
      expect(result).toEqual(mockResult);
    });

    it('should handle test errors', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Bad Request',
      });

      await expect(
        DraftAPI.testDraftNode('workflow-123', 'node-1', {})
      ).rejects.toThrow('Failed to test node: Bad Request');
    });
  });

  describe('deployVersion', () => {
    it('should deploy version successfully', async () => {
      const mockResponse = { success: true };
      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await DraftAPI.deployVersion(
        'workflow-123',
        1,
        'Deploy notes'
      );

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/agents/workflow-123/deploy',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            version: 1,
            notes: 'Deploy notes',
          }),
        }
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
    expect(autoSaver.agentId).toBe('workflow-123');
    expect(autoSaver.saveDelay).toBe(1000);
    expect(autoSaver.isSaving).toBe(false);
    expect(autoSaver.pendingChanges).toBe(null);
    expect(autoSaver.saveTimeout).toBe(null);
  });

  it('should schedule save with debouncing', async () => {
    const draftData = { nodes: [], edges: [] };

    // Mock fetch for the DraftAPI call
    global.fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    autoSaver.scheduleSave(draftData);
    autoSaver.scheduleSave(draftData); // Should debounce

    expect(autoSaver.isSaving).toBe(false);
    expect(autoSaver.pendingChanges).toEqual(draftData);

    // Reset fetch call count before advancing timers
    global.fetch.mockClear();

    // Fast-forward past the delay
    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/agents/workflow-123/draft',
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(draftData),
      }
    );
  });

  it('should save immediately when saveNow is called', async () => {
    const draftData = { nodes: [], edges: [] };

    global.fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // Clear any previous fetch calls
    global.fetch.mockClear();

    autoSaver.pendingChanges = draftData;
    await autoSaver.saveNow();

    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/agents/workflow-123/draft',
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(draftData),
      }
    );
  });

  it('should handle save errors', async () => {
    const draftData = { nodes: [], edges: [] };
    const onErrorSpy = vi.fn();

    global.fetch.mockResolvedValueOnce({
      ok: false,
      statusText: 'Save failed',
    });

    autoSaver = new AutoSaver('workflow-123', null, onErrorSpy);
    autoSaver.scheduleSave(draftData);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(onErrorSpy).toHaveBeenCalledWith(expect.any(Error));
  });

  it('should update save status correctly', async () => {
    const draftData = { nodes: [], edges: [] };
    const onSaveSpy = vi.fn();

    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    autoSaver = new AutoSaver('workflow-123', onSaveSpy);
    autoSaver.scheduleSave(draftData);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await vi.runAllTimersAsync();
    });

    expect(autoSaver.isSaving).toBe(false);
    expect(onSaveSpy).toHaveBeenCalled();
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
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
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
  });

  it('should cleanup AutoSaver on unmount', () => {
    const { unmount } = renderHook(() => useAutoSaver('workflow-123'));

    // Should not throw when unmounting
    expect(() => unmount()).not.toThrow();
  });
});
