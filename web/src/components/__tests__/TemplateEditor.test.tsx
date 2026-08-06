import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { useState } from 'react';
import TemplateEditor from '../TemplateEditor';

/** Wrapper that keeps the editor controlled, like the config panel does. */
function ControlledEditor({ initial = '', ...rest }) {
  const [value, setValue] = useState(initial);
  return <TemplateEditor value={value} onChange={setValue} {...rest} />;
}

describe('TemplateEditor', () => {
  it('renders the current expression', () => {
    render(
      <TemplateEditor value="Hello, {{.input.name}}!" onChange={vi.fn()} />
    );

    expect(screen.getByLabelText('Template expression')).toHaveValue(
      'Hello, {{.input.name}}!'
    );
  });

  it('calls onChange when the expression is edited', () => {
    const onChange = vi.fn();
    render(<TemplateEditor value="" onChange={onChange} />);

    fireEvent.change(screen.getByLabelText('Template expression'), {
      target: { value: '{{.input.name}}' },
    });

    expect(onChange).toHaveBeenCalledWith('{{.input.name}}');
  });

  it('previews the rendered output against the sample data', () => {
    render(
      <TemplateEditor value="Hello, {{.input.name}}!" onChange={vi.fn()} />
    );

    expect(screen.getByTestId('template-preview')).toHaveTextContent(
      'Hello, Alice!'
    );
  });

  it('previews workflow variables', () => {
    render(<TemplateEditor value="{{.vars.role}}" onChange={vi.fn()} />);

    expect(screen.getByTestId('template-preview')).toHaveTextContent('admin');
  });

  it('flags an unclosed action as an error', () => {
    render(<TemplateEditor value="Hello, {{.input.name" onChange={vi.fn()} />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unclosed action');
    expect(screen.queryByTestId('template-preview')).not.toBeInTheDocument();
    expect(screen.getByText('Preview unavailable')).toBeInTheDocument();
  });

  it('warns without erroring for unsupported control flow', () => {
    render(
      <TemplateEditor value="{{if .input.name}}x{{end}}" onChange={vi.fn()} />
    );

    expect(screen.getByRole('status')).toHaveTextContent('does not support');
    expect(screen.getByText('Preview unavailable')).toBeInTheDocument();
  });

  it('inserts a path when its chip is clicked', () => {
    render(<ControlledEditor />);

    fireEvent.click(screen.getByTitle('Insert {{.input.name}}'));

    expect(screen.getByLabelText('Template expression')).toHaveValue(
      '{{.input.name}}'
    );
  });

  it('inserts a path at the caret position', () => {
    render(<ControlledEditor initial="Hi !" />);

    const textarea = screen.getByLabelText(
      'Template expression'
    ) as HTMLTextAreaElement;
    textarea.focus();
    textarea.setSelectionRange(3, 3);

    fireEvent.click(screen.getByTitle('Insert {{.input.name}}'));

    expect(textarea).toHaveValue('Hi {{.input.name}}!');
  });

  it('re-renders the preview when the sample data changes', () => {
    render(<TemplateEditor value="{{.input.name}}" onChange={vi.fn()} />);

    fireEvent.change(screen.getByLabelText('Sample input'), {
      target: { value: JSON.stringify({ input: { name: 'Bob' } }) },
    });

    expect(screen.getByTestId('template-preview')).toHaveTextContent('Bob');
  });

  it('reports invalid sample JSON instead of previewing', () => {
    render(<TemplateEditor value="{{.input.name}}" onChange={vi.fn()} />);

    fireEvent.change(screen.getByLabelText('Sample input'), {
      target: { value: '{not json' },
    });

    expect(screen.getByRole('alert')).toHaveTextContent('not valid JSON');
    expect(screen.queryByTestId('template-preview')).not.toBeInTheDocument();
  });

  it('hides insert chips and blocks edits when read-only', () => {
    render(
      <TemplateEditor value="{{.input.name}}" onChange={vi.fn()} readOnly />
    );

    expect(screen.getByLabelText('Template expression')).toHaveAttribute(
      'readonly'
    );
    expect(
      screen.queryByTitle('Insert {{.input.name}}')
    ).not.toBeInTheDocument();
  });

  it('accepts custom sample data', () => {
    render(
      <TemplateEditor
        value="{{.input.city}}"
        onChange={vi.fn()}
        initialSample={{ input: { city: 'Berlin' } }}
      />
    );

    expect(screen.getByTestId('template-preview')).toHaveTextContent('Berlin');
  });
});
