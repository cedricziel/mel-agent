import { describe, it, expect, vi, beforeEach } from 'vitest';
import { nodeTypesApi } from '../nodeTypesApi';

// Mock the generated client
vi.mock('../client', () => ({
  nodeTypesApi: {
    listNodeTypes: vi.fn(),
  },
}));

import { nodeTypesApi as generatedNodeTypesApi } from '../client';

describe('nodeTypesApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getAllNodeTypes', () => {
    it('should fetch all node types from backend', async () => {
      const allNodeTypes = [
        { type: 'agent', label: 'Agent', category: 'Core' },
        { type: 'http_request', label: 'HTTP Request', category: 'Actions' },
        {
          type: 'openai_model',
          label: 'OpenAI Model',
          category: 'Configuration',
        },
        {
          type: 'anthropic_model',
          label: 'Anthropic Model',
          category: 'Configuration',
        },
        {
          type: 'local_memory',
          label: 'Local Memory',
          category: 'Configuration',
        },
        {
          type: 'workflow_tools',
          label: 'Workflow Tools',
          category: 'Configuration',
        },
      ];

      generatedNodeTypesApi.listNodeTypes.mockResolvedValueOnce({
        data: allNodeTypes,
      });

      const result = await nodeTypesApi.getAllNodeTypes();

      expect(generatedNodeTypesApi.listNodeTypes).toHaveBeenCalledWith();
      expect(result).toHaveLength(6);
      expect(result.some((node) => node.type === 'agent')).toBe(true);
      expect(result.some((node) => node.type === 'openai_model')).toBe(true);
      expect(result.some((node) => node.type === 'anthropic_model')).toBe(true);
      expect(result.some((node) => node.type === 'local_memory')).toBe(true);
      expect(result.some((node) => node.type === 'workflow_tools')).toBe(true);
    });

    it('should throw error on API failure', async () => {
      generatedNodeTypesApi.listNodeTypes.mockRejectedValueOnce(
        new Error('API Error')
      );

      await expect(nodeTypesApi.getAllNodeTypes()).rejects.toThrow(
        'Failed to fetch all node types: API Error'
      );
    });
  });

  describe('getNodeTypes with filtering', () => {
    it('should call API with kind filter for single kind', async () => {
      const modelNodes = [
        { type: 'openai_model', category: 'Configuration' },
        { type: 'anthropic_model', category: 'Configuration' },
      ];

      generatedNodeTypesApi.listNodeTypes.mockResolvedValueOnce({
        data: modelNodes,
      });

      const result = await nodeTypesApi.getNodeTypes('model');

      expect(generatedNodeTypesApi.listNodeTypes).toHaveBeenCalledWith({
        kind: 'model',
      });
      expect(result).toEqual(modelNodes);
    });

    it('should call API with kind filter for multiple kinds', async () => {
      const filteredNodes = [
        { type: 'agent', category: 'Core' },
        { type: 'openai_model', category: 'Configuration' },
      ];

      generatedNodeTypesApi.listNodeTypes.mockResolvedValueOnce({
        data: filteredNodes,
      });

      const result = await nodeTypesApi.getNodeTypes(['action', 'model']);

      expect(generatedNodeTypesApi.listNodeTypes).toHaveBeenCalledWith({
        kind: 'action,model',
      });
      expect(result).toEqual(filteredNodes);
    });

    it('should throw error on API failure', async () => {
      generatedNodeTypesApi.listNodeTypes.mockRejectedValueOnce(
        new Error('API Error')
      );

      await expect(nodeTypesApi.getNodeTypes(['model'])).rejects.toThrow(
        'Failed to fetch node types: API Error'
      );
    });
  });

  describe('config node definitions', () => {
    it('should have proper structure for openai_model', async () => {
      const modelNodes = [
        {
          type: 'openai_model',
          label: 'OpenAI Model',
          category: 'Configuration',
          parameters: [
            { name: 'model', type: 'enum' },
            { name: 'temperature', type: 'number' },
            { name: 'maxTokens', type: 'integer' },
            { name: 'credential', type: 'credential' },
          ],
        },
      ];

      generatedNodeTypesApi.listNodeTypes.mockResolvedValueOnce({
        data: modelNodes,
      });

      const result = await nodeTypesApi.getNodeTypes(['model']);
      const openaiModel = result.find((node) => node.type === 'openai_model');

      expect(openaiModel).toBeDefined();
      expect(openaiModel.label).toBe('OpenAI Model');
      expect(openaiModel.category).toBe('Configuration');
      expect(openaiModel.parameters).toBeDefined();
      expect(openaiModel.parameters.some((p) => p.name === 'model')).toBe(true);
      expect(openaiModel.parameters.some((p) => p.name === 'temperature')).toBe(
        true
      );
      expect(openaiModel.parameters.some((p) => p.name === 'maxTokens')).toBe(
        true
      );
    });

    it('should have proper structure for local_memory', async () => {
      const memoryNodes = [
        {
          type: 'local_memory',
          label: 'Local Memory',
          category: 'Configuration',
          parameters: [
            { name: 'storageType', type: 'string' },
            { name: 'namespace', type: 'string' },
            { name: 'maxEntries', type: 'integer' },
            { name: 'persistent', type: 'boolean' },
          ],
        },
      ];

      generatedNodeTypesApi.listNodeTypes.mockResolvedValueOnce({
        data: memoryNodes,
      });

      const result = await nodeTypesApi.getNodeTypes(['memory']);
      const localMemory = result.find((node) => node.type === 'local_memory');

      expect(localMemory).toBeDefined();
      expect(localMemory.label).toBe('Local Memory');
      expect(localMemory.category).toBe('Configuration');
      expect(localMemory.parameters).toBeDefined();
      expect(localMemory.parameters.some((p) => p.name === 'storageType')).toBe(
        true
      );
      expect(localMemory.parameters.some((p) => p.name === 'persistent')).toBe(
        true
      );
    });
  });
});
