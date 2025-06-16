import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import CustomEdge from '../CustomEdge';

// Mock ReactFlow
vi.mock('reactflow', () => ({
  BaseEdge: ({ onMouseEnter, onMouseLeave, style }) => (
    <svg>
      <path
        data-testid="base-edge"
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
        style={style}
      />
    </svg>
  ),
  EdgeLabelRenderer: ({ children }) => (
    <div data-testid="edge-label-renderer">{children}</div>
  ),
  getBezierPath: () => ['M100,100 L200,200', 150, 150],
}));

describe('Edge Deletion Integration', () => {
  const defaultProps = {
    id: 'test-edge-1',
    sourceX: 100,
    sourceY: 100,
    targetX: 200,
    targetY: 200,
    sourcePosition: 'right',
    targetPosition: 'left',
  };

  it('should integrate delete functionality correctly', () => {
    const mockOnDelete = vi.fn();

    render(<CustomEdge {...defaultProps} onDelete={mockOnDelete} />);

    // Edge should render
    expect(screen.getByTestId('base-edge')).toBeInTheDocument();

    // No delete button initially
    expect(screen.queryByTitle('Delete edge')).not.toBeInTheDocument();

    // Hover to show delete button
    fireEvent.mouseEnter(screen.getByTestId('base-edge'));

    // Delete button should appear
    const deleteButton = screen.getByTitle('Delete edge');
    expect(deleteButton).toBeInTheDocument();
    expect(deleteButton).toHaveTextContent('🗑️');

    // Click delete button
    fireEvent.click(deleteButton);

    // onDelete should be called with correct ID
    expect(mockOnDelete).toHaveBeenCalledTimes(1);
    expect(mockOnDelete).toHaveBeenCalledWith('test-edge-1');
  });

  it('should show visual feedback on hover', () => {
    render(<CustomEdge {...defaultProps} />);

    const baseEdge = screen.getByTestId('base-edge');

    // Initial style
    expect(baseEdge).toHaveStyle({ strokeWidth: '2' });

    // Hover
    fireEvent.mouseEnter(baseEdge);
    expect(baseEdge).toHaveStyle({ strokeWidth: '3', stroke: '#ef4444' });

    // Unhover
    fireEvent.mouseLeave(baseEdge);
    expect(baseEdge).toHaveStyle({ strokeWidth: '2' });
  });

  it('should handle missing onDelete prop gracefully', () => {
    render(<CustomEdge {...defaultProps} />);

    // Hover to show delete button
    fireEvent.mouseEnter(screen.getByTestId('base-edge'));

    // Delete button should still appear
    const deleteButton = screen.getByTitle('Delete edge');
    expect(deleteButton).toBeInTheDocument();

    // Clicking should not throw error
    expect(() => {
      fireEvent.click(deleteButton);
    }).not.toThrow();
  });

  it('should position delete button correctly', () => {
    render(<CustomEdge {...defaultProps} />);

    fireEvent.mouseEnter(screen.getByTestId('base-edge'));

    const deleteButton = screen.getByTitle('Delete edge');
    const container = deleteButton.parentElement;

    expect(container).toHaveStyle({
      position: 'absolute',
      transform: 'translate(-50%, -50%) translate(150px,150px)',
    });
  });

  it('should have correct button styling', () => {
    render(<CustomEdge {...defaultProps} />);

    fireEvent.mouseEnter(screen.getByTestId('base-edge'));

    const deleteButton = screen.getByTitle('Delete edge');

    expect(deleteButton).toHaveClass(
      'w-6',
      'h-6',
      'bg-red-500',
      'text-white',
      'rounded-full',
      'shadow-lg'
    );
  });

  describe('Edge Deletion Handler Logic', () => {
    it('should simulate BuilderPage edge deletion flow', async () => {
      // Simulate the flow that happens in BuilderPage
      const mockDeleteEdgeAPI = vi.fn().mockResolvedValue();
      const mockBroadcast = vi.fn();

      // This simulates the handleEdgeDelete function from BuilderPage
      const handleEdgeDelete = async (edgeId) => {
        try {
          await mockDeleteEdgeAPI(edgeId);
          mockBroadcast('edgeDeleted', { edgeId });
        } catch (err) {
          console.error('Failed to delete edge:', err);
        }
      };

      render(<CustomEdge {...defaultProps} onDelete={handleEdgeDelete} />);

      fireEvent.mouseEnter(screen.getByTestId('base-edge'));
      const deleteButton = screen.getByTitle('Delete edge');

      await fireEvent.click(deleteButton);

      // Verify the deletion flow
      expect(mockDeleteEdgeAPI).toHaveBeenCalledWith('test-edge-1');
      expect(mockBroadcast).toHaveBeenCalledWith('edgeDeleted', {
        edgeId: 'test-edge-1',
      });
    });

    it('should handle edge deletion API errors', async () => {
      const mockDeleteEdgeAPI = vi
        .fn()
        .mockRejectedValue(new Error('API Error'));
      const mockBroadcast = vi.fn();
      const consoleSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      const handleEdgeDelete = async (edgeId) => {
        try {
          await mockDeleteEdgeAPI(edgeId);
          mockBroadcast('edgeDeleted', { edgeId });
        } catch (err) {
          console.error('Failed to delete edge:', err);
        }
      };

      render(<CustomEdge {...defaultProps} onDelete={handleEdgeDelete} />);

      fireEvent.mouseEnter(screen.getByTestId('base-edge'));
      const deleteButton = screen.getByTitle('Delete edge');

      await fireEvent.click(deleteButton);

      expect(mockDeleteEdgeAPI).toHaveBeenCalled();
      expect(mockBroadcast).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to delete edge:',
        expect.any(Error)
      );

      consoleSpy.mockRestore();
    });
  });

  describe('Node Deletion Edge Cleanup Logic', () => {
    it('should simulate node deletion with edge cleanup', async () => {
      const mockDeleteNodeAPI = vi.fn().mockResolvedValue();
      const mockDeleteEdgeAPI = vi.fn().mockResolvedValue();
      const mockBroadcast = vi.fn();

      const edges = [
        { id: 'edge-1', source: 'node-1', target: 'node-2' },
        { id: 'edge-2', source: 'node-3', target: 'node-1' },
      ];

      // Simulate handleNodeDelete from BuilderPage
      const handleNodeDelete = async (nodeId) => {
        try {
          const connectedEdges = edges.filter(
            (edge) => edge.source === nodeId || edge.target === nodeId
          );

          await mockDeleteNodeAPI(nodeId);
          mockBroadcast('nodeDeleted', { nodeId });

          for (const edge of connectedEdges) {
            try {
              await mockDeleteEdgeAPI(edge.id);
              mockBroadcast('edgeDeleted', { edgeId: edge.id });
            } catch (edgeErr) {
              console.error('Failed to delete edge:', edge.id, edgeErr);
            }
          }
        } catch (err) {
          console.error('Failed to delete node:', err);
        }
      };

      // Simulate deleting node-1 which has 2 connected edges
      await handleNodeDelete('node-1');

      expect(mockDeleteNodeAPI).toHaveBeenCalledWith('node-1');
      expect(mockBroadcast).toHaveBeenCalledWith('nodeDeleted', {
        nodeId: 'node-1',
      });

      // Should delete both connected edges
      expect(mockDeleteEdgeAPI).toHaveBeenCalledTimes(2);
      expect(mockDeleteEdgeAPI).toHaveBeenCalledWith('edge-1');
      expect(mockDeleteEdgeAPI).toHaveBeenCalledWith('edge-2');

      expect(mockBroadcast).toHaveBeenCalledWith('edgeDeleted', {
        edgeId: 'edge-1',
      });
      expect(mockBroadcast).toHaveBeenCalledWith('edgeDeleted', {
        edgeId: 'edge-2',
      });
    });
  });
});
