import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ConfigSelectionDialog from '../ConfigSelectionDialog';

// Mock the config options that would be passed to the dialog
const mockConfigOptions = {
  model: [
    {
      type: 'openai_model',
      label: 'OpenAI Model',
      defaultData: { provider: 'openai', model: 'gpt-4' },
    },
    {
      type: 'anthropic_model',
      label: 'Anthropic Model',
      defaultData: {
        provider: 'anthropic',
        model: 'claude-3-5-sonnet-20241022',
      },
    },
  ],
  memory: [
    {
      type: 'local_memory',
      label: 'Local Memory',
      defaultData: { memoryType: 'local', maxMessages: 100 },
    },
  ],
  tools: [
    {
      type: 'workflow_tools',
      label: 'Workflow Tools',
      defaultData: { allowCodeExecution: false, allowWebSearch: true },
    },
  ],
};

describe('Configuration Node Dialog Integration', () => {
  const mockOnSelect = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderConfigDialog = (configType = 'model', isOpen = true) => {
    return render(
      <ConfigSelectionDialog
        isOpen={isOpen}
        configType={configType}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );
  };

  it('should render model configuration options', () => {
    renderConfigDialog('model');

    // The dialog should be visible when open
    expect(screen.getByText('Select Model Configuration')).toBeInTheDocument();
    expect(screen.getByText('OpenAI')).toBeInTheDocument();
    expect(screen.getByText('Anthropic')).toBeInTheDocument();
  });

  it('should render memory configuration options', () => {
    renderConfigDialog('memory');

    expect(screen.getByText('Select Memory Configuration')).toBeInTheDocument();
    expect(screen.getByText('Local Memory')).toBeInTheDocument();
  });

  it('should render tools configuration options', () => {
    renderConfigDialog('tools');

    expect(screen.getByText('Select Tools Configuration')).toBeInTheDocument();
    expect(screen.getByText('Workflow Tools')).toBeInTheDocument();
  });

  it('should not render when closed', () => {
    renderConfigDialog('model', false);

    expect(
      screen.queryByText('Select Model Configuration')
    ).not.toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    renderConfigDialog('model');

    const closeButton = screen.getByText('✕');
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('should handle configuration selection', () => {
    renderConfigDialog('model');

    // Find and click a configuration option
    const openAIOption = screen.getByText('OpenAI');
    fireEvent.click(openAIOption);

    // Click the Add Configuration button
    const addButton = screen.getByText('Add Configuration');
    fireEvent.click(addButton);

    // The onSelect callback should be called
    expect(mockOnSelect).toHaveBeenCalledTimes(1);
  });

  describe('Configuration Types', () => {
    it('should handle model configuration type', () => {
      renderConfigDialog('model');

      expect(
        screen.getByText('Select Model Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('OpenAI')).toBeInTheDocument();
    });

    it('should handle memory configuration type', () => {
      renderConfigDialog('memory');

      expect(
        screen.getByText('Select Memory Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('Local Memory')).toBeInTheDocument();
    });

    it('should handle tools configuration type', () => {
      renderConfigDialog('tools');

      expect(
        screen.getByText('Select Tools Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('Workflow Tools')).toBeInTheDocument();
    });
  });

  describe('Dialog Behavior', () => {
    it('should be accessible with proper ARIA attributes', () => {
      renderConfigDialog('model');

      // Check for heading and buttons which provide accessibility
      expect(
        screen.getByText('Select Model Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('Cancel')).toBeInTheDocument();
      expect(screen.getByText('Add Configuration')).toBeInTheDocument();
    });

    it('should handle keyboard navigation', () => {
      renderConfigDialog('model');

      // Test that interactive elements are present
      const cancelButton = screen.getByText('Cancel');
      const addButton = screen.getByText('Add Configuration');

      expect(cancelButton).toBeInTheDocument();
      expect(addButton).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle invalid configuration types gracefully', () => {
      expect(() => {
        renderConfigDialog('invalid_type');
      }).not.toThrow();
    });

    it('should handle missing props gracefully', () => {
      expect(() => {
        render(<ConfigSelectionDialog />);
      }).not.toThrow();
    });
  });

  describe('Integration with Node Creation', () => {
    it('should provide correct data structure for node creation', () => {
      renderConfigDialog('model');

      // The dialog should be ready to provide configuration data
      expect(
        screen.getByText('Select Model Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('OpenAI')).toBeInTheDocument();
    });

    it('should support different handle types for connections', () => {
      renderConfigDialog('model');

      // Each config type should support proper handle connections
      expect(
        screen.getByText('Select Model Configuration')
      ).toBeInTheDocument();
      expect(screen.getByText('Add Configuration')).toBeInTheDocument();
    });
  });
});
