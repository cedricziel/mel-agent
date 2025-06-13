import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import DefaultNode from '../DefaultNode';

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

describe('DefaultNode', () => {
  const defaultProps = {
    id: 'test-node-1',
    data: {
      label: 'Test Node',
      nodeTypeLabel: 'Test Type',
    },
  };

  it('should render node with label and type', () => {
    render(<DefaultNode {...defaultProps} />);

    expect(screen.getByText('Test Node')).toBeInTheDocument();
    expect(screen.getByText('Test Type')).toBeInTheDocument();
  });

  it('should render handles for input and output', () => {
    render(<DefaultNode {...defaultProps} />);

    expect(screen.getByTestId('handle-target-left-in')).toBeInTheDocument();
    expect(screen.getByTestId('handle-source-right-out')).toBeInTheDocument();
  });

  it('should show error state when error is true', () => {
    const propsWithError = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        error: true,
      },
    };

    render(<DefaultNode {...propsWithError} />);

    const nodeContainer = screen.getByText('Test Node').closest('.relative');
    expect(nodeContainer).toHaveClass('border-red-500');
  });

  it('should show running status indicator', () => {
    const propsWithRunning = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        status: 'running',
      },
    };

    render(<DefaultNode {...propsWithRunning} />);

    const statusIndicator = document.querySelector('.animate-pulse');
    expect(statusIndicator).toBeInTheDocument();
    expect(statusIndicator).toHaveClass('bg-blue-500');
  });

  it('should display parameter summary', () => {
    const propsWithParams = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        url: 'https://api.example.com',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      },
    };

    render(<DefaultNode {...propsWithParams} />);

    expect(
      screen.getByText(/url: https:\/\/api\.example\.com/)
    ).toBeInTheDocument();
    expect(screen.getByText(/method: POST/)).toBeInTheDocument();
    // Should show ellipsis for additional params
    expect(screen.getByText('…')).toBeInTheDocument();
  });

  it('should handle null and undefined parameter values', () => {
    const propsWithNullParams = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        nullValue: null,
        undefinedValue: undefined,
      },
    };

    render(<DefaultNode {...propsWithNullParams} />);

    expect(screen.getByText(/nullValue: null/)).toBeInTheDocument();
    expect(screen.getByText(/undefinedValue: undefined/)).toBeInTheDocument();
  });

  it('should handle empty string and zero values', () => {
    const propsWithEmptyValues = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        emptyString: '',
        zeroValue: 0,
      },
    };

    render(<DefaultNode {...propsWithEmptyValues} />);

    expect(screen.getByText('emptyString:')).toBeInTheDocument();
    expect(screen.getByText(/zeroValue: 0/)).toBeInTheDocument();
  });

  it('should render object parameters as JSON strings', () => {
    const propsWithObjectParam = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        config: { timeout: 5000, retries: 3 },
      },
    };

    render(<DefaultNode {...propsWithObjectParam} />);

    expect(
      screen.getByText(/config: {"timeout":5000,"retries":3}/)
    ).toBeInTheDocument();
  });

  it('should call onAddClick when add button is clicked', () => {
    const mockOnAddClick = vi.fn();
    const propsWithAddClick = {
      ...defaultProps,
      onAddClick: mockOnAddClick,
    };

    render(<DefaultNode {...propsWithAddClick} />);

    const addButton = screen.getByText('+');
    fireEvent.click(addButton);

    expect(mockOnAddClick).toHaveBeenCalledTimes(1);
  });

  it('should stop propagation when add button is clicked', () => {
    const mockOnAddClick = vi.fn();
    const mockStopPropagation = vi.fn();

    const propsWithAddClick = {
      ...defaultProps,
      onAddClick: mockOnAddClick,
    };

    render(<DefaultNode {...propsWithAddClick} />);

    const addButton = screen.getByText('+');
    fireEvent.click(addButton, { stopPropagation: mockStopPropagation });

    expect(mockOnAddClick).toHaveBeenCalled();
  });

  it('should not render add button when onAddClick is not provided', () => {
    render(<DefaultNode {...defaultProps} />);

    expect(screen.queryByText('+')).not.toBeInTheDocument();
  });

  it('should use fallback label when label is not provided', () => {
    const propsWithoutLabel = {
      ...defaultProps,
      data: {
        nodeTypeLabel: 'Test Type',
      },
    };

    render(<DefaultNode {...propsWithoutLabel} />);

    expect(screen.getByText('Test Type')).toBeInTheDocument();
  });

  it('should exclude certain keys from parameter summary', () => {
    const propsWithExcludedKeys = {
      ...defaultProps,
      data: {
        label: 'Test Node',
        nodeTypeLabel: 'Test Type',
        status: 'running',
        error: false,
        displayParam: 'should show',
        hiddenParam: 'should also show',
      },
    };

    render(<DefaultNode {...propsWithExcludedKeys} />);

    // Should not show label, nodeTypeLabel, status, or error in summary
    expect(screen.queryByText(/label:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/nodeTypeLabel:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/status:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/error:/)).not.toBeInTheDocument();

    // Should show other params
    expect(screen.getByText(/displayParam: should show/)).toBeInTheDocument();
    expect(
      screen.getByText(/hiddenParam: should also show/)
    ).toBeInTheDocument();
  });
});
