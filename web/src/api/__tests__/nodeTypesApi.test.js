import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import { nodeTypesApi } from '../nodeTypesApi';

// Mock axios
vi.mock('axios');
const mockedAxios = vi.mocked(axios);

describe('nodeTypesApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getAllNodeTypes', () => {
    it('should fetch all node types and merge with config nodes', async () => {
      const backendNodeTypes = [
        { type: 'agent', label: 'Agent', category: 'Core' },
        { type: 'http_request', label: 'HTTP Request', category: 'Actions' },
      ];

      mockedAxios.get.mockResolvedValueOnce({
        data: backendNodeTypes,
      });

      const result = await nodeTypesApi.getAllNodeTypes();

      expect(mockedAxios.get).toHaveBeenCalledWith('/api/node-types');
      expect(result).toHaveLength(6); // 2 backend + 4 config nodes
      expect(result.some((node) => node.type === 'agent')).toBe(true);
      expect(result.some((node) => node.type === 'openai_model')).toBe(true);
      expect(result.some((node) => node.type === 'anthropic_model')).toBe(true);
      expect(result.some((node) => node.type === 'local_memory')).toBe(true);
      expect(result.some((node) => node.type === 'workflow_tools')).toBe(true);
    });

    it('should return fallback config nodes on API error', async () => {
      mockedAxios.get.mockRejectedValueOnce(new Error('API Error'));

      const result = await nodeTypesApi.getAllNodeTypes();

      expect(result).toHaveLength(4); // Only config nodes
      expect(result.every((node) => node.category === 'Configuration')).toBe(
        true
      );
    });
  });

  describe('getNodeTypes with filtering', () => {
    it('should call API with type filter for single type', async () => {
      const modelNodes = [{ type: 'openai_model', nodeTypeCategory: 'model' }];

      mockedAxios.get.mockResolvedValueOnce({
        data: modelNodes,
      });

      const result = await nodeTypesApi.getNodeTypes('model');

      expect(mockedAxios.get).toHaveBeenCalledWith(
        '/api/node-types?type=model'
      );
      expect(result).toEqual(modelNodes);
    });

    it('should call API with type filter for multiple types', async () => {
      const filteredNodes = [
        { type: 'agent', nodeTypeCategory: 'action' },
        { type: 'openai_model', nodeTypeCategory: 'model' },
      ];

      mockedAxios.get.mockResolvedValueOnce({
        data: filteredNodes,
      });

      const result = await nodeTypesApi.getNodeTypes(['action', 'model']);

      expect(mockedAxios.get).toHaveBeenCalledWith(
        '/api/node-types?type=action%2Cmodel'
      );
      expect(result).toEqual(filteredNodes);
    });

    it('should return filtered fallback nodes on API error', async () => {
      mockedAxios.get.mockRejectedValueOnce(new Error('API Error'));

      const result = await nodeTypesApi.getNodeTypes(['model']);

      expect(result).toHaveLength(2); // openai_model + anthropic_model
      expect(result.every((node) => node.nodeTypeCategory === 'model')).toBe(
        true
      );
    });
  });

  describe('config node definitions', () => {
    it('should have proper structure for openai_model', async () => {
      const result = await nodeTypesApi.getNodeTypes(['model']);
      const openaiModel = result.find((node) => node.type === 'openai_model');

      expect(openaiModel).toBeDefined();
      expect(openaiModel.label).toBe('OpenAI Model');
      expect(openaiModel.nodeTypeCategory).toBe('model');
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
      const result = await nodeTypesApi.getNodeTypes(['memory']);
      const localMemory = result.find((node) => node.type === 'local_memory');

      expect(localMemory).toBeDefined();
      expect(localMemory.label).toBe('Local Memory');
      expect(localMemory.nodeTypeCategory).toBe('memory');
      expect(localMemory.parameters).toBeDefined();
      expect(localMemory.parameters.some((p) => p.name === 'maxMessages')).toBe(
        true
      );
      expect(
        localMemory.parameters.some((p) => p.name === 'enableSummarization')
      ).toBe(true);
    });
  });
});
