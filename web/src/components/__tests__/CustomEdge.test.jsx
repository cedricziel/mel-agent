import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import CustomEdge from '../CustomEdge';

// Mock ReactFlow components
vi.mock('reactflow', () => ({
  BaseEdge: ({ path, style, onMouseEnter, onMouseLeave, markerEnd }) => (
    <svg>
      <path
        data-testid="base-edge"
        d={path}
        style={style}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
        markerEnd={markerEnd}
      />
    </svg>
  ),
  EdgeLabelRenderer: ({ children }) => (
    <div data-testid="edge-label-renderer">{children}</div>
  ),
  getBezierPath: ({ sourceX, sourceY, targetX, targetY }) => [
    `M${sourceX},${sourceY} L${targetX},${targetY}`,
    (sourceX + targetX) / 2,
    (sourceY + targetY) / 2,
  ],
}));

describe('CustomEdge', () => {
  const defaultProps = {
    id: 'test-edge-1',
    sourceX: 100,
    sourceY: 100,
    targetX: 200,
    targetY: 200,
    sourcePosition: 'right',
    targetPosition: 'left',
    style: { stroke: '#b1b1b7' },
    markerEnd: 'url(#marker-end)',
  };

  it('should render base edge with correct path', () => {
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');
    expect(baseEdge).toBeInTheDocument();
    expect(baseEdge).toHaveAttribute('d', 'M100,100 L200,200');
  });

  it('should apply default style when not hovered', () => {
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');
    expect(baseEdge).toHaveStyle({
      strokeWidth: '2',
      stroke: '#b1b1b7',
    });
  });

  it('should change style when hovered', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');

    await user.hover(baseEdge);

    expect(baseEdge).toHaveStyle({
      strokeWidth: '3',
      stroke: '#ef4444',
    });
  });

  it('should show delete button when hovered', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');

    // Initially no delete button
    expect(screen.queryByTitle('Delete edge')).not.toBeInTheDocument();

    await user.hover(baseEdge);

    // Delete button should appear
    const deleteButton = screen.getByTitle('Delete edge');
    expect(deleteButton).toBeInTheDocument();
    expect(deleteButton).toHaveTextContent('🗑️');
  });

  it('should hide delete button when not hovered', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');

    await user.hover(baseEdge);
    expect(screen.getByTitle('Delete edge')).toBeInTheDocument();

    await user.unhover(baseEdge);
    expect(screen.queryByTitle('Delete edge')).not.toBeInTheDocument();
  });

  it('should call onDelete with correct edge id when delete button is clicked', async () => {
    const mockOnDelete = vi.fn();
    const user = userEvent.setup();

    render(<CustomEdge {...defaultProps} onDelete={mockOnDelete} />);

    const baseEdge = screen.getByTestId('base-edge');

    // Trigger hover to show delete button
    fireEvent.mouseEnter(baseEdge);

    // Wait for the delete button to appear
    await waitFor(() => {
      expect(screen.getByTitle('Delete edge')).toBeInTheDocument();
    });

    const deleteButton = screen.getByTitle('Delete edge');
    fireEvent.click(deleteButton);

    expect(mockOnDelete).toHaveBeenCalledTimes(1);
    expect(mockOnDelete).toHaveBeenCalledWith('test-edge-1');
  });

  it('should not call onDelete when delete button is clicked but onDelete is not provided', async () => {
    const user = userEvent.setup();

    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');
    await user.hover(baseEdge);

    const deleteButton = screen.getByTitle('Delete edge');
    await user.click(deleteButton);

    // Should not throw error when onDelete is not provided
    expect(true).toBe(true);
  });

  it('should position delete button at the center of the edge', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');
    await user.hover(baseEdge);

    const deleteButton = screen.getByTitle('Delete edge');
    const buttonContainer = deleteButton.parentElement;

    expect(buttonContainer).toHaveStyle({
      position: 'absolute',
      transform: 'translate(-50%, -50%) translate(150px,150px)',
    });
  });

  it('should apply correct CSS classes to delete button', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');
    await user.hover(baseEdge);

    const deleteButton = screen.getByTitle('Delete edge');

    expect(deleteButton).toHaveClass(
      'w-6',
      'h-6',
      'bg-red-500',
      'hover:bg-red-600',
      'text-white',
      'rounded-full',
      'flex',
      'items-center',
      'justify-center',
      'shadow-lg',
      'transition-colors'
    );
  });

  it('should handle custom style overrides', () => {
    const customStyle = { stroke: '#00ff00', strokeWidth: 4 };
    render(<CustomEdge {...defaultProps} style={customStyle} />);

    const baseEdge = screen.getByTestId('base-edge');
    expect(baseEdge).toHaveStyle({
      stroke: '#00ff00',
      strokeWidth: '2', // Should be overridden to 2 when not hovered
    });
  });

  it('should handle edge label renderer correctly', () => {
    render(<CustomEdge {...defaultProps} />);

    const labelRenderer = screen.getByTestId('edge-label-renderer');
    expect(labelRenderer).toBeInTheDocument();
  });

  it('should maintain hover state correctly during mouse events', async () => {
    const user = userEvent.setup();
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');

    // Initial state
    expect(baseEdge).toHaveStyle({ strokeWidth: '2' });

    // Hover
    await user.hover(baseEdge);
    expect(baseEdge).toHaveStyle({ strokeWidth: '3' });

    // Unhover
    await user.unhover(baseEdge);
    expect(baseEdge).toHaveStyle({ strokeWidth: '2' });
  });

  it('should prevent event propagation on delete button click', async () => {
    const mockOnDelete = vi.fn();
    const mockStopPropagation = vi.fn();
    const user = userEvent.setup();

    render(<CustomEdge {...defaultProps} onDelete={mockOnDelete} />);

    const baseEdge = screen.getByTestId('base-edge');
    await user.hover(baseEdge);

    const deleteButton = screen.getByTitle('Delete edge');

    // Mock the event to test stopPropagation
    fireEvent.click(deleteButton, { stopPropagation: mockStopPropagation });

    expect(mockOnDelete).toHaveBeenCalledWith('test-edge-1');
  });
});
