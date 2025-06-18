import axios from 'axios';

/**
 * API client for node types with filtering support
 */
export const nodeTypesApi = {
  /**
   * Get node types with optional filtering
   * @param {string|string[]} types - Filter by node type category (action, model, memory, trigger, tools)
   * @returns {Promise<Array>} Node type definitions
   */
  async getNodeTypes(types = null) {
    let url = '/api/node-types';

    if (types) {
      const typeFilter = Array.isArray(types) ? types.join(',') : types;
      url += `?type=${encodeURIComponent(typeFilter)}`;
    }

    const response = await axios.get(url);
    return response.data;
  },

  /**
   * Get all node types (convenience method)
   */
  async getAllNodeTypes() {
    const response = await axios.get('/api/node-types');
    return response.data;
  },
};
