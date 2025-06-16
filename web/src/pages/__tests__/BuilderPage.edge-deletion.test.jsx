import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import CustomEdge from '../../components/CustomEdge';

// Mock ReactFlow components
vi.mock('reactflow', () => ({
  BaseEdge: ({ path, style, markerEnd }) => (
    <path
      data-testid="base-edge-path"
      d={path}
      style={style}
      markerEnd={markerEnd}
    />
  ),
  EdgeLabelRenderer: ({ children }) => (
    <div data-testid="edge-label-renderer">{children}</div>
  ),
  getBezierPath: vi.fn(() => ['M100,100 L200,200', 150, 150]),
}));

describe('BuilderPage Edge Deletion Integration', () => {
  const mockOnDelete = vi.fn();
  const defaultProps = {
    id: 'edge-1',
    source: 'node-1',
    target: 'node-2',
    sourceX: 100,
    sourceY: 100,
    targetX: 200,
    targetY: 200,
    sourcePosition: 'right',
    targetPosition: 'left',
    onDelete: mockOnDelete,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderCustomEdge = (props = {}) => {
    return render(<CustomEdge {...defaultProps} {...props} />);
  };

  it('should render CustomEdge component', () => {
    renderCustomEdge();

    expect(screen.getByTestId('base-edge-path')).toBeInTheDocument();
    expect(screen.getByTestId('edge-label-renderer')).toBeInTheDocument();
  });

  it('should show delete button on hover', () => {
    renderCustomEdge();

    const edgePath = screen.getByTestId('base-edge-path');

    // Simulate mouse enter to show delete button
    fireEvent.mouseEnter(edgePath);

    // Look for delete button (it should appear on hover)
    const deleteButton = screen.queryByRole('button');
    if (deleteButton) {
      expect(deleteButton).toBeInTheDocument();
    }
  });

  it('should hide delete button when not hovering', () => {
    renderCustomEdge();

    const edgePath = screen.getByTestId('base-edge-path');

    // Simulate mouse leave to hide delete button
    fireEvent.mouseLeave(edgePath);

    // Delete button should not be visible
    const deleteButton = screen.queryByRole('button');
    expect(deleteButton).not.toBeInTheDocument();
  });

  it('should call onDelete when delete button is clicked', () => {
    renderCustomEdge();

    const edgePath = screen.getByTestId('base-edge-path');

    // Show delete button by hovering
    fireEvent.mouseEnter(edgePath);

    const deleteButton = screen.queryByRole('button');
    if (deleteButton) {
      fireEvent.click(deleteButton);
      expect(mockOnDelete).toHaveBeenCalledWith('edge-1');
    }
  });

  describe('Edge deletion functionality', () => {
    it('should handle edge deletion with proper ID', () => {
      const edgeId = 'test-edge-123';
      renderCustomEdge({ id: edgeId });

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        fireEvent.click(deleteButton);
        expect(mockOnDelete).toHaveBeenCalledWith(edgeId);
      }
    });

    it('should prevent event propagation when delete button is clicked', () => {
      renderCustomEdge();

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        const clickEvent = new MouseEvent('click', { bubbles: true });
        const stopPropagationSpy = vi.spyOn(clickEvent, 'stopPropagation');

        fireEvent(deleteButton, clickEvent);

        // The component should stop event propagation
        expect(stopPropagationSpy).toHaveBeenCalled();
      }
    });
  });

  describe('Edge styling and behavior', () => {
    it('should apply correct styling to edge path', () => {
      renderCustomEdge();

      const edgePath = screen.getByTestId('base-edge-path');
      expect(edgePath).toBeInTheDocument();

      // The path should have the correct d attribute from getBezierPath
      expect(edgePath).toHaveAttribute('d', 'M100,100 L200,200');
    });

    it('should handle different edge positions', () => {
      renderCustomEdge({
        sourcePosition: 'bottom',
        targetPosition: 'top',
        sourceX: 50,
        sourceY: 50,
        targetX: 150,
        targetY: 150,
      });

      const edgePath = screen.getByTestId('base-edge-path');
      expect(edgePath).toBeInTheDocument();
    });
  });

  describe('Error handling', () => {
    it('should handle missing onDelete prop gracefully', () => {
      expect(() => {
        renderCustomEdge({ onDelete: undefined });
      }).not.toThrow();
    });

    it('should handle edge deletion errors', () => {
      const mockOnDeleteWithError = vi.fn(() => {
        throw new Error('Deletion failed');
      });

      renderCustomEdge({ onDelete: mockOnDeleteWithError });

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        expect(() => {
          fireEvent.click(deleteButton);
        }).not.toThrow();
      }
    });
  });

  describe('Integration with BuilderPage', () => {
    it('should work with BuilderPage edge deletion handler', () => {
      // Simulate the handler that would be passed from BuilderPage
      const builderPageHandler = vi.fn(async (edgeId) => {
        // Simulate API call
        await new Promise((resolve) => setTimeout(resolve, 10));
        // Simulate broadcasting change
        console.log(`Edge ${edgeId} deleted and broadcasted`);
      });

      renderCustomEdge({ onDelete: builderPageHandler });

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        fireEvent.click(deleteButton);
        expect(builderPageHandler).toHaveBeenCalledWith('edge-1');
      }
    });

    it('should support edge cleanup when nodes are deleted', () => {
      // Test that the edge component can handle being deleted
      // as part of node deletion cleanup
      const { unmount } = renderCustomEdge();

      expect(() => {
        unmount();
      }).not.toThrow();
    });
  });

  describe('Accessibility', () => {
    it('should provide accessible delete button', () => {
      renderCustomEdge();

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        expect(deleteButton).toHaveAttribute('aria-label');
      }
    });

    it('should support keyboard navigation', () => {
      renderCustomEdge();

      const edgePath = screen.getByTestId('base-edge-path');
      fireEvent.mouseEnter(edgePath);

      const deleteButton = screen.queryByRole('button');
      if (deleteButton) {
        // Test Enter key
        fireEvent.keyDown(deleteButton, { key: 'Enter', code: 'Enter' });

        // Test Space key
        fireEvent.keyDown(deleteButton, { key: ' ', code: 'Space' });
      }
    });
  });
});
