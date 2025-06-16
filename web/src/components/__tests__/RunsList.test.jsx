import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import RunsList from '../RunsList';

// Wrapper component for router context
const RouterWrapper = ({ children }) => (
  <BrowserRouter>{children}</BrowserRouter>
);

describe('RunsList', () => {
  const mockOnRunSelect = vi.fn();
  const mockAgentId = 'agent-123';
  const mockRuns = [
    { id: 'run-1', created_at: '2024-01-01T10:00:00Z' },
    { id: 'run-2', created_at: '2024-01-02T10:00:00Z' },
    { id: 'run-3', created_at: '2024-01-03T10:00:00Z' },
  ];

  const defaultProps = {
    agentId: mockAgentId,
    runs: mockRuns,
    selectedRunID: null,
    onRunSelect: mockOnRunSelect,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render agent title with correct ID', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    expect(
      screen.getByText(`Runs for Agent ${mockAgentId}`)
    ).toBeInTheDocument();
  });

  it('should render all runs in the list', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    mockRuns.forEach((run) => {
      expect(screen.getByText(run.created_at)).toBeInTheDocument();
    });
  });

  it('should call onRunSelect when a run is clicked', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    const firstRunButton = screen.getByText(mockRuns[0].created_at);
    fireEvent.click(firstRunButton);

    expect(mockOnRunSelect).toHaveBeenCalledWith('run-1');
    expect(mockOnRunSelect).toHaveBeenCalledTimes(1);
  });

  it('should highlight selected run with correct styling', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} selectedRunID="run-2" />
      </RouterWrapper>
    );

    const selectedRunButton = screen.getByText(mockRuns[1].created_at);
    const unselectedRunButton = screen.getByText(mockRuns[0].created_at);

    expect(selectedRunButton).toHaveClass('bg-gray-200');
    expect(unselectedRunButton).toHaveClass('hover:bg-gray-100');
    expect(unselectedRunButton).not.toHaveClass('bg-gray-200');
  });

  it('should render back to builder link with correct href', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    const backLink = screen.getByText('← Back to Builder');
    expect(backLink).toBeInTheDocument();
    expect(backLink.closest('a')).toHaveAttribute(
      'href',
      `/agents/${mockAgentId}/edit`
    );
  });

  it('should handle empty runs list', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} runs={[]} />
      </RouterWrapper>
    );

    expect(
      screen.getByText(`Runs for Agent ${mockAgentId}`)
    ).toBeInTheDocument();
    expect(screen.getByText('← Back to Builder')).toBeInTheDocument();

    // Should not render any run buttons
    expect(
      screen.queryByRole('button', { name: /2024/ })
    ).not.toBeInTheDocument();
  });

  it('should handle multiple run selections', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    // Click first run
    fireEvent.click(screen.getByText(mockRuns[0].created_at));
    expect(mockOnRunSelect).toHaveBeenCalledWith('run-1');

    // Click second run
    fireEvent.click(screen.getByText(mockRuns[1].created_at));
    expect(mockOnRunSelect).toHaveBeenCalledWith('run-2');

    expect(mockOnRunSelect).toHaveBeenCalledTimes(2);
  });

  it('should have correct container styling', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    const container = screen
      .getByText(`Runs for Agent ${mockAgentId}`)
      .closest('div');
    expect(container).toHaveClass(
      'w-1/4',
      'border-r',
      'p-4',
      'overflow-auto',
      'h-full'
    );
  });

  it('should have correct button styling for runs', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    const runButton = screen.getByText(mockRuns[0].created_at);
    expect(runButton).toHaveClass(
      'w-full',
      'text-left',
      'px-2',
      'py-1',
      'rounded',
      'hover:bg-gray-100'
    );
  });

  it('should handle runs with different date formats', () => {
    const runsWithDifferentDates = [
      { id: 'run-1', created_at: '2024-01-01' },
      { id: 'run-2', created_at: '2024-12-31T23:59:59.999Z' },
      { id: 'run-3', created_at: 'Invalid Date' },
    ];

    render(
      <RouterWrapper>
        <RunsList {...defaultProps} runs={runsWithDifferentDates} />
      </RouterWrapper>
    );

    runsWithDifferentDates.forEach((run) => {
      expect(screen.getByText(run.created_at)).toBeInTheDocument();
    });
  });

  it('should maintain selection state correctly', () => {
    const { rerender } = render(
      <RouterWrapper>
        <RunsList {...defaultProps} selectedRunID="run-1" />
      </RouterWrapper>
    );

    expect(screen.getByText(mockRuns[0].created_at)).toHaveClass('bg-gray-200');

    // Change selection
    rerender(
      <RouterWrapper>
        <RunsList {...defaultProps} selectedRunID="run-2" />
      </RouterWrapper>
    );

    expect(screen.getByText(mockRuns[0].created_at)).not.toHaveClass(
      'bg-gray-200'
    );
    expect(screen.getByText(mockRuns[1].created_at)).toHaveClass('bg-gray-200');
  });

  it('should handle rapid clicking without issues', () => {
    render(
      <RouterWrapper>
        <RunsList {...defaultProps} />
      </RouterWrapper>
    );

    const runButton = screen.getByText(mockRuns[0].created_at);

    // Click multiple times rapidly
    fireEvent.click(runButton);
    fireEvent.click(runButton);
    fireEvent.click(runButton);

    expect(mockOnRunSelect).toHaveBeenCalledTimes(3);
    expect(mockOnRunSelect).toHaveBeenCalledWith('run-1');
  });
});
