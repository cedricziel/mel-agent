import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import AddNodeModal from '../AddNodeModal';

describe('AddNodeModal', () => {
  const mockOnClose = vi.fn();
  const mockOnAddNode = vi.fn();

  const mockCategories = [
    {
      category: 'Triggers',
      types: [
        {
          type: 'webhook',
          label: 'Webhook',
          description: 'Receive HTTP requests',
        },
        {
          type: 'schedule',
          label: 'Schedule',
          description: 'Run on a schedule',
        },
      ],
    },
    {
      category: 'Actions',
      types: [
        {
          type: 'http_request',
          label: 'HTTP Request',
          description: 'Make HTTP requests',
        },
        {
          type: 'email',
          label: 'Send Email',
          description: 'Send email notifications',
        },
      ],
    },
    {
      category: 'AI',
      types: [
        {
          type: 'openai_model',
          label: 'OpenAI Model',
          description: 'Use OpenAI models',
        },
      ],
    },
  ];

  const defaultProps = {
    isOpen: true,
    categories: mockCategories,
    onClose: mockOnClose,
    onAddNode: mockOnAddNode,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when isOpen is false', () => {
    render(<AddNodeModal {...defaultProps} isOpen={false} />);

    expect(screen.queryByText('Add Node')).not.toBeInTheDocument();
  });

  it('should render modal when isOpen is true', () => {
    render(<AddNodeModal {...defaultProps} />);

    expect(screen.getByText('Add Node')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search nodes...')).toBeInTheDocument();
  });

  it('should render all categories and node types', () => {
    render(<AddNodeModal {...defaultProps} />);

    // Check categories
    expect(screen.getByText('Triggers')).toBeInTheDocument();
    expect(screen.getByText('Actions')).toBeInTheDocument();
    expect(screen.getByText('AI')).toBeInTheDocument();

    // Check node types
    expect(screen.getByText('Webhook')).toBeInTheDocument();
    expect(screen.getByText('Schedule')).toBeInTheDocument();
    expect(screen.getByText('HTTP Request')).toBeInTheDocument();
    expect(screen.getByText('Send Email')).toBeInTheDocument();
    expect(screen.getByText('OpenAI Model')).toBeInTheDocument();

    // Check descriptions
    expect(screen.getByText('Receive HTTP requests')).toBeInTheDocument();
    expect(screen.getByText('Run on a schedule')).toBeInTheDocument();
    expect(screen.getByText('Make HTTP requests')).toBeInTheDocument();
    expect(screen.getByText('Send email notifications')).toBeInTheDocument();
    expect(screen.getByText('Use OpenAI models')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    render(<AddNodeModal {...defaultProps} />);

    const closeButton = screen.getByText('✕');
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('should call onAddNode and onClose when node type is selected', () => {
    render(<AddNodeModal {...defaultProps} />);

    const webhookButton = screen.getByText('Webhook');
    fireEvent.click(webhookButton);

    expect(mockOnAddNode).toHaveBeenCalledWith('webhook');
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('should filter node types based on search input', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(searchInput, { target: { value: 'webhook' } });

    // Should show webhook
    expect(screen.getByText('Webhook')).toBeInTheDocument();
    expect(screen.getByText('Receive HTTP requests')).toBeInTheDocument();

    // Should hide other nodes
    expect(screen.queryByText('Schedule')).not.toBeInTheDocument();
    expect(screen.queryByText('HTTP Request')).not.toBeInTheDocument();
    expect(screen.queryByText('Send Email')).not.toBeInTheDocument();
    expect(screen.queryByText('OpenAI Model')).not.toBeInTheDocument();

    // Should still show the category if it has matching nodes
    expect(screen.getByText('Triggers')).toBeInTheDocument();
    // Should hide categories with no matching nodes
    expect(screen.queryByText('Actions')).not.toBeInTheDocument();
    expect(screen.queryByText('AI')).not.toBeInTheDocument();
  });

  it('should filter by node type as well as label', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(searchInput, { target: { value: 'http_request' } });

    // Should show HTTP Request node
    expect(screen.getByText('HTTP Request')).toBeInTheDocument();
    expect(screen.getByText('Make HTTP requests')).toBeInTheDocument();

    // Should hide other nodes
    expect(screen.queryByText('Webhook')).not.toBeInTheDocument();
    expect(screen.queryByText('Schedule')).not.toBeInTheDocument();
    expect(screen.queryByText('Send Email')).not.toBeInTheDocument();
    expect(screen.queryByText('OpenAI Model')).not.toBeInTheDocument();
  });

  it('should be case insensitive when searching', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(searchInput, { target: { value: 'WEBHOOK' } });

    expect(screen.getByText('Webhook')).toBeInTheDocument();
    expect(screen.queryByText('Schedule')).not.toBeInTheDocument();
  });

  it('should show no results when search matches nothing', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

    // Should not show any categories or nodes
    expect(screen.queryByText('Triggers')).not.toBeInTheDocument();
    expect(screen.queryByText('Actions')).not.toBeInTheDocument();
    expect(screen.queryByText('AI')).not.toBeInTheDocument();
    expect(screen.queryByText('Webhook')).not.toBeInTheDocument();
    expect(screen.queryByText('HTTP Request')).not.toBeInTheDocument();
  });

  it('should clear search and show all nodes when search is cleared', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');

    // First filter
    fireEvent.change(searchInput, { target: { value: 'webhook' } });
    expect(screen.queryByText('Schedule')).not.toBeInTheDocument();

    // Then clear
    fireEvent.change(searchInput, { target: { value: '' } });
    expect(screen.getByText('Schedule')).toBeInTheDocument();
    expect(screen.getByText('HTTP Request')).toBeInTheDocument();
    expect(screen.getByText('OpenAI Model')).toBeInTheDocument();
  });

  it('should handle categories with no matching types', () => {
    const categoriesWithEmptyResult = [
      {
        category: 'Empty Category',
        types: [
          {
            type: 'hidden_node',
            label: 'Hidden Node',
            description: 'This will be filtered out',
          },
        ],
      },
      {
        category: 'Visible Category',
        types: [
          {
            type: 'webhook',
            label: 'Webhook',
            description: 'This will be visible',
          },
        ],
      },
    ];

    render(
      <AddNodeModal {...defaultProps} categories={categoriesWithEmptyResult} />
    );

    const searchInput = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(searchInput, { target: { value: 'webhook' } });

    expect(screen.getByText('Visible Category')).toBeInTheDocument();
    expect(screen.getByText('Webhook')).toBeInTheDocument();
    expect(screen.queryByText('Empty Category')).not.toBeInTheDocument();
    expect(screen.queryByText('Hidden Node')).not.toBeInTheDocument();
  });

  it('should handle empty categories array', () => {
    render(<AddNodeModal {...defaultProps} categories={[]} />);

    expect(screen.getByText('Add Node')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search nodes...')).toBeInTheDocument();

    // Should not show any categories or nodes
    expect(screen.queryByText('Triggers')).not.toBeInTheDocument();
    expect(screen.queryByText('Actions')).not.toBeInTheDocument();
  });

  it('should handle node types without descriptions', () => {
    const categoriesWithoutDescriptions = [
      {
        category: 'Test Category',
        types: [
          {
            type: 'test_node',
            label: 'Test Node',
            // No description
          },
        ],
      },
    ];

    render(
      <AddNodeModal
        {...defaultProps}
        categories={categoriesWithoutDescriptions}
      />
    );

    expect(screen.getByText('Test Node')).toBeInTheDocument();
    // Should not crash when description is missing
  });

  it('should maintain search state when typing', () => {
    render(<AddNodeModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search nodes...');

    fireEvent.change(searchInput, { target: { value: 'web' } });
    expect(searchInput.value).toBe('web');

    fireEvent.change(searchInput, { target: { value: 'webhook' } });
    expect(searchInput.value).toBe('webhook');
  });

  it('should have correct modal styling and structure', () => {
    render(<AddNodeModal {...defaultProps} />);

    // Check modal backdrop
    const backdrop = screen.getByText('Add Node').closest('.fixed');
    expect(backdrop).toHaveClass('inset-0', 'bg-black', 'bg-opacity-50');

    // Check modal content
    const modal = screen.getByText('Add Node').closest('.bg-white');
    expect(modal).toHaveClass('bg-white', 'rounded-lg', 'p-6');

    // Check search input styling
    const searchInput = screen.getByPlaceholderText('Search nodes...');
    expect(searchInput).toHaveClass(
      'w-full',
      'border',
      'rounded',
      'px-3',
      'py-2'
    );
  });

  it('should handle rapid clicking on node types', () => {
    render(<AddNodeModal {...defaultProps} />);

    const webhookButton = screen.getByText('Webhook');

    // Click multiple times rapidly
    fireEvent.click(webhookButton);
    fireEvent.click(webhookButton);
    fireEvent.click(webhookButton);

    // Should call for each click (modal doesn't prevent rapid clicking)
    expect(mockOnAddNode).toHaveBeenCalledTimes(3);
    expect(mockOnAddNode).toHaveBeenCalledWith('webhook');
    expect(mockOnClose).toHaveBeenCalledTimes(3);
  });
});
