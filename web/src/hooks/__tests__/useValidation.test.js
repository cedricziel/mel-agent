import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useValidation } from '../useValidation';

describe('useValidation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with empty validation errors', () => {
    const { result } = renderHook(() => useValidation());

    expect(result.current.validationErrors).toEqual({});
    expect(result.current.hasValidationErrors()).toBe(false);
    expect(result.current.getNodeValidationErrors('any-node')).toEqual([]);
  });

  it('should validate HTTP request nodes with missing URL', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Test HTTP Node',
          url: '', // Missing URL
          method: 'GET',
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    expect(result.current.validationErrors['node-1']).toContain(
      'Node "Test HTTP Node" is missing a URL'
    );
    expect(result.current.hasValidationErrors()).toBe(true);
    expect(result.current.getNodeValidationErrors('node-1')).toContain(
      'Node "Test HTTP Node" is missing a URL'
    );
  });

  it('should validate HTTP request nodes with missing method', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Test HTTP Node',
          url: 'https://api.example.com',
          method: '', // Missing method
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    expect(result.current.validationErrors['node-1']).toContain(
      'Node "Test HTTP Node" is missing a method'
    );
  });

  it('should validate HTTP request nodes with both URL and method missing', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Test HTTP Node',
          url: '',
          method: '',
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    const errors = result.current.validationErrors['node-1'];
    expect(errors).toContain('Node "Test HTTP Node" is missing a URL');
    expect(errors).toContain('Node "Test HTTP Node" is missing a method');
    expect(errors).toHaveLength(2);
  });

  it('should pass validation for valid HTTP request nodes', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Test HTTP Node',
          url: 'https://api.example.com',
          method: 'GET',
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(true);
    });

    expect(result.current.validationErrors).toEqual({});
    expect(result.current.hasValidationErrors()).toBe(false);
  });

  it('should validate async webhook nodes without http_response', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Async Webhook',
          mode: 'async',
        },
      },
      {
        id: 'other-node',
        type: 'http_request',
        data: {
          label: 'Other Node',
        },
      },
    ];

    const edges = [
      {
        source: 'webhook-1',
        target: 'other-node',
      },
    ];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    expect(result.current.validationErrors['webhook-1']).toContain(
      'Async Webhook node "Async Webhook" must be followed by a Webhook Response node'
    );
  });

  it('should pass validation for async webhook with http_response', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Async Webhook',
          mode: 'async',
        },
      },
      {
        id: 'response-1',
        type: 'http_response',
        data: {
          label: 'Response Node',
        },
      },
    ];

    const edges = [
      {
        source: 'webhook-1',
        target: 'response-1',
      },
    ];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(true);
    });

    expect(result.current.validationErrors).toEqual({});
  });

  it('should validate async webhook with indirect path to http_response', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Async Webhook',
          mode: 'async',
        },
      },
      {
        id: 'middle-node',
        type: 'http_request',
        data: {
          label: 'Middle Node',
          url: 'https://api.example.com',
          method: 'GET',
        },
      },
      {
        id: 'response-1',
        type: 'http_response',
        data: {
          label: 'Response Node',
        },
      },
    ];

    const edges = [
      {
        source: 'webhook-1',
        target: 'middle-node',
      },
      {
        source: 'middle-node',
        target: 'response-1',
      },
    ];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(true);
    });

    expect(result.current.validationErrors).toEqual({});
  });

  it('should ignore non-async webhook nodes', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Sync Webhook',
          mode: 'sync', // Not async
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(true);
    });

    expect(result.current.validationErrors).toEqual({});
  });

  it('should handle nodes without labels gracefully', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          // No label
          url: '',
          method: '',
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    const errors = result.current.validationErrors['node-1'];
    expect(errors).toContain('Node "node-1" is missing a URL');
    expect(errors).toContain('Node "node-1" is missing a method');
  });

  it('should clear validation errors', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Test Node',
          url: '',
          method: '',
        },
      },
    ];

    const edges = [];

    // First, create some validation errors
    act(() => {
      result.current.validateWorkflow(nodes, edges);
    });

    expect(result.current.hasValidationErrors()).toBe(true);

    // Then clear them
    act(() => {
      result.current.clearValidationErrors();
    });

    expect(result.current.validationErrors).toEqual({});
    expect(result.current.hasValidationErrors()).toBe(false);
  });

  it('should handle multiple validation errors across different nodes', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'http-1',
        type: 'http_request',
        data: {
          label: 'HTTP Node 1',
          url: '',
          method: 'GET',
        },
      },
      {
        id: 'http-2',
        type: 'http_request',
        data: {
          label: 'HTTP Node 2',
          url: 'https://api.example.com',
          method: '',
        },
      },
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Async Webhook',
          mode: 'async',
        },
      },
    ];

    const edges = [];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    expect(result.current.validationErrors['http-1']).toContain(
      'Node "HTTP Node 1" is missing a URL'
    );
    expect(result.current.validationErrors['http-2']).toContain(
      'Node "HTTP Node 2" is missing a method'
    );
    expect(result.current.validationErrors['webhook-1']).toContain(
      'Async Webhook node "Async Webhook" must be followed by a Webhook Response node'
    );

    expect(Object.keys(result.current.validationErrors)).toHaveLength(3);
  });

  it('should handle circular references in webhook validation', () => {
    const { result } = renderHook(() => useValidation());

    const nodes = [
      {
        id: 'webhook-1',
        type: 'webhook',
        data: {
          label: 'Async Webhook',
          mode: 'async',
        },
      },
      {
        id: 'node-1',
        type: 'http_request',
        data: {
          label: 'Node 1',
        },
      },
      {
        id: 'node-2',
        type: 'http_request',
        data: {
          label: 'Node 2',
        },
      },
    ];

    const edges = [
      {
        source: 'webhook-1',
        target: 'node-1',
      },
      {
        source: 'node-1',
        target: 'node-2',
      },
      {
        source: 'node-2',
        target: 'node-1', // Circular reference
      },
    ];

    act(() => {
      const isValid = result.current.validateWorkflow(nodes, edges);
      expect(isValid).toBe(false);
    });

    // Should still detect that there's no http_response node
    expect(result.current.validationErrors['webhook-1']).toContain(
      'Async Webhook node "Async Webhook" must be followed by a Webhook Response node'
    );
  });
});
