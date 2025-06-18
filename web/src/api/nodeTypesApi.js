import axios from 'axios';

/**
 * API client for node types with filtering support
 */
export const nodeTypesApi = {
  /**
   * Get node types with optional filtering
   * @param {string|string[]} kinds - Filter by node kind (action, model, memory, trigger, tool)
   * @returns {Promise<Array>} Node type definitions
   */
  async getNodeTypes(kinds = null) {
    let url = '/api/node-types';

    if (kinds) {
      const kindFilter = Array.isArray(kinds) ? kinds.join(',') : kinds;
      url += `?kind=${encodeURIComponent(kindFilter)}`;
    }

    try {
      const response = await axios.get(url);
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
      const response = await axios.get('/api/node-types');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch all node types: ${error.message}`);
    }
  },
};
