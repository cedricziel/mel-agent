import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import NodeExecutionPanel from '../NodeExecutionPanel';

describe('NodeExecutionPanel', () => {
  const defaultProps = {
    viewMode: 'editor',
    selectedExecution: null,
    onExecute: vi.fn(),
    outputData: {},
    setOutputData: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render test controls in editor mode', () => {
    render(<NodeExecutionPanel {...defaultProps} />);

    expect(screen.getByText('Test Node')).toBeInTheDocument();
    expect(
      screen.getByText('Run this node with current configuration')
    ).toBeInTheDocument();
    expect(screen.getByText('Mode: Editor')).toBeInTheDocument();
  });

  it('should render execution info in executions mode', () => {
    const selectedExecution = {
      id: 'exec-1',
      created_at: '2023-01-01T00:00:00Z',
    };

    render(
      <NodeExecutionPanel
        {...defaultProps}
        viewMode="executions"
        selectedExecution={selectedExecution}
      />
    );

    // Use timezone-independent assertion - check for date components instead of exact format
    const expectedDate = new Date('2023-01-01T00:00:00Z');
    const formattedDate = expectedDate.toLocaleString();
    expect(
      screen.getByText(`Viewing execution data from ${formattedDate}`)
    ).toBeInTheDocument();
    expect(screen.getByText('Mode: Execution View')).toBeInTheDocument();
    expect(screen.queryByText('Test Node')).not.toBeInTheDocument();
  });

  it('should handle test node execution', async () => {
    const onExecute = vi.fn().mockResolvedValue({ result: 'success' });
    const setOutputData = vi.fn();

    render(
      <NodeExecutionPanel
        {...defaultProps}
        onExecute={onExecute}
        setOutputData={setOutputData}
      />
    );

    const testButton = screen.getByText('Test Node');
    fireEvent.click(testButton);

    await waitFor(() => {
      expect(onExecute).toHaveBeenCalledWith({});
      expect(setOutputData).toHaveBeenCalledWith({ result: 'success' });
    });
  });

  it('should handle test node execution errors', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    const onExecute = vi.fn().mockRejectedValue(new Error('Test error'));
    const setOutputData = vi.fn();

    render(
      <NodeExecutionPanel
        {...defaultProps}
        onExecute={onExecute}
        setOutputData={setOutputData}
      />
    );

    const testButton = screen.getByText('Test Node');
    fireEvent.click(testButton);

    await waitFor(() => {
      expect(onExecute).toHaveBeenCalledWith({});
      expect(setOutputData).toHaveBeenCalledWith({
        error: 'Test error',
      });
      expect(consoleSpy).toHaveBeenCalledWith(
        'Node execution error:',
        expect.any(Error)
      );
    });

    consoleSpy.mockRestore();
  });

  it('should handle execution with no result', async () => {
    const onExecute = vi.fn().mockResolvedValue(null);
    const setOutputData = vi.fn();

    render(
      <NodeExecutionPanel
        {...defaultProps}
        onExecute={onExecute}
        setOutputData={setOutputData}
      />
    );

    const testButton = screen.getByText('Test Node');
    fireEvent.click(testButton);

    await waitFor(() => {
      expect(setOutputData).toHaveBeenCalledWith({});
    });
  });

  it('should handle missing onExecute prop', async () => {
    const setOutputData = vi.fn();

    render(
      <NodeExecutionPanel
        {...defaultProps}
        onExecute={null}
        setOutputData={setOutputData}
      />
    );

    const testButton = screen.getByText('Test Node');
    fireEvent.click(testButton);

    // Should not crash or call setOutputData
    expect(setOutputData).not.toHaveBeenCalled();
  });

  it('should show unknown time when selectedExecution is null in executions mode', () => {
    render(
      <NodeExecutionPanel
        {...defaultProps}
        viewMode="executions"
        selectedExecution={null}
      />
    );

    expect(
      screen.getByText('Viewing execution data from unknown time')
    ).toBeInTheDocument();
  });

  it('should handle execution errors without message', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();
    const onExecute = vi.fn().mockRejectedValue({});
    const setOutputData = vi.fn();

    render(
      <NodeExecutionPanel
        {...defaultProps}
        onExecute={onExecute}
        setOutputData={setOutputData}
      />
    );

    const testButton = screen.getByText('Test Node');
    fireEvent.click(testButton);

    await waitFor(() => {
      expect(setOutputData).toHaveBeenCalledWith({
        error: 'Execution failed',
      });
    });

    consoleSpy.mockRestore();
  });

  it('should have correct styling classes', () => {
    const { container } = render(<NodeExecutionPanel {...defaultProps} />);

    const panel = container.firstChild;
    expect(panel).toHaveClass('border-t', 'p-4', 'bg-gray-50');
  });

  it('should render test button with correct styling', () => {
    render(<NodeExecutionPanel {...defaultProps} />);

    const testButton = screen.getByText('Test Node');
    expect(testButton).toHaveClass(
      'px-4',
      'py-2',
      'bg-green-600',
      'text-white',
      'rounded',
      'hover:bg-green-700'
    );
  });
});
