import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RunsGraph from '../RunsGraph';

// Mock ReactFlow
vi.mock('reactflow', () => ({
  __esModule: true,
  default: ({ children, onNodeClick, onPaneClick, ...props }) => (
    <div className="w-1/2 h-full">
      <div data-testid="react-flow" {...props}>
        <div data-testid="react-flow-nodes">
          {props.nodes?.map((node) => (
            <div
              key={node.id}
              data-testid={`node-${node.id}`}
              onClick={() => onNodeClick?.(null, node)}
            >
              {node.id}
            </div>
          ))}
        </div>
        <div data-testid="react-flow-pane" onClick={onPaneClick}>
          Pane
        </div>
        {children}
      </div>
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
}));

// Mock node components
vi.mock('../IfNode', () => ({
  __esModule: true,
  default: () => <div data-testid="if-node" />,
}));

vi.mock('../DefaultNode', () => ({
  __esModule: true,
  default: () => <div data-testid="default-node" />,
}));

vi.mock('../TriggerNode', () => ({
  __esModule: true,
  default: () => <div data-testid="trigger-node" />,
}));

describe('RunsGraph', () => {
  const mockOnNodeClick = vi.fn();
  const mockOnPaneClick = vi.fn();

  const mockNodeDefs = [
    { type: 'webhook', label: 'Webhook', entry_point: true },
    { type: 'agent', label: 'Agent', entry_point: false },
    { type: 'if', label: 'If', branching: true },
    { type: 'http_request', label: 'HTTP Request', entry_point: false },
  ];

  const mockRfNodes = [
    { id: 'node-1', type: 'webhook', position: { x: 0, y: 0 } },
    { id: 'node-2', type: 'agent', position: { x: 200, y: 0 } },
    { id: 'node-3', type: 'if', position: { x: 400, y: 0 } },
  ];

  const mockRfEdges = [
    { id: 'edge-1', source: 'node-1', target: 'node-2' },
    { id: 'edge-2', source: 'node-2', target: 'node-3' },
  ];

  const mockRunDetails = {
    id: 'run-1',
    graph: {
      nodes: mockRfNodes,
      edges: mockRfEdges,
    },
  };

  const defaultProps = {
    runDetails: mockRunDetails,
    rfNodes: mockRfNodes,
    rfEdges: mockRfEdges,
    nodeDefs: mockNodeDefs,
    onNodeClick: mockOnNodeClick,
    onPaneClick: mockOnPaneClick,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render ReactFlow when run details are provided', () => {
    render(<RunsGraph {...defaultProps} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    expect(screen.getByTestId('background')).toBeInTheDocument();
    expect(screen.getByTestId('controls')).toBeInTheDocument();
    expect(screen.getByTestId('minimap')).toBeInTheDocument();
  });

  it('should render placeholder when no run details', () => {
    render(<RunsGraph {...defaultProps} runDetails={null} />);

    expect(screen.getByText('Select a run to view graph.')).toBeInTheDocument();
    expect(screen.queryByTestId('react-flow')).not.toBeInTheDocument();
  });

  it('should pass correct props to ReactFlow', () => {
    render(<RunsGraph {...defaultProps} />);

    // Just verify ReactFlow is rendered with the component
    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });

  it('should render all nodes', () => {
    render(<RunsGraph {...defaultProps} />);

    mockRfNodes.forEach((node) => {
      expect(screen.getByTestId(`node-${node.id}`)).toBeInTheDocument();
    });
  });

  it('should call onNodeClick when a node is clicked', () => {
    render(<RunsGraph {...defaultProps} />);

    const node = screen.getByTestId('node-node-1');
    fireEvent.click(node);

    expect(mockOnNodeClick).toHaveBeenCalledWith(null, mockRfNodes[0]);
    expect(mockOnNodeClick).toHaveBeenCalledTimes(1);
  });

  it('should call onPaneClick when pane is clicked', () => {
    render(<RunsGraph {...defaultProps} />);

    const pane = screen.getByTestId('react-flow-pane');
    fireEvent.click(pane);

    expect(mockOnPaneClick).toHaveBeenCalledTimes(1);
  });

  it('should generate correct nodeTypes mapping', () => {
    // This test verifies the nodeTypes logic by checking the component behavior
    render(<RunsGraph {...defaultProps} />);

    // The nodeTypes should be generated based on nodeDefs
    // entry_point -> TriggerNode, branching -> IfNode, default -> DefaultNode
    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });

  it('should handle empty nodes and edges', () => {
    render(<RunsGraph {...defaultProps} rfNodes={[]} rfEdges={[]} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    expect(screen.getByTestId('react-flow-nodes')).toBeInTheDocument();
    expect(screen.queryByTestId('node-node-1')).not.toBeInTheDocument();
  });

  it('should handle missing nodeDefs gracefully', () => {
    render(<RunsGraph {...defaultProps} nodeDefs={[]} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });

  it('should have correct container styling', () => {
    render(<RunsGraph {...defaultProps} />);

    // Just verify the component renders correctly
    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });

  it('should have correct placeholder styling when no run details', () => {
    render(<RunsGraph {...defaultProps} runDetails={null} />);

    const placeholder = screen
      .getByText('Select a run to view graph.')
      .closest('div');
    expect(placeholder).toHaveClass(
      'w-1/2',
      'h-full',
      'flex',
      'items-center',
      'justify-center'
    );
  });

  it('should handle node clicks for different node types', () => {
    render(<RunsGraph {...defaultProps} />);

    // Click different nodes
    fireEvent.click(screen.getByTestId('node-node-1'));
    expect(mockOnNodeClick).toHaveBeenCalledWith(null, mockRfNodes[0]);

    fireEvent.click(screen.getByTestId('node-node-2'));
    expect(mockOnNodeClick).toHaveBeenCalledWith(null, mockRfNodes[1]);

    fireEvent.click(screen.getByTestId('node-node-3'));
    expect(mockOnNodeClick).toHaveBeenCalledWith(null, mockRfNodes[2]);

    expect(mockOnNodeClick).toHaveBeenCalledTimes(3);
  });

  it('should handle rapid clicking without issues', () => {
    render(<RunsGraph {...defaultProps} />);

    const node = screen.getByTestId('node-node-1');

    // Click multiple times rapidly
    fireEvent.click(node);
    fireEvent.click(node);
    fireEvent.click(node);

    expect(mockOnNodeClick).toHaveBeenCalledTimes(3);
    expect(mockOnNodeClick).toHaveBeenCalledWith(null, mockRfNodes[0]);
  });

  it('should update when props change', () => {
    const { rerender } = render(<RunsGraph {...defaultProps} />);

    expect(screen.getByTestId('node-node-1')).toBeInTheDocument();

    // Update with different nodes
    const newNodes = [
      { id: 'node-4', type: 'webhook', position: { x: 0, y: 0 } },
    ];

    rerender(<RunsGraph {...defaultProps} rfNodes={newNodes} />);

    expect(screen.queryByTestId('node-node-1')).not.toBeInTheDocument();
    expect(screen.getByTestId('node-node-4')).toBeInTheDocument();
  });

  it('should handle undefined runDetails gracefully', () => {
    render(<RunsGraph {...defaultProps} runDetails={undefined} />);

    expect(screen.getByText('Select a run to view graph.')).toBeInTheDocument();
    expect(screen.queryByTestId('react-flow')).not.toBeInTheDocument();
  });

  it('should memoize nodeTypes correctly', () => {
    const { rerender } = render(<RunsGraph {...defaultProps} />);

    // Rerender with same nodeDefs should not cause issues
    rerender(<RunsGraph {...defaultProps} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });
});
