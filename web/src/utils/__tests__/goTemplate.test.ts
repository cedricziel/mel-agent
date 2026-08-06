import { describe, it, expect } from 'vitest';
import {
  NO_VALUE,
  checkTemplate,
  formatValue,
  lookupPath,
  renderTemplate,
  suggestPaths,
} from '../goTemplate';

describe('checkTemplate', () => {
  it('accepts plain text', () => {
    expect(checkTemplate('just text')).toEqual({ previewable: true });
  });

  it('accepts field lookups with and without whitespace', () => {
    expect(checkTemplate('{{.input.name}}').previewable).toBe(true);
    expect(checkTemplate('{{ .vars.role }}').previewable).toBe(true);
  });

  it('flags unbalanced braces as invalid', () => {
    const result = checkTemplate('Hello {{.input.name');
    expect(result.previewable).toBe(false);
    expect(result.invalid).toBe(true);
    expect(result.message).toMatch(/Unclosed action/);
  });

  it('flags empty actions as invalid', () => {
    const result = checkTemplate('{{}}');
    expect(result.invalid).toBe(true);
  });

  it('marks control flow as unsupported but not invalid', () => {
    const result = checkTemplate('{{if .input.ok}}yes{{end}}');
    expect(result.previewable).toBe(false);
    expect(result.invalid).toBeUndefined();
    expect(result.message).toMatch(/does not support/);
  });

  it('marks pipelines as unsupported', () => {
    const result = checkTemplate('{{.input.name | upper}}');
    expect(result.previewable).toBe(false);
    expect(result.invalid).toBeUndefined();
  });
});

describe('lookupPath', () => {
  const ctx = { input: { user: { name: 'Alice' } }, vars: { role: 'admin' } };

  it('resolves nested paths', () => {
    expect(lookupPath(ctx, '.input.user.name')).toBe('Alice');
    expect(lookupPath(ctx, '.vars.role')).toBe('admin');
  });

  it('returns undefined for missing paths', () => {
    expect(lookupPath(ctx, '.input.missing')).toBeUndefined();
    expect(lookupPath(ctx, '.input.user.name.deeper')).toBeUndefined();
  });
});

describe('formatValue', () => {
  it('renders strings verbatim and objects as JSON', () => {
    expect(formatValue('hi')).toBe('hi');
    expect(formatValue({ a: 1 })).toBe('{"a":1}');
    expect(formatValue([1, 2])).toBe('[1,2]');
    expect(formatValue(3)).toBe('3');
    expect(formatValue(true)).toBe('true');
  });

  it('renders missing values the way Go does', () => {
    expect(formatValue(undefined)).toBe(NO_VALUE);
    expect(formatValue(null)).toBe(NO_VALUE);
  });
});

describe('renderTemplate', () => {
  const ctx = {
    input: { name: 'Alice', tags: ['a'] },
    vars: { role: 'admin' },
  };

  it('substitutes field lookups', () => {
    expect(renderTemplate('Hello, {{.input.name}}!', ctx)).toBe(
      'Hello, Alice!'
    );
  });

  it('substitutes variables and tolerates whitespace', () => {
    expect(renderTemplate('{{ .vars.role }}-{{.input.name}}', ctx)).toBe(
      'admin-Alice'
    );
  });

  it('renders whole objects as JSON', () => {
    expect(renderTemplate('{{.input}}', ctx)).toBe(
      '{"name":"Alice","tags":["a"]}'
    );
  });

  it('renders missing lookups as <no value>', () => {
    expect(renderTemplate('{{.input.nope}}', ctx)).toBe(NO_VALUE);
  });

  it('leaves unsupported actions untouched', () => {
    expect(renderTemplate('{{if .x}}y{{end}}', ctx)).toBe('{{if .x}}y{{end}}');
  });
});

describe('suggestPaths', () => {
  it('lists top-level and nested paths', () => {
    const paths = suggestPaths({
      input: { user: { name: 'Alice' } },
      vars: { role: 'admin' },
    });
    expect(paths).toContain('.input');
    expect(paths).toContain('.input.user');
    expect(paths).toContain('.input.user.name');
    expect(paths).toContain('.vars.role');
  });

  it('stops at the configured depth', () => {
    const paths = suggestPaths({ input: { a: { b: { c: 1 } } } }, 2);
    expect(paths).toContain('.input.a');
    expect(paths).not.toContain('.input.a.b');
  });

  it('does not descend into arrays', () => {
    const paths = suggestPaths({ input: { tags: ['x'] } });
    expect(paths).toContain('.input.tags');
    expect(paths).not.toContain('.input.tags.0');
  });
});
