import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import NodeModal from '../NodeModal';

// Mock child components
vi.mock('../DataViewer', () => ({
  default: ({ data, title }) => (
    <div data-testid="data-viewer">
      {title}: {JSON.stringify(data)}
    </div>
  ),
}));

vi.mock('../NodeConfigurationPanel', () => ({
  default: ({ node, handleChange }) => (
    <div data-testid="node-configuration-panel">
      Configuration for {node?.id}
      <button onClick={() => handleChange('test', 'value')}>Change</button>
    </div>
  ),
}));

vi.mock('../NodeExecutionPanel', () => ({
  default: ({ onExecute }) => (
    <div data-testid="node-execution-panel">
      <button onClick={() => onExecute({ test: 'data' })}>Execute</button>
    </div>
  ),
}));

// Mock the useNodeModalState hook
vi.mock('../../hooks/useNodeModalState', () => ({
  useNodeModalState: () => ({
    currentFormData: { test: 'form data' },
    dynamicOptions: {},
    loadingOptions: false,
    credentials: [],
    handleChange: vi.fn(),
    inputData: { input: 'test data' },
    outputData: { output: 'test result' },
    setOutputData: vi.fn(),
    activeTab: 'config',
    setActiveTab: vi.fn(),
    loadNodeExecutionData: vi.fn(),
  }),
}));

describe('NodeModal', () => {
  const mockOnClose = vi.fn();
  const mockOnChange = vi.fn();
  const mockOnExecute = vi.fn();
  const mockOnSave = vi.fn();

  const mockNode = {
    id: 'test-node-1',
    type: 'webhook',
    data: { label: 'Test Webhook' },
    position: { x: 100, y: 100 },
  };

  const mockNodeDef = {
    type: 'webhook',
    label: 'Webhook',
    icon: '🔗',
    description: 'Test webhook node',
    parameters: [],
  };

  const defaultProps = {
    node: mockNode,
    nodeDef: mockNodeDef,
    nodes: [mockNode],
    isOpen: true,
    onClose: mockOnClose,
    onChange: mockOnChange,
    onExecute: mockOnExecute,
    onSave: mockOnSave,
    viewMode: 'editor',
    selectedExecution: null,
    agentId: 'test-agent',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic rendering', () => {
    it('should not render when isOpen is false', () => {
      render(<NodeModal {...defaultProps} isOpen={false} />);

      expect(screen.queryByText('Webhook')).not.toBeInTheDocument();
    });

    it('should not render when node is null', () => {
      render(<NodeModal {...defaultProps} node={null} />);

      expect(screen.queryByText('Webhook')).not.toBeInTheDocument();
    });

    it('should not render when nodeDef is null', () => {
      render(<NodeModal {...defaultProps} nodeDef={null} />);

      expect(screen.queryByText('Webhook')).not.toBeInTheDocument();
    });

    it('should render modal when isOpen is true with valid props', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('Webhook')).toBeInTheDocument();
      expect(screen.getByText('(test-node-1)')).toBeInTheDocument();
      expect(screen.getByText('🔗')).toBeInTheDocument();
    });
  });

  describe('Modal header', () => {
    it('should display node icon, label, and ID', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('🔗')).toBeInTheDocument();
      expect(screen.getByText('Webhook')).toBeInTheDocument();
      expect(screen.getByText('(test-node-1)')).toBeInTheDocument();
    });

    it('should have Save and Close buttons', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('Save')).toBeInTheDocument();
      expect(screen.getByText('Close')).toBeInTheDocument();
    });

    it('should disable Save button in executions mode', () => {
      render(<NodeModal {...defaultProps} viewMode="executions" />);

      const saveButton = screen.getByText('Save');
      expect(saveButton).toBeDisabled();
      expect(saveButton).toHaveClass('cursor-not-allowed');
    });

    it('should enable Save button in editor mode', () => {
      render(<NodeModal {...defaultProps} viewMode="editor" />);

      const saveButton = screen.getByText('Save');
      expect(saveButton).not.toBeDisabled();
      expect(saveButton).not.toHaveClass('cursor-not-allowed');
    });
  });

  describe('Modal panels', () => {
    it('should render input data panel', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('Input Data')).toBeInTheDocument();
      expect(screen.getByTestId('data-viewer')).toBeInTheDocument();
    });

    it('should render configuration panel', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('Configuration')).toBeInTheDocument();
      expect(
        screen.getByTestId('node-configuration-panel')
      ).toBeInTheDocument();
    });

    it('should render output data panel', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByText('Output Data')).toBeInTheDocument();
    });

    it('should render execution panel', () => {
      render(<NodeModal {...defaultProps} />);

      expect(screen.getByTestId('node-execution-panel')).toBeInTheDocument();
    });

    it('should show "Node Data" tab in executions mode', () => {
      render(<NodeModal {...defaultProps} viewMode="executions" />);

      expect(screen.getByText('Node Data')).toBeInTheDocument();
      expect(screen.queryByText('Configuration')).not.toBeInTheDocument();
    });
  });

  describe('Button interactions', () => {
    it('should call onSave when Save button is clicked', () => {
      render(<NodeModal {...defaultProps} />);

      const saveButton = screen.getByText('Save');
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when Close button is clicked', () => {
      render(<NodeModal {...defaultProps} />);

      const closeButton = screen.getByText('Close');
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Execution info in executions mode', () => {
    it('should show execution info when selectedExecution is provided', () => {
      const selectedExecution = {
        id: 'exec-1',
        created_at: '2023-01-01T12:00:00Z',
      };

      render(
        <NodeModal
          {...defaultProps}
          viewMode="executions"
          selectedExecution={selectedExecution}
        />
      );

      expect(screen.getByText(/From execution:/)).toBeInTheDocument();
      expect(screen.getByText(/1\/1\/2023/)).toBeInTheDocument();
    });
  });

  describe('Escape key functionality', () => {
    it('should call onClose when Escape key is pressed and modal is open', () => {
      render(<NodeModal {...defaultProps} />);

      // Simulate escape key press
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should not call onClose when Escape key is pressed and modal is closed', () => {
      render(<NodeModal {...defaultProps} isOpen={false} />);

      // Simulate escape key press
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('should only respond to Escape key, not other keys', () => {
      render(<NodeModal {...defaultProps} />);

      // Simulate other key presses
      fireEvent.keyDown(document, { key: 'Enter', code: 'Enter' });
      fireEvent.keyDown(document, { key: 'Space', code: 'Space' });
      fireEvent.keyDown(document, { key: 'Tab', code: 'Tab' });

      expect(mockOnClose).not.toHaveBeenCalled();

      // Now simulate Escape
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should cleanup event listener when modal is closed', () => {
      const { rerender } = render(<NodeModal {...defaultProps} />);

      // Verify escape works when modal is open
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      expect(mockOnClose).toHaveBeenCalledTimes(1);

      // Clear the mock
      mockOnClose.mockClear();

      // Close the modal
      rerender(<NodeModal {...defaultProps} isOpen={false} />);

      // Escape should no longer trigger onClose
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('should cleanup event listener when component unmounts', () => {
      const { unmount } = render(<NodeModal {...defaultProps} />);

      // Unmount the component
      unmount();

      // Escape should no longer trigger onClose
      fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });
      expect(mockOnClose).not.toHaveBeenCalled();
    });
  });

  describe('Modal structure and accessibility', () => {
    it('should have correct modal backdrop styling', () => {
      render(<NodeModal {...defaultProps} />);

      const backdrop = screen.getByText('Webhook').closest('.fixed');
      expect(backdrop).toHaveClass(
        'inset-0',
        'bg-black',
        'bg-opacity-50',
        'flex',
        'items-center',
        'justify-center',
        'z-50'
      );
    });

    it('should have correct modal content styling', () => {
      render(<NodeModal {...defaultProps} />);

      const modal = screen.getByText('Webhook').closest('.bg-white');
      expect(modal).toHaveClass(
        'bg-white',
        'rounded-lg',
        'w-full',
        'h-full',
        'flex',
        'flex-col'
      );
    });

    it('should have proper three-panel layout', () => {
      render(<NodeModal {...defaultProps} />);

      // Check for the main content flex container
      const mainContent = screen
        .getByText('Input Data')
        .closest('.flex-1.flex');
      expect(mainContent).toBeInTheDocument();

      // Check that we have the three main sections
      expect(screen.getByText('Input Data')).toBeInTheDocument();
      expect(screen.getByText('Configuration')).toBeInTheDocument();
      expect(screen.getByText('Output Data')).toBeInTheDocument();
    });
  });

  describe('Error handling', () => {
    it('should handle missing onClose prop gracefully', () => {
      expect(() => {
        render(<NodeModal {...defaultProps} onClose={undefined} />);
      }).not.toThrow();
    });

    it('should handle missing node properties', () => {
      const incompleteNode = { id: 'test' };
      const incompleteNodeDef = { type: 'test' };

      expect(() => {
        render(
          <NodeModal
            {...defaultProps}
            node={incompleteNode}
            nodeDef={incompleteNodeDef}
          />
        );
      }).not.toThrow();
    });
  });
});
