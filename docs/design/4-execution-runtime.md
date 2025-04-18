 # Execution Runtime – Design Document

 Building on our data-flow and builder designs (see **3-data-flow.md**, **2-builder.md**), this document outlines a safe, low-overhead execution environment for agent workflows.

 ## 1. Goals
 1. **Sandboxed execution**: run arbitrary node logic (built-in or user-supplied) without risking host stability or security.
 2. **Low latency & footprint**: minimal startup time, small memory/CPU overhead to support high throughput.
 3. **Polyglot support**: enable future user scripts in multiple languages via a common compilation target.
 4. **Observability & control**: trace, log, and meter node invocations; enforce resource quotas.

 ## 2. Background & Constraints
 - Current stub executor (Go HTTP handler) simply echoes input.
 - We want to evolve to a real runtime: built-in nodes in Go + user code in a confined environment.
 - Inspired by Windmill (uses WebAssembly) and other Wasm-first platforms.
 - ExecutionPayload schema (see **3-data-flow.md**) defines the JSON envelope passed into each node.

 ## 3. High-Level Architecture
 ```
       ┌──────────────────────┐
       │  Agent Runner (Go)   │
       │  • Graph dispatch    │
       │  • State & context   │
       │  • Observability     │
       └─────────┬────────────┘
                 │
       ┌─────────┴────────────┐
       │   Node Execution     │
       │   ┌───────────────┐  │
       │   │ Built-in Go    │  │
       │   │ executors      │  │
       │   └───────────────┘  │
       │   ┌───────────────┐  │
       │   │ Wasm Sandbox   │◄─┤ compiled modules from user scripts
       │   └───────────────┘  │
       └─────────────────────┘
 ```

 ### 3.1 Host Process (Go)
 - Orchestrates graph traversal: loads `ExecutionPayload`, invokes each node in sequence/parallel.
 - Manages **context**, merges outputs, handles branching.
 - Collects spans, metrics, logs around node calls.

 ### 3.2 Built-in Executors
 - HTTP request, If/Branch, Cron trigger, default transformations, LLM calls.
 - Continues as Go functions until fully replaced by Wasm in the long term.

 ### 3.3 WASM Sandbox
- Users author scripts in high-level languages (e.g., JavaScript, TypeScript, Python), which are compiled to WebAssembly by the system.
 - Embedded engine (e.g. Wasmtime or Wasmer) loads modules once, caches compiled artifacts.
 - Modules export a standard `_run(payloadPtr, length) -> resultPtr` entrypoint.

 ## 4. WebAssembly Sandbox Details
 ### 4.1 Compilation Targets
 - **JavaScript / TypeScript**: user scripts are compiled (e.g., via AssemblyScript or swc-based toolchain) into Wasm modules.
 - **Python**: user scripts compiled via Pyodide or similar Python-Wasm tool.
 - **Go**: advanced users can author modules in Go, compiled via TinyGo.
 - **Rust**: optional support for Rust modules targeting WASI.

 ### 4.2 Host Functions (imports)
  | Namespace | Function               | Purpose                           |
  |-----------|------------------------|-----------------------------------|
  | env       | `get_payload(ptr,len)` | Read serialized ExecutionPayload. |
  | env       | `set_payload(ptr,len)` | Write new ExecutionPayload.       |
  | secrets   | `get_secret(name)`     | Fetch connection/webhook secrets. |
  | http      | `fetch(ptr,len)->ptr`  | Perform HTTP/HTTPS requests.      |
  | state     | `get(key)/set(key,val)`| Shared context store.             |
  | log       | `log(level,ptr,len)`   | Emit structured logs.             |
  | metrics   | `metric(name,val)`     | Increment or record metrics.      |

 ### 4.3 Resource Limits
 - **Memory**: per-module cap (e.g. 64 MB).
 - **CPU**: instruction or fuel limits to enforce timeouts.
 - **Isolation**: no direct file I/O or syscalls outside defined imports.

 ### 4.4 Caching & Pooling
 - **Compile cache**: modules are JIT-compiled once per version, reused across runs.
 - **Instance pool**: keep warm Wasm instances to amortize instantiation cost.

 ## 5. Data Flow & Serialization
 - Payloads serialized to JSON and passed via shared memory between host and Wasm.
 - For performance, consider binary encoding (MessagePack) in later phases.
 - Host unmarshals result and continues graph execution.

 ## 6. Observability & Telemetry
 - **Tracing**: host spans around Wasm calls, recording durations and outcome.
 - **Logs**: captured via `log()` import and funneled into centralized logging with `runId`/`nodeId` tags.
 - **Metrics**: track Wasm compile times, execution latency, memory usage.

 ## 7. Security & Auditing
 - **ACLs**: restrict which modules can call certain host functions (e.g. HTTP).
 - **Sandbox escape**: rely on mature Wasm engines with proven isolation.
 - **Audit logs**: record module names, versions, and function calls for compliance.

 ## 8. Phased Integration Plan
 1. **Phase 1**: integrate Wasm engine, run simple arithmetic modules.
 2. **Phase 2**: migrate “Code” node to Wasm sandbox with host API.
 3. **Phase 3**: refactor built-in nodes as Wasm modules.
 4. **Phase 4**: expose CLI tool for authors to compile scripts to Wasm.

 ## 9. Open Questions
 1. Should we support **streaming** Wasm (e.g. for LLM token streams)?
 2. What **binary format** for high-volume payloads?
 3. How to enable in-editor debugging/step-through for Wasm modules?
 4. Versioning & migration of host API for backward compatibility.

 _Last updated: 2025-04-18_