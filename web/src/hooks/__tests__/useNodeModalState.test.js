import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNodeModalState } from '../useNodeModalState';

// Mock fetch globally
global.fetch = vi.fn();

describe('useNodeModalState', () => {
  const mockNode = {
    id: 'node-1',
    data: {
      label: 'Test Node',
      param1: 'value1',
    },
  };

  const mockNodeDef = {
    type: 'test-node',
    parameters: [
      {
        name: 'param1',
        type: 'string',
        required: true,
      },
    ],
  };

  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  it('should initialize with node data', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    expect(result.current.currentFormData).toEqual(mockNode.data);
    expect(result.current.activeTab).toBe('config');
    expect(result.current.inputData).toEqual({});
    expect(result.current.outputData).toEqual({});
  });

  it('should update form data when node data changes', () => {
    const { result, rerender } = renderHook(
      ({ node }) => useNodeModalState(node, mockNodeDef, mockOnChange),
      {
        initialProps: { node: mockNode },
      }
    );

    const updatedNode = {
      ...mockNode,
      data: { ...mockNode.data, param1: 'updated-value' },
    };

    rerender({ node: updatedNode });

    expect(result.current.currentFormData).toEqual(updatedNode.data);
  });

  it('should handle form field changes', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    act(() => {
      result.current.handleChange('param1', 'new-value');
    });

    expect(result.current.currentFormData.param1).toBe('new-value');
    expect(mockOnChange).toHaveBeenCalledWith('param1', 'new-value');
  });

  it('should update UI state', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    act(() => {
      result.current.setActiveTab('executions');
    });

    expect(result.current.activeTab).toBe('executions');

    act(() => {
      result.current.setInputData({ newInput: 'data' });
    });

    expect(result.current.inputData).toEqual({ newInput: 'data' });

    act(() => {
      result.current.setOutputData({ newOutput: 'data' });
    });

    expect(result.current.outputData).toEqual({ newOutput: 'data' });
  });

  it('should handle missing node gracefully', () => {
    const { result } = renderHook(() =>
      useNodeModalState(null, mockNodeDef, mockOnChange)
    );

    expect(result.current.currentFormData).toEqual({});
  });

  it('should provide loadDynamicOptionsForParam function', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    expect(typeof result.current.loadDynamicOptionsForParam).toBe('function');
  });

  it('should provide loadNodeExecutionData function', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    expect(typeof result.current.loadNodeExecutionData).toBe('function');
  });

  it('should initialize with empty loading states and credentials', () => {
    const { result } = renderHook(() =>
      useNodeModalState(mockNode, mockNodeDef, mockOnChange)
    );

    // dynamicOptions may have side effects from useEffect, so we just check it exists
    expect(typeof result.current.dynamicOptions).toBe('object');
    expect(result.current.loadingOptions).toEqual({});
    expect(result.current.credentials).toEqual({});
  });
});
