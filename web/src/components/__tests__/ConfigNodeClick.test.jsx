import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import OpenAIModelNode from '../OpenAIModelNode';
import AnthropicModelNode from '../AnthropicModelNode';
import LocalMemoryNode from '../LocalMemoryNode';
import ToolsNode from '../ToolsNode';
import MemoryNode from '../MemoryNode';
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
    TOOLS_CONFIG: 'tools-config',
    MEMORY_CONFIG: 'memory-config',
  },
  getHandleColor: () => '#3b82f6',
}));

describe('Configuration Node Click Functionality', () => {
  const defaultData = {
    label: 'Test Config',
    provider: 'test',
  };

  const defaultProps = {
    id: 'test-config-1',
    data: defaultData,
  };

  describe('OpenAIModelNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<OpenAIModelNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = screen
        .getByText('OpenAI')
        .closest('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });

    it('should have cursor pointer when onClick is provided', () => {
      const mockOnClick = vi.fn();

      render(<OpenAIModelNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = screen
        .getByText('OpenAI')
        .closest('[class*="cursor-pointer"]');
      expect(nodeContainer).toHaveClass('cursor-pointer');
    });

    it('should not prevent delete button clicks from propagating to onClick', () => {
      const mockOnClick = vi.fn();
      const mockOnDelete = vi.fn();

      render(
        <OpenAIModelNode
          {...defaultProps}
          onClick={mockOnClick}
          onDelete={mockOnDelete}
        />
      );

      const deleteButton = screen.getByTitle('Delete node');
      fireEvent.click(deleteButton);

      expect(mockOnDelete).toHaveBeenCalledTimes(1);
      expect(mockOnClick).not.toHaveBeenCalled();
    });
  });

  describe('AnthropicModelNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<AnthropicModelNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = screen
        .getByText('Anthropic')
        .closest('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('LocalMemoryNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<LocalMemoryNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = screen
        .getByText('Local Memory')
        .closest('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('ToolsNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<ToolsNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = document.querySelector('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('MemoryNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<MemoryNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = document.querySelector('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('ModelNode', () => {
    it('should call onClick when node is clicked', () => {
      const mockOnClick = vi.fn();

      render(<ModelNode {...defaultProps} onClick={mockOnClick} />);

      const nodeContainer = document.querySelector('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);

      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('Click Handler Integration', () => {
    it('should support different click handlers for different node types', () => {
      const mockOpenAIClick = vi.fn();
      const mockAnthropicClick = vi.fn();
      const mockMemoryClick = vi.fn();

      const { rerender } = render(
        <OpenAIModelNode {...defaultProps} onClick={mockOpenAIClick} />
      );
      fireEvent.click(
        screen.getByText('OpenAI').closest('[class*="cursor-pointer"]')
      );
      expect(mockOpenAIClick).toHaveBeenCalledTimes(1);

      rerender(
        <AnthropicModelNode {...defaultProps} onClick={mockAnthropicClick} />
      );
      fireEvent.click(
        screen.getByText('Anthropic').closest('[class*="cursor-pointer"]')
      );
      expect(mockAnthropicClick).toHaveBeenCalledTimes(1);

      rerender(<LocalMemoryNode {...defaultProps} onClick={mockMemoryClick} />);
      fireEvent.click(
        screen.getByText('Local Memory').closest('[class*="cursor-pointer"]')
      );
      expect(mockMemoryClick).toHaveBeenCalledTimes(1);
    });

    it('should handle missing onClick gracefully', () => {
      expect(() => {
        render(<OpenAIModelNode {...defaultProps} />);
        const nodeContainer = screen.getByText('OpenAI').closest('div');
        fireEvent.click(nodeContainer);
      }).not.toThrow();
    });

    it('should maintain other functionality while adding click support', () => {
      const mockOnClick = vi.fn();
      const mockOnDelete = vi.fn();
      const mockOnAddClick = vi.fn();

      render(
        <OpenAIModelNode
          {...defaultProps}
          onClick={mockOnClick}
          onDelete={mockOnDelete}
          onAddClick={mockOnAddClick}
        />
      );

      // All buttons should be present and functional
      expect(screen.getByTitle('Delete node')).toBeInTheDocument();
      expect(screen.getByText('+')).toBeInTheDocument();

      // Click handlers should work independently
      fireEvent.click(screen.getByTitle('Delete node'));
      expect(mockOnDelete).toHaveBeenCalledTimes(1);
      expect(mockOnClick).not.toHaveBeenCalled();

      fireEvent.click(screen.getByText('+'));
      expect(mockOnAddClick).toHaveBeenCalledTimes(1);
      expect(mockOnClick).not.toHaveBeenCalled();

      // Main node click should work
      const nodeContainer = screen
        .getByText('OpenAI')
        .closest('[class*="cursor-pointer"]');
      fireEvent.click(nodeContainer);
      expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('Visual Feedback', () => {
    it('should add cursor pointer for clickable nodes', () => {
      const configComponents = [
        OpenAIModelNode,
        AnthropicModelNode,
        LocalMemoryNode,
        ToolsNode,
        MemoryNode,
        ModelNode,
      ];

      configComponents.forEach((Component) => {
        const mockOnClick = vi.fn();
        const { unmount } = render(
          <Component {...defaultProps} onClick={mockOnClick} />
        );

        const nodeContainer = document.querySelector(
          '[class*="cursor-pointer"]'
        );
        expect(nodeContainer).toBeInTheDocument();
        expect(nodeContainer).toHaveClass('cursor-pointer');

        unmount();
      });
    });

    it('should not show cursor pointer when not clickable', () => {
      render(<OpenAIModelNode {...defaultProps} />);

      // Without onClick, should not have cursor pointer
      const nodeContainer = screen.getByText('OpenAI').closest('div');
      expect(nodeContainer).not.toHaveClass('cursor-pointer');
    });
  });
});
