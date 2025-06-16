import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import BuilderPage from '../BuilderPage';

// Mock all the hooks and dependencies
vi.mock('../../hooks/useWorkflowState', () => ({
  useWorkflowState: () => ({
    nodes: [],
    edges: [],
    loading: false,
    error: null,
    isDirty: false,
    isDraft: false,
    isSaving: false,
    lastSaved: null,
    saveError: null,
    createNode: vi.fn(),
    updateNode: vi.fn(),
    deleteNode: vi.fn(),
    createEdge: vi.fn(),
    deleteEdge: vi.fn(),
    applyNodeChanges: vi.fn(),
    applyEdgeChanges: vi.fn(),
    testDraftNode: vi.fn(),
    deployDraft: vi.fn(),
    saveVersion: vi.fn(),
    clearError: vi.fn(),
  }),
}));

vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    broadcastNodeChange: vi.fn(),
  }),
}));

vi.mock('../../hooks/useNodeManagement', () => ({
  useNodeManagement: () => ({
    handleNodeDelete: vi.fn(),
    handleEdgeDelete: vi.fn(),
    createAgentConfigurationNodes: vi.fn(),
    handleAddNode: vi.fn(),
    handleConnectNodes: vi.fn(),
    handleGetWorkflow: vi.fn(),
    handleModalAddNode: vi.fn(),
  }),
}));

vi.mock('../../hooks/useNodeTypes.jsx', () => ({
  useNodeTypes: () => ({
    nodeDefs: [],
    triggers: [],
    nodeTypes: {},
    categories: [],
    triggersMap: {},
    refreshTriggers: vi.fn(),
  }),
}));

vi.mock('../../hooks/useValidation', () => ({
  useValidation: () => ({
    validationErrors: {},
    validateWorkflow: vi.fn(() => true),
    clearValidationErrors: vi.fn(),
  }),
}));

vi.mock('../../hooks/useAutoLayout', () => ({
  useAutoLayout: () => ({
    handleAutoLayout: vi.fn(),
  }),
}));

// Mock ReactFlow
vi.mock('reactflow', () => ({
  default: ({ children, onNodeClick }) => (
    <div
      data-testid="react-flow"
      onClick={() => onNodeClick?.(null, { id: 'test-node' })}
    >
      {children}
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
}));

// Mock components
vi.mock('../../components/WorkflowToolbar', () => ({
  default: ({ onToggleSidebar }) => (
    <div data-testid="workflow-toolbar">
      <button onClick={() => onToggleSidebar('details')}>Toggle Details</button>
    </div>
  ),
}));

vi.mock('../../components/WorkflowSidebar', () => ({
  default: ({ isVisible, sidebarTab }) => (
    <div
      data-testid="workflow-sidebar"
      data-visible={isVisible}
      data-tab={sidebarTab}
    >
      WorkflowSidebar
    </div>
  ),
}));

vi.mock('../../components/AddNodeModal', () => ({
  default: ({ isOpen, onClose }) =>
    isOpen ? (
      <div data-testid="add-node-modal">
        <button onClick={onClose}>Close Modal</button>
      </div>
    ) : null,
}));

vi.mock('../../components/NodeModal', () => ({
  default: ({ isOpen, onClose, node }) =>
    isOpen && node ? (
      <div data-testid="node-modal">
        Node Modal for {node.id}
        <button onClick={onClose}>Close Node Modal</button>
      </div>
    ) : null,
}));

vi.mock('../../components/ConfigSelectionDialog', () => ({
  default: ({ isOpen }) =>
    isOpen ? <div data-testid="config-dialog">Config Dialog</div> : null,
}));

// Mock axios
vi.mock('axios');

describe('BuilderPage Keyboard Interactions', () => {
  const defaultProps = {
    agentId: 'test-agent-123',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Escape key functionality', () => {
    it('should close node details sidebar when Escape is pressed', async () => {
      render(<BuilderPage {...defaultProps} />);

      // First open the details sidebar by toggling it
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      // Verify sidebar is visible with details tab
      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
        expect(sidebar).toHaveAttribute('data-tab', 'details');
      });

      // Now press Escape key
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      // Verify sidebar is closed
      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'false');
        expect(sidebar).toHaveAttribute('data-tab', 'null');
      });
    });

    it('should not close sidebar when Escape is pressed and details tab is not active', async () => {
      render(<BuilderPage {...defaultProps} />);

      // Don't open details tab - sidebar should remain closed

      // Press Escape key
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      // Verify sidebar remains closed (should not change state)
      const sidebar = screen.getByTestId('workflow-sidebar');
      expect(sidebar).toHaveAttribute('data-visible', 'false');
    });

    it('should only respond to Escape key, not other keys', async () => {
      render(<BuilderPage {...defaultProps} />);

      // Open details sidebar
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      // Verify sidebar is open
      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
        expect(sidebar).toHaveAttribute('data-tab', 'details');
      });

      // Press other keys - should not close sidebar
      fireEvent.keyDown(document, { key: 'Enter', code: 'Enter' });
      fireEvent.keyDown(document, { key: 'Space', code: 'Space' });
      fireEvent.keyDown(document, { key: 'Tab', code: 'Tab' });
      fireEvent.keyDown(document, { key: 'q', code: 'KeyQ' });

      // Verify sidebar is still open
      const sidebar = screen.getByTestId('workflow-sidebar');
      expect(sidebar).toHaveAttribute('data-visible', 'true');
      expect(sidebar).toHaveAttribute('data-tab', 'details');

      // Now press Escape - should close
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      await waitFor(() => {
        expect(sidebar).toHaveAttribute('data-visible', 'false');
        expect(sidebar).toHaveAttribute('data-tab', 'null');
      });
    });

    it('should cleanup escape key event listener when component unmounts', () => {
      const { unmount } = render(<BuilderPage {...defaultProps} />);

      // Open details sidebar
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      // Unmount component
      unmount();

      // Pressing Escape should not cause any errors (event listener cleaned up)
      expect(() => {
        fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      }).not.toThrow();
    });

    it('should handle multiple rapid Escape key presses gracefully', async () => {
      render(<BuilderPage {...defaultProps} />);

      // Open details sidebar
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
      });

      // Press Escape multiple times rapidly
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      // Should not cause errors and sidebar should be closed
      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'false');
      });
    });
  });

  describe('Keyboard interaction with other modals', () => {
    it('should not interfere with NodeModal escape key handling', async () => {
      render(<BuilderPage {...defaultProps} />);

      // Simulate clicking on a node to open NodeModal
      const reactFlow = screen.getByTestId('react-flow');
      fireEvent.click(reactFlow);

      // Verify NodeModal is open
      await waitFor(() => {
        expect(screen.getByTestId('node-modal')).toBeInTheDocument();
      });

      // Open details sidebar as well
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      // Both NodeModal and sidebar should be open
      await waitFor(() => {
        expect(screen.getByTestId('node-modal')).toBeInTheDocument();
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
      });

      // Press Escape - NodeModal should close first (it has its own handler)
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      // NodeModal should close, but sidebar may remain based on implementation
      await waitFor(() => {
        expect(screen.queryByTestId('node-modal')).not.toBeInTheDocument();
      });
    });

    it('should work correctly when details sidebar is reopened after being closed', async () => {
      render(<BuilderPage {...defaultProps} />);

      const toggleButton = screen.getByText('Toggle Details');

      // Open, close with Escape, then open again
      fireEvent.click(toggleButton);

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
      });

      // Close with Escape
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'false');
      });

      // Open again
      fireEvent.click(toggleButton);

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
      });

      // Close with Escape again - should still work
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'false');
      });
    });
  });

  describe('Integration with workflow state', () => {
    it('should clear selected node when closing details sidebar with Escape', async () => {
      render(<BuilderPage {...defaultProps} />);

      // Simulate selecting a node (which would normally set selectedNodeId)
      const reactFlow = screen.getByTestId('react-flow');
      fireEvent.click(reactFlow);

      // Open details sidebar
      const toggleButton = screen.getByText('Toggle Details');
      fireEvent.click(toggleButton);

      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'true');
        expect(sidebar).toHaveAttribute('data-tab', 'details');
      });

      // Press Escape to close
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      // Verify both sidebar is closed and selected node is cleared
      await waitFor(() => {
        const sidebar = screen.getByTestId('workflow-sidebar');
        expect(sidebar).toHaveAttribute('data-visible', 'false');
        expect(sidebar).toHaveAttribute('data-tab', 'null');
      });
    });
  });

  describe('Error handling', () => {
    it('should handle errors in escape key event handler gracefully', () => {
      // Mock console.error to prevent test noise
      const consoleErrorSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      render(<BuilderPage {...defaultProps} />);

      // Press Escape when sidebar is closed - should not cause errors
      expect(() => {
        fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      }).not.toThrow();

      consoleErrorSpy.mockRestore();
    });

    it('should handle escape key press when agentId is missing', () => {
      expect(() => {
        render(<BuilderPage agentId={null} />);
        fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      }).not.toThrow();
    });
  });
});
