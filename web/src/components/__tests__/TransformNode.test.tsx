import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import TransformNode from '../TransformNode';

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

describe('TransformNode', () => {
  const defaultProps = {
    id: 'transform-1',
    data: {
      label: 'Greeting',
      nodeTypeLabel: 'Transform',
      expression: 'Hello, {{.input.name}}!',
    },
  };

  it('renders the label and the configured expression', () => {
    render(<TransformNode {...defaultProps} />);

    expect(screen.getByText('Greeting')).toBeInTheDocument();
    expect(screen.getByText('Transform')).toBeInTheDocument();
    expect(screen.getByText('Hello, {{.input.name}}!')).toBeInTheDocument();
  });

  it('shows a placeholder when no expression is configured', () => {
    render(<TransformNode {...defaultProps} data={{ label: 'Greeting' }} />);

    expect(screen.getByText('no expression set')).toBeInTheDocument();
  });

  it('renders workflow input and output handles', () => {
    render(<TransformNode {...defaultProps} />);

    expect(
      screen.getByTestId('handle-target-left-workflow-in')
    ).toBeInTheDocument();
    expect(
      screen.getByTestId('handle-source-right-workflow-out')
    ).toBeInTheDocument();
  });

  it('uses the icon from the node definition when provided', () => {
    render(<TransformNode {...defaultProps} icon="✨" />);

    expect(screen.getByText('✨')).toBeInTheDocument();
  });

  it('marks the node when it has an error', () => {
    render(
      <TransformNode
        {...defaultProps}
        data={{ ...defaultProps.data, error: true }}
      />
    );

    const container = screen.getByText('Greeting').closest('.relative');
    expect(container).toHaveClass('border-red-500');
  });

  it('shows a running indicator', () => {
    render(
      <TransformNode
        {...defaultProps}
        data={{ ...defaultProps.data, status: 'running' }}
      />
    );

    expect(document.querySelector('.animate-pulse')).toBeInTheDocument();
  });

  it('calls onDelete with the node id', () => {
    const onDelete = vi.fn();
    render(<TransformNode {...defaultProps} onDelete={onDelete} />);

    fireEvent.click(screen.getByTitle('Delete node'));

    expect(onDelete).toHaveBeenCalledWith('transform-1');
  });

  it('calls onAddClick from the quick-add button', () => {
    const onAddClick = vi.fn();
    render(<TransformNode {...defaultProps} onAddClick={onAddClick} />);

    fireEvent.click(screen.getByText('+'));

    expect(onAddClick).toHaveBeenCalledTimes(1);
  });
});
