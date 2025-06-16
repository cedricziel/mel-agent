import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import WorkflowToolbar from '../WorkflowToolbar';

describe('WorkflowToolbar', () => {
  const defaultProps = {
    isDraft: false,
    isDirty: false,
    isSaving: false,
    lastSaved: null,
    saveError: null,
    viewMode: 'editor',
    testing: false,
    isLiveMode: false,
    sidebarTab: null,
    onViewModeChange: vi.fn(),
    onAddNode: vi.fn(),
    onSave: vi.fn(),
    onTestRun: vi.fn(),
    onAutoLayout: vi.fn(),
    onToggleSidebar: vi.fn(),
    onToggleLiveMode: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render all toolbar buttons', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    expect(screen.getByText('Editor')).toBeInTheDocument();
    expect(screen.getByText('Executions')).toBeInTheDocument();
    expect(screen.getByText('+ Add Node')).toBeInTheDocument();
    expect(screen.getByText('Save')).toBeInTheDocument();
    expect(screen.getByText('Test Run')).toBeInTheDocument();
    expect(screen.getByText('Auto Layout')).toBeInTheDocument();
    expect(screen.getByText('💬 Chat')).toBeInTheDocument();
    expect(screen.getByText('Edit Mode')).toBeInTheDocument();
  });

  it('should show draft status when isDraft is true', () => {
    render(<WorkflowToolbar {...defaultProps} isDraft={true} />);

    expect(screen.getByText('Draft')).toBeInTheDocument();
    expect(screen.getByText('Deploy')).toBeInTheDocument();
  });

  it('should show deployed status when isDraft is false', () => {
    render(<WorkflowToolbar {...defaultProps} isDraft={false} />);

    expect(screen.getByText('Deployed')).toBeInTheDocument();
    expect(screen.getByText('Save')).toBeInTheDocument();
  });

  it('should show saving status when isSaving is true', () => {
    render(
      <WorkflowToolbar {...defaultProps} isDraft={true} isSaving={true} />
    );

    expect(screen.getByText('Saving...')).toBeInTheDocument();
  });

  it('should show last saved time when available', () => {
    const lastSaved = new Date('2023-01-01T12:00:00Z').toISOString();
    render(
      <WorkflowToolbar {...defaultProps} isDraft={true} lastSaved={lastSaved} />
    );

    expect(screen.getByText(/Saved/)).toBeInTheDocument();
  });

  it('should show save error when present', () => {
    render(
      <WorkflowToolbar
        {...defaultProps}
        isDraft={true}
        saveError="Network error"
      />
    );

    expect(screen.getByText('Save failed')).toBeInTheDocument();
  });

  it('should disable buttons in executions mode', () => {
    render(<WorkflowToolbar {...defaultProps} viewMode="executions" />);

    expect(screen.getByText('+ Add Node')).toBeDisabled();
    expect(screen.getByText('Save')).toBeDisabled();
    expect(screen.getByText('Test Run')).toBeDisabled();
    expect(screen.getByText('Auto Layout')).toBeDisabled();
  });

  it('should show live mode when isLiveMode is true', () => {
    render(<WorkflowToolbar {...defaultProps} isLiveMode={true} />);

    expect(screen.getByText('Live Mode')).toBeInTheDocument();
  });

  it('should show testing state when testing is true', () => {
    render(<WorkflowToolbar {...defaultProps} testing={true} />);

    expect(screen.getByText('Running...')).toBeInTheDocument();
  });

  it('should call onViewModeChange when editor/executions toggle is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('Executions'));
    expect(defaultProps.onViewModeChange).toHaveBeenCalledWith('executions');

    fireEvent.click(screen.getByText('Editor'));
    expect(defaultProps.onViewModeChange).toHaveBeenCalledWith('editor');
  });

  it('should call onAddNode when add node button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('+ Add Node'));
    expect(defaultProps.onAddNode).toHaveBeenCalled();
  });

  it('should call onSave when save button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} isDirty={true} />);

    fireEvent.click(screen.getByText('Save'));
    expect(defaultProps.onSave).toHaveBeenCalled();
  });

  it('should call onTestRun when test run button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('Test Run'));
    expect(defaultProps.onTestRun).toHaveBeenCalled();
  });

  it('should call onAutoLayout when auto layout button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('Auto Layout'));
    expect(defaultProps.onAutoLayout).toHaveBeenCalled();
  });

  it('should call onToggleSidebar when chat button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('💬 Chat'));
    expect(defaultProps.onToggleSidebar).toHaveBeenCalledWith('chat');
  });

  it('should call onToggleLiveMode when live mode button is clicked', () => {
    render(<WorkflowToolbar {...defaultProps} />);

    fireEvent.click(screen.getByText('Edit Mode'));
    expect(defaultProps.onToggleLiveMode).toHaveBeenCalled();
  });

  it('should highlight active sidebar tab', () => {
    render(<WorkflowToolbar {...defaultProps} sidebarTab="chat" />);

    const chatButton = screen.getByText('💬 Chat');
    expect(chatButton).toHaveClass('bg-blue-500', 'text-white');
  });

  it('should highlight active view mode', () => {
    render(<WorkflowToolbar {...defaultProps} viewMode="executions" />);

    const executionsButton = screen.getByText('Executions');
    expect(executionsButton).toHaveClass(
      'bg-white',
      'text-gray-900',
      'shadow-sm'
    );
  });

  it('should enable save button when dirty or draft', () => {
    const { rerender } = render(
      <WorkflowToolbar {...defaultProps} isDirty={true} />
    );
    expect(screen.getByText('Save')).not.toBeDisabled();

    rerender(<WorkflowToolbar {...defaultProps} isDraft={true} />);
    expect(screen.getByText('Deploy')).not.toBeDisabled();
  });

  it('should disable save button when not dirty and not draft', () => {
    render(
      <WorkflowToolbar {...defaultProps} isDirty={false} isDraft={false} />
    );
    expect(screen.getByText('Save')).toBeDisabled();
  });

  it('should handle deploy action for drafts', async () => {
    render(<WorkflowToolbar {...defaultProps} isDraft={true} />);

    fireEvent.click(screen.getByText('Deploy'));
    expect(defaultProps.onSave).toHaveBeenCalled();
  });
});
