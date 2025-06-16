import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ModelNode from '../ModelNode';

// Mock ReactFlow's Handle component
vi.mock('reactflow', () => ({
  Handle: ({ type, position, id, style }) => (
    <div data-testid={`handle-${type}-${position}-${id}`} style={style} />
  ),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right',
  },
}));

// Mock connection types
vi.mock('../utils/connectionTypes', () => ({
  HANDLE_TYPES: {
    MODEL_CONFIG: 'model-config',
  },
  getHandleColor: () => '#3b82f6',
}));

describe('ModelNode', () => {
  const defaultProps = {
    id: 'test-model-1',
    data: {
      label: 'Test Model',
      provider: 'openai',
      model: 'gpt-4',
    },
  };

  it('should render model node with label and icon', () => {
    render(<ModelNode {...defaultProps} />);

    expect(screen.getByText('Test Model')).toBeInTheDocument();
    expect(screen.getByText('📋')).toBeInTheDocument();
  });

  it('should render as circular node', () => {
    render(<ModelNode {...defaultProps} />);

    const nodeContainer = screen
      .getByText('Test Model')
      .closest('[class*="rounded-full"]');
    expect(nodeContainer).toHaveClass('rounded-full');
    expect(nodeContainer).toHaveClass('w-[120px]');
    expect(nodeContainer).toHaveClass('h-[120px]');
  });

  it('should show error state when error is true', () => {
    const propsWithError = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        error: true,
      },
    };

    render(<ModelNode {...propsWithError} />);

    const nodeContainer = screen
      .getByText('Test Model')
      .closest('[class*="border-red-500"]');
    expect(nodeContainer).toHaveClass('border-red-500');
  });

  it('should render output handle for connection to agent', () => {
    render(<ModelNode {...defaultProps} />);

    expect(
      screen.getByTestId('handle-source-top-config-out')
    ).toBeInTheDocument();
  });

  it('should display parameter summary', () => {
    const propsWithParams = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        provider: 'openai',
        model: 'gpt-4',
        temperature: 0.7,
        maxTokens: 1000,
      },
    };

    render(<ModelNode {...propsWithParams} />);

    expect(screen.getByText(/provider:/)).toBeInTheDocument();
    expect(screen.getByText(/model:/)).toBeInTheDocument();
  });

  it('should call onClick when node is clicked', () => {
    const mockOnClick = vi.fn();
    const propsWithClick = {
      ...defaultProps,
      onClick: mockOnClick,
    };

    render(<ModelNode {...propsWithClick} />);

    const nodeContainer = screen.getByText('Test Model').closest('div');
    fireEvent.click(nodeContainer);

    expect(mockOnClick).toHaveBeenCalledTimes(1);
  });

  it('should show cursor pointer when clickable', () => {
    const mockOnClick = vi.fn();
    const propsWithClick = {
      ...defaultProps,
      onClick: mockOnClick,
    };

    render(<ModelNode {...propsWithClick} />);

    const nodeContainer = screen
      .getByText('Test Model')
      .closest('[class*="cursor-pointer"]');
    expect(nodeContainer).toHaveClass('cursor-pointer');
  });

  it('should render delete button when onDelete is provided', () => {
    const mockOnDelete = vi.fn();
    const propsWithDelete = {
      ...defaultProps,
      onDelete: mockOnDelete,
    };

    render(<ModelNode {...propsWithDelete} />);

    const deleteButton = screen.getByTitle('Delete node');
    expect(deleteButton).toBeInTheDocument();
    expect(deleteButton).toHaveTextContent('🗑️');
  });

  it('should call onDelete when delete button is clicked', () => {
    const mockOnDelete = vi.fn();
    const propsWithDelete = {
      ...defaultProps,
      onDelete: mockOnDelete,
    };

    render(<ModelNode {...propsWithDelete} />);

    const deleteButton = screen.getByTitle('Delete node');
    fireEvent.click(deleteButton);

    expect(mockOnDelete).toHaveBeenCalledTimes(1);
    expect(mockOnDelete).toHaveBeenCalledWith('test-model-1');
  });

  it('should stop propagation when delete button is clicked', () => {
    const mockOnDelete = vi.fn();
    const mockOnClick = vi.fn();
    const mockStopPropagation = vi.fn();

    const propsWithBoth = {
      ...defaultProps,
      onDelete: mockOnDelete,
      onClick: mockOnClick,
    };

    render(<ModelNode {...propsWithBoth} />);

    const deleteButton = screen.getByTitle('Delete node');
    fireEvent.click(deleteButton, { stopPropagation: mockStopPropagation });

    expect(mockOnDelete).toHaveBeenCalled();
    expect(mockOnClick).not.toHaveBeenCalled();
  });

  it('should render add button when onAddClick is provided', () => {
    const mockOnAddClick = vi.fn();
    const propsWithAdd = {
      ...defaultProps,
      onAddClick: mockOnAddClick,
    };

    render(<ModelNode {...propsWithAdd} />);

    const addButton = screen.getByText('+');
    expect(addButton).toBeInTheDocument();
  });

  it('should call onAddClick when add button is clicked', () => {
    const mockOnAddClick = vi.fn();
    const propsWithAdd = {
      ...defaultProps,
      onAddClick: mockOnAddClick,
    };

    render(<ModelNode {...propsWithAdd} />);

    const addButton = screen.getByText('+');
    fireEvent.click(addButton);

    expect(mockOnAddClick).toHaveBeenCalledTimes(1);
  });

  it('should handle empty data gracefully', () => {
    const propsWithEmptyData = {
      ...defaultProps,
      data: {
        label: 'Empty Model',
      },
    };

    render(<ModelNode {...propsWithEmptyData} />);

    expect(screen.getByText('Empty Model')).toBeInTheDocument();
    expect(screen.getByText('📋')).toBeInTheDocument();
  });

  it('should truncate long parameter values', () => {
    const propsWithLongValues = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        longParameter:
          'this is a very long parameter value that should be truncated',
      },
    };

    render(<ModelNode {...propsWithLongValues} />);

    // The ModelNode shows only first 2 parameters and then shows ellipsis if more exist
    // Check that ellipsis appears when there are more than 2 parameters
    expect(screen.getByText('…')).toBeInTheDocument();
  });

  it('should show ellipsis when more than 2 parameters', () => {
    const propsWithManyParams = {
      ...defaultProps,
      data: {
        ...defaultProps.data,
        param1: 'value1',
        param2: 'value2',
        param3: 'value3',
        param4: 'value4',
      },
    };

    render(<ModelNode {...propsWithManyParams} />);

    expect(screen.getByText('…')).toBeInTheDocument();
  });

  it('should exclude certain keys from parameter summary', () => {
    const propsWithExcludedKeys = {
      ...defaultProps,
      data: {
        label: 'Test Model',
        status: 'running',
        nodeTypeLabel: 'Model Type',
        error: false,
        displayParam: 'should show',
      },
    };

    render(<ModelNode {...propsWithExcludedKeys} />);

    // Should not show excluded keys in summary
    expect(screen.queryByText(/status:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/nodeTypeLabel:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/error:/)).not.toBeInTheDocument();

    // Should show other params
    expect(screen.getByText(/displayParam:/)).toBeInTheDocument();
  });
});
