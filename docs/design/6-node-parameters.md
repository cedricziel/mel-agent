 # 6. Node Parameter Configuration – Design Document

 _Last updated: 2025‑04‑18_

 This document proposes an evolution to our node metadata model to support:
 1. Parameter groups
 2. Required vs. optional parameters
 3. Conditional visibility via CEL expressions
 4. Centralized metadata-driven validation and UI rendering

 ## 1. Motivation

 - Flat `defaults` maps become unwieldy as nodes gain options.
 - Some parameters only apply in certain modes or configurations.
 - Need a single source of truth for:
   - Builder form rendering
   - Validation rules
   - Documentation/tooltips
   - Client & server consistency

 ## 2. New Metadata Model

 Replace `Defaults map[string]string` with a slice of `ParameterDefinition`:

 ```go
 type ParameterDefinition struct {
   Name                string   // key in node.Data
   Label               string   // user-facing label
   Type                string   // "string", "number", "boolean", "enum", "json"
   Required            bool     // must be provided (non-empty)
   Default             any      // default value
   Group               string   // logical grouping in UI (e.g. "Request", "Advanced")
   VisibilityCondition string   // CEL expression for conditional display
   Options             []string // for enum types
   // Validators apply validation rules to the field
   Validators          []ValidatorSpec // list of validator specifications
   Description         string   // help text or tooltip
 }
```
 
### 2.1.1 Validators

`Validators` is an ordered list of `ValidatorSpec` entries (see 2.1.2) specifying validation rules to apply to the field.
Each `ValidatorSpec` references a named validator and optional parameters (e.g. regex patterns).
Validation is performed in sequence; upon the first failure, the field is marked invalid and the save/test is blocked.

This allows declarative, reusable validation logic on both client and server sides.

 Extend `NodeType`:

 ```go
 type NodeType struct {
   Type       string
   Label      string
   Category   string
   EntryPoint bool
   Branching  bool
   Parameters []ParameterDefinition
 }
 ```

### 2.1.2 ValidatorSpec

Validator specifications allow parameterized validators:
```go
// ValidatorSpec defines a validation rule for a parameter.
type ValidatorSpec struct {
  Type   string                 // validator name, e.g. "notEmpty", "url", "regex"
  Params map[string]interface{} // optional parameters, e.g. {"pattern": "^\\d+$"} for regex
}
```
 Validators are applied in order; the first failing validator blocks save/test.

### 2.1.3 Validator Definitions

We maintain a registry of available validators with metadata and optional parameter schemas:
```go
// ValidatorDefinition describes a named validation rule.
type ValidatorDefinition struct {
  Type        string         // unique validator name, e.g. "notEmpty", "url"
  Label       string         // display label for UI
  Description string         // help text or tooltip
  ParamSchema *JSONSchema    // optional JSON Schema for Params
}

// Example registry:
var Validators = []ValidatorDefinition{
  {Type: "notEmpty", Label: "Not Empty", Description: "Value must not be empty"},
  {Type: "url", Label: "URL", Description: "Must be a valid URL"},
  {Type: "json", Label: "JSON", Description: "Must be a valid JSON object"},
  {Type: "regex", Label: "Pattern", Description: "Must match a regex", ParamSchema: &JSONSchema{Type: "object", Properties: map[string]JSONSchema{"pattern": {Type: "string"}}}},
}
```
Client and server can validate `ValidatorSpec.Params` against `ValidatorDefinition.ParamSchema` before applying the rule.

 ## 2.2 Exportable Schema

 We support exporting each `NodeType` and its `Parameters` as a JSON Schema document. The schema includes:
 - `properties` for each parameter (`type`, `description`, `default`, etc.)
 - A `required` array listing parameters with `Required=true`
 - `enum` constraints for parameters of type `enum`
 - Default values from `Default`
 - Conditional visibility rules via `if/then` clauses driven by `VisibilityCondition`

 By providing JSON Schema for every node type, external tools can:
 - Generate node configuration forms automatically (e.g., React JSONSchema Form)
 - Perform offline validation of node configuration
 - Produce documentation or editors from standard JSON Schema tooling

 ## 3. Visibility via CEL

 - Use [CEL (Common Expression Language)](https://github.com/google/cel-go) to write expressions on current node data:
   - Examples: `mode == "sync"`, `method == "POST" && body != ""`
 - Builder evaluates these on-the-fly to show/hide fields.

 ## 4. Required Parameter Enforcement

 - UI marks required parameters with an asterisk.
 - On save/test, run validation:
   - Fields where `Required && empty` are flagged inline.
 - Block save until all required fields are filled.

 ## 5. Example: HTTP Request Node

 ```jsonc
 {
   "type": "http_request",
   "label": "HTTP Request",
   "category": "Integration",
   "parameters": [
     {"name":"url","label":"URL","type":"string","required":true,"default":"","group":"Request","description":"Endpoint to call"},
     {"name":"method","label":"Method","type":"enum","options":["GET","POST","PUT","DELETE"],"required":true,"default":"GET","group":"Request"},
     {"name":"headers","label":"Headers","type":"json","required":false,"default":"{}","group":"Request"},
     {"name":"body","label":"Body","type":"string","required":false,"default":"","group":"Request","visibilityCondition":"method != 'GET'"},
     {"name":"timeout","label":"Timeout (sec)","type":"number","required":false,"default":30,"group":"Advanced"}
   ]
 }
```

 ## 6. UI Changes

 - **Form Sections**: Collapse by `Group` (e.g., "Request", "Advanced").
 - **Conditional Rendering**: Hide fields when `VisibilityCondition` evaluates false.
 - **Validation**: Show inline errors under each invalid field.
 - **Save Button**: Disable until validation passes.

 ## 7. Migration Path

 1. Interpret existing `Defaults` as parameters with no groups or conditions.
 2. Gradually replace built‑in `Defaults` with explicit `Parameters`.
 3. Keep backward compatibility in the API (auto-generate `Parameters` from `Defaults`).

 ## 8. Next Steps

 - Evaluate and integrate a CEL library on client + server.
 - Refactor `NodeDefinition.Meta()` to return the new `Parameters` slice.
 - Update Builder UI to consume `Parameters` for rendering.
 - Add unit tests for visibility logic and validation errors.

 ## 9. Schema Export & Tooling

 - **JSON Schema Generation**: Provide an API endpoint `/api/node-types/schema/{type}` returning a JSON Schema for each node type, describing properties, required fields, defaults, and conditional rules.
 - **Schema Registry**: Maintain versioned schemas for node types to support backward compatibility and evolution.
 - **Tooling Integration**: Leverage JSON Schema form generators (e.g., React JSONSchema Form) to build standalone node configuration editors outside the core builder.
 - **Auto-Documentation**: Generate up-to-date API and UI docs from the JSON Schemas, ensuring consistent reference for all node types.

 This design centralizes node parameter metadata, supports dynamic UIs, and improves developer ergonomics.