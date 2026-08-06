import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import BuilderPage from '../../pages/BuilderPage';

// Mock the generated API client
vi.mock('../../api/client', () => ({
  workflowsApi: {
    getWorkflowDraft: vi.fn(),
    updateWorkflowDraft: vi.fn(),
    testWorkflowDraftNode: vi.fn(),
    deployWorkflowVersion: vi.fn(),
    createWorkflow: vi.fn(),
    getWorkflow: vi.fn(),
    updateWorkflow: vi.fn(),
    deleteWorkflow: vi.fn(),
    listWorkflowNodes: vi.fn(),
    createWorkflowNode: vi.fn(),
    getWorkflowNode: vi.fn(),
    updateWorkflowNode: vi.fn(),
    deleteWorkflowNode: vi.fn(),
    listWorkflowEdges: vi.fn(),
    createWorkflowEdge: vi.fn(),
    deleteWorkflowEdge: vi.fn(),
    autoLayoutWorkflow: vi.fn(),
    listWorkflowVersions: vi.fn(),
    createWorkflowVersion: vi.fn(),
    getLatestWorkflowVersion: vi.fn(),
    executeWorkflow: vi.fn(),
  },
  workflowRunsApi: {
    listWorkflowRuns: vi.fn(),
    createWorkflowRun: vi.fn(),
    getWorkflowRun: vi.fn(),
  },
  nodeTypesApi: {
    listNodeTypes: vi.fn(),
  },
  triggersApi: {
    listTriggers: vi.fn(),
  },
}));

vi.mock('../../api/nodeTypesApi', () => ({
  nodeTypesApi: {
    getAllNodeTypes: vi.fn(),
    getNodeTypes: vi.fn(),
  },
}));
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

// Import the mocked APIs
import {
  workflowsApi,
  nodeTypesApi as generatedNodeTypesApi,
  triggersApi,
} from '../../api/client';
import { nodeTypesApi } from '../../api/nodeTypesApi';

// Helper function to setup default API mocks
const setupDefaultMocks = () => {
  // Setup nodeTypesApi mocks
  nodeTypesApi.getAllNodeTypes.mockResolvedValue([
    { type: 'agent', label: 'Agent', category: 'Core' },
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
  ]);

  generatedNodeTypesApi.listNodeTypes.mockResolvedValue({
    data: [{ type: 'agent', label: 'Agent', category: 'Core' }],
  });

  // Setup triggersApi mocks
  triggersApi.listTriggers.mockResolvedValue({ data: [] });

  // Setup workflowsApi mocks with successful data
  workflowsApi.getWorkflowDraft.mockResolvedValue({
    data: {
      workflow_id: 'test-agent',
      definition: {
        nodes: [
          {
            id: 'config-node-1',
            type: 'openai_model',
            position: { x: 100, y: 100 },
            data: {
              label: 'OpenAI Model',
              nodeTypeLabel: 'OpenAI Model',
              model: 'gpt-4',
              temperature: 0.7,
              maxTokens: 1000,
            },
          },
        ],
        edges: [],
      },
      updated_at: new Date().toISOString(),
    },
  });
  workflowsApi.updateWorkflowDraft.mockResolvedValue({ data: {} });
  workflowsApi.getWorkflow.mockResolvedValue({
    data: { id: 'test-agent', name: 'Test Agent' },
  });
  workflowsApi.listWorkflowNodes.mockResolvedValue({
    data: [
      {
        id: 'config-node-1',
        type: 'openai_model',
        position: { x: 100, y: 100 },
        data: {
          label: 'OpenAI Model',
          nodeTypeLabel: 'OpenAI Model',
          model: 'gpt-4',
          temperature: 0.7,
          maxTokens: 1000,
        },
      },
    ],
  });
  workflowsApi.listWorkflowEdges.mockResolvedValue({
    data: [],
  });
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
    setupDefaultMocks();
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

    // Check that modal content is visible - there should be at least one save button
    const saveButtons = screen.getAllByRole('button', { name: 'Save' });
    expect(saveButtons.length).toBeGreaterThanOrEqual(1);

    // Find a save button that's enabled
    const enabledSaveButton = saveButtons.find((btn) => !btn.disabled);
    expect(enabledSaveButton).toBeInTheDocument();

    expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument();

    // Check that we have the modal header with the node label
    expect(
      screen.getByRole('heading', { name: 'OpenAI Model' })
    ).toBeInTheDocument();
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
    workflowsApi.getWorkflowDraft.mockResolvedValue({
      data: {
        workflow_id: 'test-agent',
        definition: {
          nodes: [
            {
              id: 'memory-node-1',
              type: 'local_memory',
              position: { x: 200, y: 100 },
              data: {
                label: 'Local Memory',
                nodeTypeLabel: 'Local Memory',
                maxMessages: 100,
                enableSummarization: true,
              },
            },
          ],
          edges: [],
        },
        updated_at: new Date().toISOString(),
      },
    });

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
    expect(
      screen.getByRole('heading', { name: 'Local Memory' })
    ).toBeInTheDocument();
  });
});
