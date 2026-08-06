import { useMemo, useRef, useState } from 'react';
import {
  checkTemplate,
  renderTemplate,
  suggestPaths,
  type TemplateContext,
} from '../utils/goTemplate';

interface TemplateEditorProps {
  value: string;
  onChange: (value: string) => void;
  readOnly?: boolean;
  placeholder?: string;
  rows?: number;
  /** Sample data used for the preview; defaults to a small example payload. */
  initialSample?: TemplateContext;
}

const DEFAULT_SAMPLE: TemplateContext = {
  input: { name: 'Alice', count: 2 },
  vars: { role: 'admin' },
};

/**
 * Editor for Go template expressions with an inline preview.
 *
 * The preview renders a supported subset of the template against editable
 * sample data so authors can see the shape of the output without executing the
 * workflow. The server remains the authority on the actual rendering.
 */
export default function TemplateEditor({
  value,
  onChange,
  readOnly = false,
  placeholder = 'Hello, {{.input.name}}!',
  rows = 6,
  initialSample = DEFAULT_SAMPLE,
}: TemplateEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [sampleText, setSampleText] = useState(() =>
    JSON.stringify(initialSample, null, 2)
  );

  const { sample, sampleError } = useMemo(() => {
    try {
      const parsed = JSON.parse(sampleText);
      if (parsed === null || typeof parsed !== 'object') {
        return { sample: {}, sampleError: 'Sample data must be a JSON object' };
      }
      return { sample: parsed as TemplateContext, sampleError: null };
    } catch {
      return { sample: {}, sampleError: 'Sample data is not valid JSON' };
    }
  }, [sampleText]);

  const check = useMemo(() => checkTemplate(value || ''), [value]);
  const preview = useMemo(() => {
    if (!check.previewable || sampleError) return '';
    return renderTemplate(value || '', sample);
  }, [check.previewable, sampleError, value, sample]);

  const paths = useMemo(() => suggestPaths(sample), [sample]);

  /** Inserts a template action at the caret, or appends it when unfocused. */
  const insertPath = (path: string) => {
    if (readOnly) return;
    const snippet = `{{${path}}}`;
    const textarea = textareaRef.current;
    const current = value || '';
    if (!textarea) {
      onChange(current + snippet);
      return;
    }
    const start = textarea.selectionStart ?? current.length;
    const end = textarea.selectionEnd ?? current.length;
    const next = current.slice(0, start) + snippet + current.slice(end);
    onChange(next);
    requestAnimationFrame(() => {
      const caret = start + snippet.length;
      textarea.focus();
      textarea.setSelectionRange(caret, caret);
    });
  };

  return (
    <div className="space-y-2">
      <textarea
        ref={textareaRef}
        rows={rows}
        value={value || ''}
        readOnly={readOnly}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        aria-label="Template expression"
        className={`w-full border rounded px-2 py-1 font-mono text-xs ${
          check.invalid ? 'border-red-500' : 'border-gray-300'
        }`}
      />

      {check.message && (
        <div
          className={`text-xs ${check.invalid ? 'text-red-600' : 'text-amber-600'}`}
          role={check.invalid ? 'alert' : 'status'}
        >
          {check.message}
        </div>
      )}

      {paths.length > 0 && !readOnly && (
        <div className="flex flex-wrap gap-1">
          {paths.map((path) => (
            <button
              key={path}
              type="button"
              onClick={() => insertPath(path)}
              title={`Insert {{${path}}}`}
              className="text-xs font-mono px-1.5 py-0.5 bg-gray-100 hover:bg-gray-200 border border-gray-300 rounded"
            >
              {path}
            </button>
          ))}
        </div>
      )}

      <details className="text-xs">
        <summary className="cursor-pointer text-gray-600">Sample input</summary>
        <textarea
          rows={5}
          value={sampleText}
          onChange={(e) => setSampleText(e.target.value)}
          aria-label="Sample input"
          className={`mt-1 w-full border rounded px-2 py-1 font-mono text-xs ${
            sampleError ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {sampleError && (
          <div className="text-xs text-red-600 mt-1" role="alert">
            {sampleError}
          </div>
        )}
      </details>

      <div>
        <div className="text-xs text-gray-600 mb-1">Preview</div>
        {check.previewable && !sampleError ? (
          <pre
            data-testid="template-preview"
            className="whitespace-pre-wrap break-words bg-gray-50 border border-gray-200 rounded px-2 py-1 font-mono text-xs text-gray-800"
          >
            {preview}
          </pre>
        ) : (
          <div className="text-xs text-gray-500 italic">
            Preview unavailable
          </div>
        )}
      </div>
    </div>
  );
}
