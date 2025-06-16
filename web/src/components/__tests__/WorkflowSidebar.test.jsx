import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import WorkflowSidebar from '../WorkflowSidebar';

// Mock the child components
vi.mock('../NodeDetailsPanel', () => ({
  default: ({ node, onChange }) => (
    <div data-testid="node-details-panel">
      NodeDetailsPanel for {node?.id}
      <button onClick={() => onChange('test', 'value')}>Change</button>
    </div>
  ),
}));

vi.mock('../ChatAssistant', () => ({
  default: ({ onClose, agentId }) => (
    <div data-testid="chat-assistant">
      ChatAssistant for {agentId}
      <button onClick={onClose}>Close</button>
    </div>
  ),
}));

describe('WorkflowSidebar', () => {
  const defaultProps = {
    isVisible: true,
    sidebarTab: null,
    viewMode: 'editor',
    selectedNode: null,
    selectedNodeDef: null,
    selectedExecution: null,
    executions: [],
    loadingExecutions: false,
    isLiveMode: false,
    isDraft: false,
    agentId: 'test-agent',
    triggersMap: {},
    onExecutionSelect: vi.fn(),
    onNodeChange: vi.fn(),
    onNodeExecute: vi.fn(),
    onChatClose: vi.fn(),
    onAddNode: vi.fn(),
    onConnectNodes: vi.fn(),
    onGetWorkflow: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when not visible', () => {
    render(<WorkflowSidebar {...defaultProps} isVisible={false} />);
    expect(screen.queryByTestId('workflow-sidebar')).not.toBeInTheDocument();
  });

  it('should render executions panel when in executions view mode', () => {
    render(<WorkflowSidebar {...defaultProps} viewMode="executions" />);

    expect(screen.getByText('Executions')).toBeInTheDocument();
    expect(
      screen.getByText(
        'No executions found. Run your workflow to see execution history.'
      )
    ).toBeInTheDocument();
  });

  it('should show loading state for executions', () => {
    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        loadingExecutions={true}
      />
    );

    expect(screen.getByText('Loading executions...')).toBeInTheDocument();
  });

  it('should render execution list when executions are available', () => {
    const executions = [
      {
        id: 'exec-1',
        created_at: '2023-01-01T12:00:00Z',
      },
      {
        id: 'exec-2',
        created_at: '2023-01-02T12:00:00Z',
      },
    ];

    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        executions={executions}
      />
    );

    expect(screen.getByText(/1\/1\/2023/)).toBeInTheDocument();
    expect(screen.getByText(/1\/2\/2023/)).toBeInTheDocument();
    expect(screen.getAllByText('Completed')).toHaveLength(2);
  });

  it('should call onExecutionSelect when execution is clicked', () => {
    const executions = [
      {
        id: 'exec-1',
        created_at: '2023-01-01T12:00:00Z',
      },
    ];

    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        executions={executions}
      />
    );

    fireEvent.click(screen.getByText(/1\/1\/2023/));
    expect(defaultProps.onExecutionSelect).toHaveBeenCalledWith(executions[0]);
  });

  it('should highlight selected execution', () => {
    const executions = [
      {
        id: 'exec-1',
        created_at: '2023-01-01T12:00:00Z',
      },
    ];

    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        executions={executions}
        selectedExecution={executions[0]}
      />
    );

    const executionElement = screen
      .getByText(/1\/1\/2023/)
      .closest('div').parentElement;
    expect(executionElement).toHaveClass('border-blue-500', 'bg-blue-50');
  });

  it('should render NodeDetailsPanel when details tab is active', () => {
    const selectedNode = { id: 'node-1', data: { label: 'Test Node' } };
    const selectedNodeDef = { type: 'default', label: 'Default Node' };

    render(
      <WorkflowSidebar
        {...defaultProps}
        sidebarTab="details"
        selectedNode={selectedNode}
        selectedNodeDef={selectedNodeDef}
      />
    );

    expect(screen.getByTestId('node-details-panel')).toBeInTheDocument();
    expect(screen.getByText('NodeDetailsPanel for node-1')).toBeInTheDocument();
  });

  it('should call onNodeChange when NodeDetailsPanel triggers change', () => {
    const selectedNode = { id: 'node-1', data: { label: 'Test Node' } };
    const selectedNodeDef = { type: 'default', label: 'Default Node' };

    render(
      <WorkflowSidebar
        {...defaultProps}
        sidebarTab="details"
        selectedNode={selectedNode}
        selectedNodeDef={selectedNodeDef}
      />
    );

    fireEvent.click(screen.getByText('Change'));
    expect(defaultProps.onNodeChange).toHaveBeenCalledWith('test', 'value');
  });

  it('should render ChatAssistant when chat tab is active', () => {
    render(<WorkflowSidebar {...defaultProps} sidebarTab="chat" />);

    expect(screen.getByTestId('chat-assistant')).toBeInTheDocument();
    expect(
      screen.getByText('ChatAssistant for test-agent')
    ).toBeInTheDocument();
  });

  it('should call onChatClose when ChatAssistant triggers close', () => {
    render(<WorkflowSidebar {...defaultProps} sidebarTab="chat" />);

    fireEvent.click(screen.getByText('Close'));
    expect(defaultProps.onChatClose).toHaveBeenCalled();
  });

  it('should not render NodeDetailsPanel when node or nodeDef is missing', () => {
    render(
      <WorkflowSidebar
        {...defaultProps}
        sidebarTab="details"
        selectedNode={null}
        selectedNodeDef={null}
      />
    );

    expect(screen.queryByTestId('node-details-panel')).not.toBeInTheDocument();
  });

  it('should not render NodeDetailsPanel in executions view mode', () => {
    const selectedNode = { id: 'node-1', data: { label: 'Test Node' } };
    const selectedNodeDef = { type: 'default', label: 'Default Node' };

    render(
      <WorkflowSidebar
        {...defaultProps}
        sidebarTab="details"
        viewMode="executions"
        selectedNode={selectedNode}
        selectedNodeDef={selectedNodeDef}
      />
    );

    expect(screen.queryByTestId('node-details-panel')).not.toBeInTheDocument();
  });

  it('should handle empty executions array', () => {
    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        executions={[]}
      />
    );

    expect(
      screen.getByText(
        'No executions found. Run your workflow to see execution history.'
      )
    ).toBeInTheDocument();
  });

  it('should show execution IDs truncated', () => {
    const executions = [
      {
        id: 'very-long-execution-id-12345',
        created_at: '2023-01-01T12:00:00Z',
      },
    ];

    render(
      <WorkflowSidebar
        {...defaultProps}
        viewMode="executions"
        executions={executions}
      />
    );

    expect(screen.getByText('ID: very-lon...')).toBeInTheDocument();
  });
});
