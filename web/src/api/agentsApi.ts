// TypeScript version of agents API client
import { agentsApi as generatedAgentsApi } from './client';
import type {
  Agent,
  CreateAgentRequest,
  UpdateAgentRequest,
  PaginatedResponse,
} from './types';

/**
 * Type-safe agents API client
 * Example of progressive TypeScript migration
 */
export class AgentsApiClient {
  /**
   * List all agents with optional pagination
   */
  async listAgents(params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<Agent>> {
    try {
      const response = await generatedAgentsApi.listAgents(
        params?.page,
        params?.limit
      );

      // Handle both paginated and direct array responses
      if (
        response.data &&
        typeof response.data === 'object' &&
        'agents' in response.data
      ) {
        return response.data as PaginatedResponse<Agent>;
      }

      // Fallback for direct array response
      return {
        data: Array.isArray(response.data) ? response.data : [],
        total: Array.isArray(response.data) ? response.data.length : 0,
        page: params?.page || 1,
        limit: params?.limit || 20,
      };
    } catch (error) {
      console.error('Failed to fetch agents:', error);
      throw new Error(
        `Failed to fetch agents: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Get a specific agent by ID
   */
  async getAgent(id: string): Promise<Agent> {
    try {
      const response = await generatedAgentsApi.getAgent(id);
      return response.data;
    } catch (error) {
      console.error(`Failed to fetch agent ${id}:`, error);
      throw new Error(
        `Failed to fetch agent: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Create a new agent
   */
  async createAgent(data: CreateAgentRequest): Promise<Agent> {
    try {
      const response = await generatedAgentsApi.createAgent(data);
      return response.data;
    } catch (error) {
      console.error('Failed to create agent:', error);
      throw new Error(
        `Failed to create agent: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Update an existing agent
   */
  async updateAgent(id: string, data: UpdateAgentRequest): Promise<Agent> {
    try {
      const response = await generatedAgentsApi.updateAgent(id, data);
      return response.data;
    } catch (error) {
      console.error(`Failed to update agent ${id}:`, error);
      throw new Error(
        `Failed to update agent: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Delete an agent
   */
  async deleteAgent(id: string): Promise<void> {
    try {
      await generatedAgentsApi.deleteAgent(id);
    } catch (error) {
      console.error(`Failed to delete agent ${id}:`, error);
      throw new Error(
        `Failed to delete agent: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }
}

// Export singleton instance for immediate use
export const agentsApiClient = new AgentsApiClient();

// Export for backward compatibility
export const agentsApi = agentsApiClient;
