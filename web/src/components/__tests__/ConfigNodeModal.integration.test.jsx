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

vi.mock('../../hooks/useNodeModalState', () => ({
  useNodeModalState: () => ({
    currentFormData: {},
    dynamicOptions: {},
    loadingOptions: false,
    credentials: [],
    handleChange: vi.fn(),
    inputData: {},
    outputData: {},
    setOutputData: vi.fn(),
    activeTab: 'config',
    setActiveTab: vi.fn(),
    loadNodeExecutionData: vi.fn(),
  }),
}));

vi.mock('../../hooks/useAutoLayout', () => ({
  useAutoLayout: () => ({
    handleAutoLayout: vi.fn(),
  }),
}));

vi.mock('../../hooks/useNodeTypes.jsx', () => ({
  useNodeTypes: () => ({
    nodeDefs: [
      {
        type: 'openai_model',
        label: 'OpenAI Model',
        icon: '🤖',
        category: 'AI',
        parameters: [
          { name: 'model', type: 'string', required: true },
          { name: 'temperature', type: 'number', required: false },
          { name: 'maxTokens', type: 'number', required: false },
        ],
      },
      {
        type: 'local_memory',
        label: 'Local Memory',
        icon: '🧠',
        category: 'Memory',
        parameters: [
          { name: 'maxMessages', type: 'number', required: false },
          { name: 'enableSummarization', type: 'boolean', required: false },
        ],
      },
    ],
    triggers: [],
    nodeTypes: ['openai_model', 'local_memory'],
    categories: ['AI', 'Memory'],
    triggersMap: {},
    refreshTriggers: vi.fn(),
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
              node_id: 'config-node-1',
              node_type: 'openai_model',
              position_x: 100,
              position_y: 100,
              config: {
                label: 'OpenAI Model',
                nodeTypeLabel: 'OpenAI Model',
                model: 'gpt-4',
                temperature: 0.7,
                maxTokens: 1000,
              },
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
          onClick={(e) => {
            if (onNodeClick) {
              onNodeClick(e, node);
            }
          }}
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

    // Wait for modal to open - look for the modal overlay
    await waitFor(() => {
      expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Check that modal content is visible
    const saveButtons = screen.getAllByRole('button', { name: 'Save' });
    expect(saveButtons).toHaveLength(2); // One on toolbar, one in modal
    
    // Find the modal save button (blue background, not disabled)
    const modalSaveButton = saveButtons.find(btn => 
      btn.classList.contains('bg-blue-600') && !btn.disabled
    );
    expect(modalSaveButton).toBeInTheDocument();
    
    expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument();
    
    // Check that we have the modal header with the node label
    expect(screen.getByRole('heading', { name: 'OpenAI Model' })).toBeInTheDocument();
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
      expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Should show parameter configuration options in the NodeConfigurationPanel
    // These should be visible as labels or field names
    expect(screen.getByText('Node Name')).toBeInTheDocument();
    
    // The parameter inputs should be rendered, even if labels aren't visible
    const inputs = screen.getAllByRole('textbox');
    expect(inputs.length).toBeGreaterThan(0);
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
      expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Close modal - get the close button specifically
    const closeButton = screen.getByRole('button', { name: 'Close' });
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
              node_id: 'memory-node-1',
              node_type: 'local_memory',
              position_x: 200,
              position_y: 100,
              config: {
                label: 'Local Memory',
                nodeTypeLabel: 'Local Memory',
                maxMessages: 100,
                enableSummarization: true,
              },
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
      expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Should show memory-specific configuration - check for the modal header
    expect(screen.getByRole('heading', { name: 'Local Memory' })).toBeInTheDocument();
  });
});
