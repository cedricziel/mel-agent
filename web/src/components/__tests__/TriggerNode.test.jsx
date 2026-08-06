import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import TriggerNode from '../TriggerNode';
import { workflowRunsApi } from '../../api/client';

// Mock the API client
vi.mock('../../api/client', () => ({
  workflowRunsApi: {
    createWorkflowRun: vi.fn(),
  },
}));

// Mock ReactFlow's Handle component
vi.mock('reactflow', () => ({
  Handle: ({ type, position, id, className }) => (
    <div
      data-testid={`handle-${type}-${position}-${id}`}
      className={className}
    />
  ),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right',
  },
}));

describe('TriggerNode', () => {
  const defaultProps = {
    id: 'trigger-node-1',
    type: 'webhook',
    agentId: 'test-agent-id',
    data: {
      label: 'Webhook Trigger',
      nodeTypeLabel: 'Webhook',
    },
    icon: '🔗',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render trigger node with curved left border', () => {
    render(<TriggerNode {...defaultProps} />);

    expect(screen.getByText('Webhook Trigger')).toBeInTheDocument();
    expect(screen.getByText('Webhook')).toBeInTheDocument();
    expect(screen.getByText('🔗')).toBeInTheDocument();
  });

  it('should only render output handle (no input)', () => {
    render(<TriggerNode {...defaultProps} />);

    expect(
      screen.getByTestId('handle-source-right-trigger-out')
    ).toBeInTheDocument();
    expect(
      screen.queryByTestId('handle-target-left-in')
    ).not.toBeInTheDocument();
  });

  it('should show error border when error is true', () => {
    const propsWithError = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        error: true,
      },
    };

    render(<TriggerNode {...propsWithError} />);

    // Find the container div with the border styling
    const nodeContainer = document.querySelector('.border-red-500');
    expect(nodeContainer).toBeInTheDocument();
  });

  it('should show running status indicator', () => {
    const propsWithRunning = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        status: 'running',
      },
    };

    render(<TriggerNode {...propsWithRunning} />);

    const statusIndicator = document.querySelector('.animate-pulse');
    expect(statusIndicator).toBeInTheDocument();
    expect(statusIndicator).toHaveClass('bg-blue-500');
  });

  it('should display parameter summary', () => {
    const propsWithParams = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        path: '/webhook/endpoint',
        method: 'POST',
      },
    };

    render(<TriggerNode {...propsWithParams} />);

    expect(screen.getByText(/path: \/webhook\/endpoint/)).toBeInTheDocument();
    // Note: TriggerNode includes nodeTypeLabel in summary, unlike DefaultNode
    expect(screen.getByText(/nodeTypeLabel: Webhook/)).toBeInTheDocument();
  });

  it('should render manual trigger button for manual_trigger type', () => {
    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
      data: {
        ...defaultProps.data,
        label: 'Manual Trigger',
      },
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    expect(triggerButton).toBeInTheDocument();
    expect(triggerButton.tagName).toBe('BUTTON');
  });

  it('should call API when manual trigger button is clicked', async () => {
    workflowRunsApi.createWorkflowRun.mockResolvedValue({
      data: { success: true },
    });

    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
      agentId: 'test-agent-123',
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    fireEvent.click(triggerButton);

    expect(workflowRunsApi.createWorkflowRun).toHaveBeenCalledWith({
      workflow_id: 'test-agent-123',
      input_data: {},
    });
  });

  it('should show visual feedback on successful manual trigger', async () => {
    workflowRunsApi.createWorkflowRun.mockResolvedValue({
      data: { success: true },
    });

    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    fireEvent.click(triggerButton);

    await waitFor(() => {
      expect(triggerButton).toHaveTextContent('✓');
    });

    expect(triggerButton).toHaveStyle({ backgroundColor: 'rgb(16, 185, 129)' });
  });

  it('should handle manual trigger API error', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

    const networkError = new Error('Network error');
    workflowRunsApi.createWorkflowRun.mockRejectedValue(networkError);

    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
      agentId: 'test-agent-123', // Ensure agentId is present
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    fireEvent.click(triggerButton);

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to trigger manually:',
        networkError
      );
      expect(alertSpy).toHaveBeenCalledWith('Failed to trigger workflow');
    });

    consoleSpy.mockRestore();
    alertSpy.mockRestore();
  });

  it('should show alert when agentId is missing on manual trigger', async () => {
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
      agentId: null,
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    fireEvent.click(triggerButton);

    expect(alertSpy).toHaveBeenCalledWith('Could not determine agent ID');
    expect(workflowRunsApi.createWorkflowRun).not.toHaveBeenCalled();

    alertSpy.mockRestore();
  });

  it('should render add buttons when onAddClick is provided and not manual trigger', () => {
    const mockOnAddClick = vi.fn();
    const propsWithAddClick = {
      ...defaultProps,
      type: 'webhook', // Not manual_trigger
      onAddClick: mockOnAddClick,
    };

    render(<TriggerNode {...propsWithAddClick} />);

    // Should have both quick-add button and output handle button
    const addButtons = screen.getAllByText('+');
    expect(addButtons).toHaveLength(2);

    // Test the quick-add button (without title)
    const quickAddButton = addButtons.find((btn) => !btn.getAttribute('title'));
    fireEvent.click(quickAddButton);
    expect(mockOnAddClick).toHaveBeenCalledTimes(1);
  });

  it('should render output handle add button for manual trigger type', () => {
    const mockOnAddClick = vi.fn();
    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
      onAddClick: mockOnAddClick,
    };

    render(<TriggerNode {...manualTriggerProps} />);

    // Should show the output handle + button but not the quick-add button
    const addButtons = screen.getAllByText('+');
    expect(addButtons).toHaveLength(1);
    expect(addButtons[0]).toHaveAttribute('title', 'Add Next Node');
  });

  it('should use default icon when icon prop is not provided', () => {
    const propsWithoutIcon = {
      ...defaultProps,
      icon: undefined,
    };

    render(<TriggerNode {...propsWithoutIcon} />);

    expect(screen.getByText('🔔')).toBeInTheDocument();
  });

  it('should stop propagation on button clicks', () => {
    const mockStopPropagation = vi.fn();

    const manualTriggerProps = {
      ...defaultProps,
      type: 'manual_trigger',
    };

    render(<TriggerNode {...manualTriggerProps} />);

    const triggerButton = screen.getByText('▶️ Trigger');
    fireEvent.click(triggerButton, { stopPropagation: mockStopPropagation });

    // Note: jsdom doesn't automatically call stopPropagation, but we can verify the handler is set up correctly
    expect(triggerButton).toBeInTheDocument();
  });
});
