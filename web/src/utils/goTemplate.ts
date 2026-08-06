/**
 * Client-side approximation of the Go text/template subset used by the
 * Transform node. It exists purely to power the live preview in the builder —
 * the authoritative rendering always happens on the server.
 *
 * Supported: field lookups such as {{.input}}, {{.input.user.name}} and
 * {{.vars.role}}, with optional surrounding whitespace.
 * Unsupported: actions with control flow ({{if}}, {{range}}, {{with}},
 * {{template}}, {{block}}) and pipelines. Those make the preview unavailable
 * rather than producing a misleading result.
 */

/** Go's text/template output for a lookup that resolves to nothing. */
export const NO_VALUE = '<no value>';

const ACTION_RE = /{{([^{}]*)}}/g;
const FIELD_RE = /^\.[A-Za-z_$][\w$]*(\.[A-Za-z_$][\w$]*)*$|^\.$/;
const CONTROL_KEYWORDS = [
  'if',
  'else',
  'end',
  'range',
  'with',
  'template',
  'block',
  'define',
];

export type TemplateContext = Record<string, unknown>;

export interface TemplateCheck {
  /** Whether the template can be previewed client-side. */
  previewable: boolean;
  /** Reason the template is not previewable, or a syntax problem. */
  message?: string;
  /** True when the template is definitely malformed (not merely unsupported). */
  invalid?: boolean;
}

/**
 * Inspects a template for syntax errors and for constructs the preview cannot
 * evaluate.
 */
export function checkTemplate(template: string): TemplateCheck {
  const opens = (template.match(/{{/g) || []).length;
  const closes = (template.match(/}}/g) || []).length;
  if (opens !== closes) {
    return {
      previewable: false,
      invalid: true,
      message: 'Unclosed action: every {{ needs a matching }}',
    };
  }

  const actions = [...template.matchAll(ACTION_RE)].map((m) => m[1].trim());
  for (const action of actions) {
    if (action === '') {
      return {
        previewable: false,
        invalid: true,
        message: 'Empty action: {{ }} needs an expression',
      };
    }
    const keyword = action.split(/\s+/)[0].replace(/^-/, '');
    if (CONTROL_KEYWORDS.includes(keyword)) {
      return {
        previewable: false,
        message: `Preview does not support {{${keyword}}} — the workflow will still render it`,
      };
    }
    if (!FIELD_RE.test(action)) {
      return {
        previewable: false,
        message: `Preview does not support the expression "${action}"`,
      };
    }
  }

  return { previewable: true };
}

/** Resolves a dotted path such as ".input.user.name" against the context. */
export function lookupPath(
  context: TemplateContext,
  path: string
): unknown | undefined {
  const segments = path.split('.').filter(Boolean);
  let current: unknown = context;
  for (const segment of segments) {
    if (current === null || typeof current !== 'object') return undefined;
    current = (current as Record<string, unknown>)[segment];
    if (current === undefined) return undefined;
  }
  return current;
}

/** Formats a resolved value the way the preview should display it. */
export function formatValue(value: unknown): string {
  if (value === undefined || value === null) return NO_VALUE;
  if (typeof value === 'string') return value;
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}

/**
 * Renders the supported subset of a template. Callers should run
 * {@link checkTemplate} first; unsupported actions are left untouched here.
 */
export function renderTemplate(
  template: string,
  context: TemplateContext
): string {
  return template.replace(ACTION_RE, (match, rawExpr: string) => {
    const expr = rawExpr.trim();
    if (!FIELD_RE.test(expr)) return match;
    if (expr === '.') return formatValue(context);
    return formatValue(lookupPath(context, expr));
  });
}

/**
 * Collects the insertable template paths for a sample context, e.g.
 * [".input", ".input.name", ".vars.role"]. Nested objects are walked up to
 * `maxDepth` levels so the suggestion list stays manageable.
 */
export function suggestPaths(context: TemplateContext, maxDepth = 3): string[] {
  const paths: string[] = [];

  const walk = (value: unknown, prefix: string, depth: number) => {
    paths.push(prefix);
    if (depth >= maxDepth) return;
    if (value === null || typeof value !== 'object' || Array.isArray(value)) {
      return;
    }
    for (const key of Object.keys(value as Record<string, unknown>)) {
      walk(
        (value as Record<string, unknown>)[key],
        `${prefix}.${key}`,
        depth + 1
      );
    }
  };

  for (const key of Object.keys(context)) {
    walk(context[key], `.${key}`, 1);
  }

  return paths;
}
