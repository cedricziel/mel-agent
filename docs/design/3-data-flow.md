 # Data Payload & State Propagation – Design Document

 This document describes the structure of the “data object” that flows between nodes in an agent workflow, drawing inspiration from platforms like Node‑RED, n8n, and Make.com.

 ## 1. Goals

 1. **Unified format** – every node consumes and produces the same envelope, easing plumbing and reuse.
 2. **Rich payload** – support arbitrary JSON, binary attachments, metadata, and error contexts.
 3. **Branching & splitting** – nodes may output multiple items or branch on conditions.
 4. **Global context** – allow shared variables/state across the run.
 5. **Observability** – record per‑node inputs/outputs for debugging and replay.

 ## 2. Patterns in Existing Systems

 ### 2.1 Node‑RED
 - Uses a single `msg` object.
 - `msg.payload` carries the main data; nodes may attach additional properties on `msg`.
 - Supports arrays, splits, and merges by manually constructing arrays in `msg.payload`.

 ### 2.2 n8n
 - Uses an `items: INodeExecutionData[]` array.
 - Each `INodeExecutionData` has `{ json: any; binary?: IBinaryData; pairedItem: number }`.
 - Nodes receive all `items[]` and return a new `items[]`, enabling splitting and merging.

 ### 2.3 Make.com
 - Similar “bundle” concept: each bundle carries `variables`, `data`, `error` fields.
 - Supports iterators and aggregated collections as first‑class constructs.

 ## 3. Proposed Data Model

 We introduce a unified `ExecutionPayload` envelope:

 ```ts
 interface Item {
   id: string;                   // unique per item in this run
   data: any;                    // arbitrary JSON payload
   binary?: Record<string, Blob>; // optional binary attachments
   error?: {                     // if node execution failed
     message: string;
     nodeId: string;
     stack?: string;
   };
 }

 interface ExecutionPayload {
   runId: string;                // UUID for this workflow execution
   items: Item[];                // current items flowing into the node
   context: Record<string, any>; // shared key/value store across nodes
   meta: {
     startTime: string;          // ISO timestamp
     lastNode?: string;          // most recently executed nodeId
     [key: string]: any;
   };
 }
 ```

 ### Usage
 1. **Trigger** node constructs an initial `ExecutionPayload`:
    - `items`: typically a single item, e.g. `{ id: "1", data: payloadFromTrigger }`.
    - `context`: empty or seeded with credentials.
    - `meta.startTime`: set now.
 2. Each **Skill** node receives the envelope, transforms `items` and/or `context`, and returns a new envelope:
    - E.g. an LLM node maps each item to a new `data` object.
    - An “If” node filters items into two branches via `error` or by emitting only the passing items.
 3. **Branching** is handled by storing `condition` on items or by routing to different edges (using nodeType handles).  The runtime inspects each item’s `error` or a custom flag to dispatch items to the correct downstream nodes.

 ## 4. JSON Schema & Validation

 We define a JSON Schema for `ExecutionPayload` to validate at each step:

 ```jsonc
 {
   "$id": "#/definitions/ExecutionPayload",
   "type": "object",
   "required": ["runId","items","context","meta"],
   "properties": {
     "runId": { "type": "string", "format": "uuid" },
     "items": {
       "type": "array",
       "items": { "$ref": "#/definitions/Item" }
     },
     "context": { "type": "object" },
     "meta": { "type": "object" }
   },
   "definitions": {
     "Item": {
       "type": "object",
       "required": ["id","data"],
       "properties": {
         "id": { "type": "string" },
         "data": {},
         "binary": { "type": "object" },
         "error": { "$ref": "#/definitions/NodeError" }
       }
     },
     "NodeError": {
       "type": "object",
       "required": ["message","nodeId"],
       "properties": {
         "message": { "type": "string" },
         "nodeId": { "type": "string" }
       }
     }
   }
 }
 ```

 ## 5. Observability, Tracing & Metrics

 We instrument every workflow execution end‑to‑end, capturing logs, traces, and metrics by default:

 - **Distributed Tracing**: Each run has a root span (`runId`) and each node execution is a child span, with attributes:
   - `nodeId`, `nodeType`, `status` (`ok`/`error`), `itemCount`, etc.
   - Start/end timestamps and duration.
   - Integrated via OpenTelemetry (OTel) spans exported to tracing backends (Jaeger, Zipkin, etc.).

 - **Metrics**: Runtime exports Prometheus‑style metrics:
   - Counters: `agent_runs_started_total`, `agent_runs_completed_total`, `agent_node_executions_total{nodeType, status}`.
   - Histograms: `agent_node_duration_seconds{nodeType}` to track per‑node latency distributions.
   - Gauges: `agent_runs_inflight` for active runs, `agent_items_queue_length` for pending items.

 - **Structured Logs**: Every node logs at INFO/DEBUG/ERROR with a common schema:
   ```json
   {
     "ts": "2025-04-18T12:34:56Z",
     "level": "INFO",
     "runId": "uuid",
     "nodeId": "node-123",
     "nodeType": "llm",
     "message": "Node execution started",
     "payload": { /* redacted snapshot */ }
   }
   ```
   Logs integrate with centralized logging (Loki, ELK, etc.) and correlate via `runId` and `nodeId`.

 - **Replay & Debug**: Recorded traces and logs power the builder’s debug sidebar:
   - Step through spans to inspect `ExecutionPayload` before/after each node.
   - Replay tokens/partial outputs for streaming nodes.

 This full‑stack observability ensures every agent run is traceable, measurable, and debuggable without extra configuration.

 ## 6. Open Questions & Next Steps

 1. **Streaming vs Batch** – do some nodes (LLM) need to emit partial responses per token?
 2. **Stateful long‑running flows** – how to persist `context` and resume across minutes/hours?
 3. **Backpressure** – for large `items[]`, do we need pagination or chunked processing?
 4. **Security** – sanitize `data` before passing between nodes when using untrusted code.

 _Last updated: 2025‑04‑18_