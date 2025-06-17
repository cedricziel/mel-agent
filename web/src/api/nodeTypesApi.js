import axios from 'axios';

/**
 * API client for node types with filtering support
 *
 * BACKEND TODO: Update /api/node-types endpoint to:
 * 1. Return ALL node types (actions, models, memory, triggers, tools)
 * 2. Support filtering: GET /api/node-types?type=action,model,memory
 * 3. Each node type should have a 'nodeTypeCategory' field for filtering
 *
 * Once backend is updated, remove getTemporaryConfigNodes() function
 */
export const nodeTypesApi = {
  /**
   * Get node types with optional filtering
   * @param {string|string[]} types - Filter by node type category (action, model, memory, trigger)
   * @returns {Promise<Array>} Node type definitions
   */
  async getNodeTypes(types = null) {
    try {
      let url = '/api/node-types';

      if (types) {
        const typeFilter = Array.isArray(types) ? types.join(',') : types;
        url += `?type=${encodeURIComponent(typeFilter)}`;
      }

      const response = await axios.get(url);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch node types:', error);

      // Fallback to temporary config nodes if backend doesn't support them yet
      if (
        types &&
        (types.includes('model') ||
          types.includes('memory') ||
          types.includes('tools'))
      ) {
        return getTemporaryConfigNodes().filter((node) => {
          const category = node.nodeTypeCategory;
          return types.includes(category);
        });
      }

      throw error;
    }
  },

  /**
   * Get all node types (convenience method)
   */
  async getAllNodeTypes() {
    try {
      const response = await axios.get('/api/node-types');

      // Until backend includes config nodes, merge them
      const backendNodes = response.data;
      const configNodes = getTemporaryConfigNodes();

      return [...backendNodes, ...configNodes];
    } catch (error) {
      console.error('Failed to fetch all node types:', error);
      return getTemporaryConfigNodes();
    }
  },
};

/**
 * Temporary config node definitions until backend is updated
 * TODO: Remove this when backend supports all node types
 */
function getTemporaryConfigNodes() {
  return [
    {
      type: 'openai_model',
      label: 'OpenAI Model',
      category: 'Configuration',
      nodeTypeCategory: 'model',
      icon: '🤖',
      parameters: [
        {
          name: 'model',
          label: 'Model',
          type: 'enum',
          required: true,
          options: ['gpt-4', 'gpt-3.5-turbo', 'gpt-4-turbo'],
          default: 'gpt-4',
        },
        {
          name: 'temperature',
          label: 'Temperature',
          type: 'number',
          required: false,
          default: 0.7,
          description: 'Controls randomness in output',
        },
        {
          name: 'maxTokens',
          label: 'Max Tokens',
          type: 'integer',
          required: false,
          default: 1000,
          description: 'Maximum number of tokens to generate',
        },
      ],
    },
    {
      type: 'anthropic_model',
      label: 'Anthropic Model',
      category: 'Configuration',
      nodeTypeCategory: 'model',
      icon: '🧠',
      parameters: [
        {
          name: 'model',
          label: 'Model',
          type: 'enum',
          required: true,
          options: [
            'claude-3-5-sonnet-20241022',
            'claude-3-haiku-20240307',
            'claude-3-opus-20240229',
          ],
          default: 'claude-3-5-sonnet-20241022',
        },
        {
          name: 'temperature',
          label: 'Temperature',
          type: 'number',
          required: false,
          default: 0.7,
          description: 'Controls randomness in output',
        },
      ],
    },
    {
      type: 'local_memory',
      label: 'Local Memory',
      category: 'Configuration',
      nodeTypeCategory: 'memory',
      icon: '💾',
      parameters: [
        {
          name: 'maxMessages',
          label: 'Max Messages',
          type: 'integer',
          required: false,
          default: 100,
          description: 'Maximum number of messages to remember',
        },
        {
          name: 'enableSummarization',
          label: 'Enable Summarization',
          type: 'boolean',
          required: false,
          default: true,
          description: 'Whether to summarize old messages',
        },
      ],
    },
    {
      type: 'workflow_tools',
      label: 'Workflow Tools',
      category: 'Configuration',
      nodeTypeCategory: 'tools',
      icon: '🔧',
      parameters: [
        {
          name: 'enabledTools',
          label: 'Enabled Tools',
          type: 'array',
          required: false,
          default: [],
          description: 'List of enabled tools',
        },
      ],
    },
  ];
}
