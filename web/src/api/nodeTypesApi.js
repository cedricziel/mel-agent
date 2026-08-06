import { nodeTypesApi as generatedNodeTypesApi } from './client';

/**
 * API client for node types with filtering support
 * Uses the generated OpenAPI client
 */
export const nodeTypesApi = {
  /**
   * Get node types with optional filtering
   * @param {string|string[]} kinds - Filter by node kind (action, model, memory, trigger, tool)
   * @returns {Promise<Array>} Node type definitions
   */
  async getNodeTypes(kinds = null) {
    try {
      let kindFilter = null;
      if (kinds) {
        kindFilter = Array.isArray(kinds) ? kinds.join(',') : kinds;
      }

      const response = await generatedNodeTypesApi.listNodeTypes({
        kind: kindFilter,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch node types: ${error.message}`);
    }
  },

  /**
   * Get all node types (convenience method)
   */
  async getAllNodeTypes() {
    try {
      const response = await generatedNodeTypesApi.listNodeTypes();
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch all node types: ${error.message}`);
    }
  },
};
