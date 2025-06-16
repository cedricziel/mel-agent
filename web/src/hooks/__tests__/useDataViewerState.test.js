import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import useDataViewerState from '../useDataViewerState';

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

describe('useDataViewerState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with default values', () => {
    const { result } = renderHook(() => useDataViewerState());

    expect(result.current.searchTerm).toBe('');
    expect(result.current.expandedPaths).toEqual(new Set(['root']));
    expect(result.current.searchable).toBe(true);
    expect(typeof result.current.setSearchTerm).toBe('function');
    expect(typeof result.current.toggleExpanded).toBe('function');
    expect(typeof result.current.shouldShowInSearch).toBe('function');
    expect(typeof result.current.copyToClipboard).toBe('function');
  });

  it('should initialize with searchable false', () => {
    const { result } = renderHook(() => useDataViewerState(false));

    expect(result.current.searchable).toBe(false);
  });

  it('should update search term', () => {
    const { result } = renderHook(() => useDataViewerState());

    act(() => {
      result.current.setSearchTerm('test search');
    });

    expect(result.current.searchTerm).toBe('test search');
  });

  it('should toggle expanded paths', () => {
    const { result } = renderHook(() => useDataViewerState());

    // Initially root is expanded
    expect(result.current.expandedPaths.has('root')).toBe(true);

    // Toggle root off
    act(() => {
      result.current.toggleExpanded('root');
    });

    expect(result.current.expandedPaths.has('root')).toBe(false);

    // Toggle root back on
    act(() => {
      result.current.toggleExpanded('root');
    });

    expect(result.current.expandedPaths.has('root')).toBe(true);

    // Add new path
    act(() => {
      result.current.toggleExpanded('root.child');
    });

    expect(result.current.expandedPaths.has('root.child')).toBe(true);
    expect(result.current.expandedPaths.has('root')).toBe(true);
  });

  it('should show all items when searchable is false', () => {
    const { result } = renderHook(() => useDataViewerState(false));

    const shouldShow = result.current.shouldShowInSearch(
      'key',
      'value',
      'path'
    );
    expect(shouldShow).toBe(true);
  });

  it('should show all items when search term is empty', () => {
    const { result } = renderHook(() => useDataViewerState(true));

    const shouldShow = result.current.shouldShowInSearch(
      'key',
      'value',
      'path'
    );
    expect(shouldShow).toBe(true);
  });

  it('should filter by key name', () => {
    const { result } = renderHook(() => useDataViewerState(true));

    act(() => {
      result.current.setSearchTerm('test');
    });

    expect(result.current.shouldShowInSearch('testKey', 'value', 'path')).toBe(
      true
    );
    expect(result.current.shouldShowInSearch('otherKey', 'value', 'path')).toBe(
      false
    );
  });

  it('should filter by string value', () => {
    const { result } = renderHook(() => useDataViewerState(true));

    act(() => {
      result.current.setSearchTerm('test');
    });

    expect(result.current.shouldShowInSearch('key', 'test value', 'path')).toBe(
      true
    );
    expect(
      result.current.shouldShowInSearch('key', 'other value', 'path')
    ).toBe(false);
    expect(result.current.shouldShowInSearch('key', 123, 'path')).toBe(false);
  });

  it('should filter by path', () => {
    const { result } = renderHook(() => useDataViewerState(true));

    act(() => {
      result.current.setSearchTerm('test');
    });

    expect(
      result.current.shouldShowInSearch('key', 'value', 'root.test.path')
    ).toBe(true);
    expect(
      result.current.shouldShowInSearch('key', 'value', 'root.other.path')
    ).toBe(false);
  });

  it('should be case insensitive in search', () => {
    const { result } = renderHook(() => useDataViewerState(true));

    act(() => {
      result.current.setSearchTerm('TEST');
    });

    expect(result.current.shouldShowInSearch('testkey', 'value', 'path')).toBe(
      true
    );
    expect(result.current.shouldShowInSearch('key', 'test value', 'path')).toBe(
      true
    );
    expect(
      result.current.shouldShowInSearch('key', 'value', 'root.test.path')
    ).toBe(true);
  });

  it('should copy to clipboard', async () => {
    const { result } = renderHook(() => useDataViewerState());

    const testValue = { key: 'value', nested: { data: 123 } };

    await act(async () => {
      result.current.copyToClipboard(testValue);
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      JSON.stringify(testValue, null, 2)
    );
  });

  it('should copy primitive values to clipboard', async () => {
    const { result } = renderHook(() => useDataViewerState());

    await act(async () => {
      result.current.copyToClipboard('simple string');
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      JSON.stringify('simple string', null, 2)
    );
  });
});
