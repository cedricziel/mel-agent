import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import BuilderPage from '../../pages/BuilderPage';

// Mock all the dependencies
vi.mock('axios', () => {
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  };
  
  return {
    default: {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      create: vi.fn(() => mockAxiosInstance),
    },
  };
});
vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    broadcastNodeChange: vi.fn(),
  }),
}));

vi.mock('../../hooks/useValidation', () => ({
  useValidation: () => ({
    validationErrors: {},
    validateWorkflow: vi.fn(),
    clearValidationErrors: vi.fn(),
  }),
}));

vi.mock('../../hooks/useAutoLayout', () => ({
  useAutoLayout: () => ({
    handleAutoLayout: vi.fn(),
  }),
}));

const mockedAxios = axios;

// Helper function to handle all URL patterns
const createMockAxiosImplementation = (customHandlers = {}) => {
  return (url) => {
    // Check for custom handlers first
    for (const [pattern, handler] of Object.entries(customHandlers)) {
      if (typeof pattern === 'string' && url.includes(pattern)) {
        return handler(url);
      } else if (pattern instanceof RegExp && pattern.test(url)) {
        return handler(url);
      }
    }

    // Default handlers
    if (url === '/api/node-types') {
      return Promise.resolve({
        data: [
          { type: 'agent', label: 'Agent', category: 'Core' },
          // Backend will eventually include config nodes, but for now they come from fallback
        ],
      });
    }
    if (url === '/api/triggers') {
      return Promise.resolve({ data: [] });
    }
    if (url.includes('/draft')) {
      return Promise.reject({ response: { status: 404 } });
    }
    if (url.includes('/workflows/')) {
      if (url.endsWith('/nodes')) {
        return Promise.resolve({
          data: [
            {
              id: 'config-node-1',
              type: 'openai_model',
              data: {
                label: 'OpenAI Model',
                nodeTypeLabel: 'OpenAI Model',
                model: 'gpt-4',
                temperature: 0.7,
                maxTokens: 1000,
              },
              position: { x: 100, y: 100 },
            },
          ],
        });
      }
      if (url.endsWith('/edges')) {
        return Promise.resolve({ data: [] });
      }
      return Promise.resolve({
        data: { id: 'test-agent', name: 'Test Agent' },
      });
    }
    return Promise.reject(new Error('Unmocked URL: ' + url));
  };
};

// Mock ReactFlow
vi.mock('reactflow', () => ({
  default: ({ children, onNodeClick, nodes }) => (
    <div data-testid="react-flow">
      {nodes.map((node) => (
        <div
          key={node.id}
          data-testid={`node-${node.type}-${node.id}`}
          onClick={(e) => onNodeClick(e, node)}
          className="cursor-pointer"
        >
          {node.data.label}
        </div>
      ))}
      {children}
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
}));

describe('Config Node Modal Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Set up default mock implementation for both direct axios and axios.create() instance
    const mockImplementation = createMockAxiosImplementation();
    mockedAxios.get.mockImplementation(mockImplementation);
    
    // Also mock the axios instance created by axios.create()
    const mockAxiosInstance = mockedAxios.create();
    mockAxiosInstance.get.mockImplementation(mockImplementation);
    mockAxiosInstance.post.mockImplementation(() => Promise.resolve({ data: {} }));
    mockAxiosInstance.put.mockImplementation(() => Promise.resolve({ data: {} }));
    mockAxiosInstance.delete.mockImplementation(() => Promise.resolve({ data: {} }));
  });

  it('should open modal when config node is clicked', async () => {
    render(<BuilderPage agentId="test-agent" />);

    // Wait for the component to load
    await waitFor(() => {
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    // Wait for nodes to be loaded and rendered
    await waitFor(() => {
      expect(
        screen.getByTestId('node-openai_model-config-node-1')
      ).toBeInTheDocument();
    });

    // Click on the config node
    const configNode = screen.getByTestId('node-openai_model-config-node-1');
    fireEvent.click(configNode);

    // Wait for modal to open
    await waitFor(() => {
      expect(screen.getByText('OpenAI Model')).toBeInTheDocument();
    });

    // Check that modal content is visible
    expect(screen.getByText('Configuration')).toBeInTheDocument();
    expect(screen.getByText('Save')).toBeInTheDocument();
    expect(screen.getByText('Close')).toBeInTheDocument();
  });

  it('should display config node parameters in modal', async () => {
    render(<BuilderPage agentId="test-agent" />);

    // Wait for loading and click config node
    await waitFor(() => {
      expect(
        screen.getByTestId('node-openai_model-config-node-1')
      ).toBeInTheDocument();
    });

    const configNode = screen.getByTestId('node-openai_model-config-node-1');
    fireEvent.click(configNode);

    // Wait for modal to open and check for parameter fields
    await waitFor(() => {
      expect(screen.getByText('OpenAI Model')).toBeInTheDocument();
    });

    // Should show parameter configuration options
    // (The exact field names depend on the NodeConfigurationPanel implementation)
    expect(
      screen.getByText(/Model|Temperature|Max Tokens/i)
    ).toBeInTheDocument();
  });

  it('should close modal when Close button is clicked', async () => {
    render(<BuilderPage agentId="test-agent" />);

    // Open modal
    await waitFor(() => {
      expect(
        screen.getByTestId('node-openai_model-config-node-1')
      ).toBeInTheDocument();
    });

    const configNode = screen.getByTestId('node-openai_model-config-node-1');
    fireEvent.click(configNode);

    await waitFor(() => {
      expect(screen.getByText('Close')).toBeInTheDocument();
    });

    // Close modal
    const closeButton = screen.getByText('Close');
    fireEvent.click(closeButton);

    // Modal should be closed
    await waitFor(() => {
      expect(screen.queryByText('Configuration')).not.toBeInTheDocument();
    });
  });

  it('should handle different config node types', async () => {
    // Set up custom mock for this test with memory node
    const customMockImplementation = createMockAxiosImplementation({
      '/nodes': () =>
        Promise.resolve({
          data: [
            {
              id: 'memory-node-1',
              type: 'local_memory',
              data: {
                label: 'Local Memory',
                nodeTypeLabel: 'Local Memory',
                maxMessages: 100,
                enableSummarization: true,
              },
              position: { x: 200, y: 100 },
            },
          ],
        }),
    });
    
    mockedAxios.get.mockImplementation(customMockImplementation);
    
    // Also update the axios instance created by axios.create()
    const mockAxiosInstance = mockedAxios.create();
    mockAxiosInstance.get.mockImplementation(customMockImplementation);

    render(<BuilderPage agentId="test-agent" />);

    await waitFor(() => {
      expect(
        screen.getByTestId('node-local_memory-memory-node-1')
      ).toBeInTheDocument();
    });

    const memoryNode = screen.getByTestId('node-local_memory-memory-node-1');
    fireEvent.click(memoryNode);

    await waitFor(() => {
      expect(screen.getByText('Local Memory')).toBeInTheDocument();
    });

    // Should show memory-specific configuration
    expect(screen.getByText('Configuration')).toBeInTheDocument();
  });
});
