import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import NodeConfigurationPanel from '../NodeConfigurationPanel';

// Mock CodeEditor component
vi.mock('../CodeEditor', () => ({
  default: function MockCodeEditor({ value, onChange, placeholder }) {
    return (
      <textarea
        data-testid="code-editor"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    );
  },
}));

describe('NodeConfigurationPanel', () => {
  const mockNode = {
    id: 'node-1',
    data: {
      label: 'Test Node',
      param1: 'value1',
    },
  };

  const mockNodeDef = {
    type: 'test-node',
    label: 'Test Node Type',
    parameters: [
      {
        name: 'param1',
        label: 'Parameter 1',
        type: 'string',
        required: true,
        description: 'First parameter',
      },
      {
        name: 'param2',
        label: 'Parameter 2',
        type: 'enum',
        options: ['option1', 'option2'],
      },
      {
        name: 'param3',
        label: 'Parameter 3',
        type: 'boolean',
      },
    ],
  };

  const mockNodes = [
    { id: 'node-2', type: 'http_request', data: { label: 'HTTP Node' } },
    { id: 'node-3', type: 'manual_trigger', data: { label: 'Trigger Node' } },
  ];

  const defaultProps = {
    node: mockNode,
    nodeDef: mockNodeDef,
    nodes: mockNodes,
    currentFormData: mockNode.data,
    dynamicOptions: {},
    loadingOptions: {},
    credentials: {},
    handleChange: vi.fn(),
    viewMode: 'editor',
    selectedExecution: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render node name field in editor mode', () => {
    render(<NodeConfigurationPanel {...defaultProps} />);

    expect(screen.getByText('Node Name')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Node')).toBeInTheDocument();
  });

  it('should render all parameters in editor mode', () => {
    render(<NodeConfigurationPanel {...defaultProps} />);

    expect(screen.getByText('Parameter 1')).toBeInTheDocument();
    expect(screen.getByText('Parameter 2')).toBeInTheDocument();
    // Use getAllByText to handle multiple elements with same text
    expect(screen.getAllByText('Parameter 3')).toHaveLength(2); // Label and span
  });

  it('should handle node name changes', () => {
    const handleChange = vi.fn();
    render(
      <NodeConfigurationPanel {...defaultProps} handleChange={handleChange} />
    );

    const nameInput = screen.getByDisplayValue('Test Node');
    fireEvent.change(nameInput, { target: { value: 'New Node Name' } });

    expect(handleChange).toHaveBeenCalledWith('label', 'New Node Name');
  });

  it('should handle string parameter changes', () => {
    const handleChange = vi.fn();
    render(
      <NodeConfigurationPanel {...defaultProps} handleChange={handleChange} />
    );

    const stringInput = screen.getByDisplayValue('value1');
    fireEvent.change(stringInput, { target: { value: 'new value' } });

    expect(handleChange).toHaveBeenCalledWith('param1', 'new value');
  });

  it('should handle enum parameter changes', () => {
    const handleChange = vi.fn();
    render(
      <NodeConfigurationPanel {...defaultProps} handleChange={handleChange} />
    );

    const enumSelect = screen.getByRole('combobox');
    fireEvent.change(enumSelect, { target: { value: 'option2' } });

    expect(handleChange).toHaveBeenCalledWith('param2', 'option2');
  });

  it('should handle boolean parameter changes', () => {
    const handleChange = vi.fn();
    render(
      <NodeConfigurationPanel {...defaultProps} handleChange={handleChange} />
    );

    const booleanCheckbox = screen.getByRole('checkbox');
    fireEvent.click(booleanCheckbox);

    expect(handleChange).toHaveBeenCalledWith('param3', true);
  });

  it('should render code editor for code parameters', () => {
    const nodeDefWithCode = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'codeParam',
          label: 'Code Parameter',
          type: 'string',
          jsonSchema: { format: 'code' },
        },
      ],
    };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithCode}
        currentFormData={{ codeParam: 'console.log("test");' }}
      />
    );

    expect(screen.getByTestId('code-editor')).toBeInTheDocument();
  });

  it('should render credential selector', () => {
    const nodeDefWithCredential = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'credParam',
          label: 'Credential Parameter',
          type: 'credential',
        },
      ],
    };

    const credentials = {
      credParam: [{ id: 'cred-1', name: 'Test Cred', integration_name: 'API' }],
    };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithCredential}
        credentials={credentials}
      />
    );

    expect(screen.getByText('Select Credential')).toBeInTheDocument();
    expect(screen.getByText('Test Cred (API)')).toBeInTheDocument();
  });

  it('should render node reference selector', () => {
    const nodeDefWithNodeRef = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'nodeRef',
          label: 'Node Reference',
          type: 'nodeReference',
        },
      ],
    };

    render(
      <NodeConfigurationPanel {...defaultProps} nodeDef={nodeDefWithNodeRef} />
    );

    expect(screen.getByText('Select Node')).toBeInTheDocument();
    expect(screen.getByText('HTTP Node (http_request)')).toBeInTheDocument();
    // Trigger nodes should be excluded
    expect(
      screen.queryByText('Trigger Node (manual_trigger)')
    ).not.toBeInTheDocument();
  });

  it('should render dynamic options selector', () => {
    const nodeDefWithDynamic = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'dynamicParam',
          label: 'Dynamic Parameter',
          type: 'string',
          dynamicOptions: true,
        },
      ],
    };

    const dynamicOptions = {
      dynamicParam: [
        { value: 'opt1', label: 'Option 1' },
        { value: 'opt2', label: 'Option 2' },
      ],
    };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithDynamic}
        dynamicOptions={dynamicOptions}
      />
    );

    expect(screen.getByText('Select Dynamic Parameter')).toBeInTheDocument();
    expect(screen.getByText('Option 1')).toBeInTheDocument();
    expect(screen.getByText('Option 2')).toBeInTheDocument();
  });

  it('should show loading state for dynamic options', () => {
    const nodeDefWithDynamic = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'dynamicParam',
          label: 'Dynamic Parameter',
          type: 'string',
          dynamicOptions: true,
        },
      ],
    };

    const loadingOptions = { dynamicParam: true };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithDynamic}
        loadingOptions={loadingOptions}
      />
    );

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('should show error styling for required fields', () => {
    const currentFormData = { param1: '' }; // Empty required field

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        currentFormData={currentFormData}
      />
    );

    const requiredInput = screen.getByPlaceholderText('First parameter');
    expect(requiredInput).toHaveClass('border-red-500');
  });

  it('should render read-only mode in executions view', () => {
    const selectedExecution = {
      id: 'exec-1',
      created_at: '2023-01-01T00:00:00Z',
    };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        viewMode="executions"
        selectedExecution={selectedExecution}
      />
    );

    expect(screen.getByText('Test Node')).toBeInTheDocument();
    expect(screen.getByText('Test Node Type')).toBeInTheDocument();
    expect(screen.getByText('1/1/2023, 1:00:00 AM')).toBeInTheDocument();

    // Should not have editable inputs
    expect(screen.queryByLabelText('Node Name')).not.toBeInTheDocument();
  });

  it('should handle parameter descriptions', () => {
    render(<NodeConfigurationPanel {...defaultProps} />);

    expect(screen.getByText('First parameter')).toBeInTheDocument();
  });

  it('should use default values for parameters', () => {
    const nodeDefWithDefaults = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'paramWithDefault',
          label: 'Parameter with Default',
          type: 'string',
          default: 'default value',
        },
      ],
    };

    const currentFormData = {}; // No value set

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithDefaults}
        currentFormData={currentFormData}
      />
    );

    expect(screen.getByDisplayValue('default value')).toBeInTheDocument();
  });

  it('should handle parameterType fallback', () => {
    const nodeDefWithParameterType = {
      ...mockNodeDef,
      parameters: [
        {
          name: 'param',
          label: 'Parameter',
          parameterType: 'string', // Using parameterType instead of type
        },
      ],
    };

    render(
      <NodeConfigurationPanel
        {...defaultProps}
        nodeDef={nodeDefWithParameterType}
      />
    );

    expect(screen.getByText('Parameter')).toBeInTheDocument();
  });
});
