<!-- File: docs/design/7-agent-node.md -->
# 7. Agent Node – Memory, Tools & Context Protocol

## 1. Purpose & Scope

This document specifies the **Agent Node**, a standalone execution unit in the AI Agents runtime. An Agent Node:
  - Maintains both short‑term and long‑term memory
  - Can invoke external and internal tools
  - Integrates with Model Context Protocol servers (MCPs) to fetch and manage context spans
  - Produces coherent outputs and updates its memory stores

This design focuses on the **runtime behavior**, **memory management**, and **tool/context integration** of each node.

## 2. Goals & Non‑Goals

Goals:
  1. **Responsive Short‑Term Memory (STM)** for multi‑turn reasoning
  2. **Durable Long‑Term Memory (LTM)** for persistent knowledge
  3. **Pluggable Tool Interface** with safe invocation
  4. **Model Context Protocol (MCP) Clients** – support multiple MCP servers both inside and outside the system
  5. **Transparent Decision Trace** for audit/logging

Non‑Goals:
  - Global orchestration of multi‑agent workflows (covered elsewhere)
  - UI/UX or API surface design (see 0‑agents.md)

## 3. High‑Level Architecture

```ascii
[ INPUT ] → Preprocessor → STM Retriever ↔ LTM Retriever
               ↓         ↓                ↓
             Planner → Context Manager ⇄ MCP Servers
               ↓         ↓
         Tool Dispatcher ↔ Tools (APIs, DBs, services)
               ↓
            Integrator → STM Updater → LTM Updater
               ↓
            [ OUTPUT ]
```

## 4. Component Breakdown

4.1 Input Preprocessor
  - Normalize inbound messages (text, JSON, events)
  - Extract metadata (user, timestamp, intent)

4.2 Memory Subsystem
  4.2.1 Short‑Term Memory (STM)
    - In‑memory buffer or Redis cache
    - Capacity limit (last N turns or token budget)
    - Fast lookups by recency or tags
  4.2.2 Long‑Term Memory (LTM)
    - Persistent store (vector DB + document DB)
    - Embeddings, facts, logs
    - Semantic search, temporal queries

4.3 Model Context Protocol (MCP) Clients
  - Standard gRPC/HTTP protocols to fetch context spans (embeddings, token windows)
  - Support for:
    • **Internal MCP Servers** – co‑located with the agent runtime for low‑latency access
    • **External MCP Servers** – third‑party or specialized context providers (e.g., domain knowledge bases)
  - Configurable per‑agent or per‑run, with ACLs and rate limits

4.4 Planner & Reasoner
  - Coordinates retrieval from STM, LTM, MCP
  - Generates sub‑task plans, decides on tool vs direct LLM invocation
  - Emits a structured “thought trace” (for logging and optional user inspection)

4.5 Tool Interface
  - Registry of tools (name, I/O schema, permissions)
  - Invocation wrapper with:
    • Timeout, retry, sandbox boundaries
    • Input validation and output parsing
  - Tools include generic services (HTTP, DB), specialized MCP clients, and **node‑provided tools**

4.5.1 Node Tool Providers
  - Every node type can act as a tool provider, exporting one or more domain‑specific operations
  - Example: an Airtable node publishes tools such as `createRecord`, `getRecord`, `listRecords`, `updateRecord`, `deleteRecord`
    • Each tool has its own name, I/O schema, permissions, and retry/timeouts
  - Node tool registration occurs at startup or dynamically when a node is loaded
  - Planner treats node‑provided tools the same as external tools in dispatch and invocation

4.6 Output Generator
  - Formats final responses (text, JSON)
  - Annotates provenance: which memory sources, tools, and MCP servers were used

## 5. Data Flow & Control Loop

1. Receive input → preprocess
2. Retrieve context slices:
   - STM: recent turns
   - LTM: persistent facts via semantic search
   - MCP: context spans via internal/external servers
3. Planner synthesizes context → decides: LLM call or tool invocation
4. If tool needed:
   - Dispatch tool
   - Await result
   - Merge into context
   - (Optionally loop back)
5. Produce output
6. Update STM (append turn)
7. Conditional consolidation: flush STM snapshots to LTM or MCP

## 6. Memory Management Strategies

- STM eviction: LRU or floating token window
- LTM consolidation: scheduled or event‑driven snapshots
- MCP indexing: manage context segment lifecycles (pin, archive)
- Versioned memory entries to prevent stale overwrites

## 7. Tool & MCP Server Management

- Declarative manifest (YAML/JSON) for tools, MCP endpoints, and node‑provided tools
- Dynamic registration/unregistration at runtime (includes node tool providers)
- Per‑tool, per‑node, and per‑MCP ACLs and rate‑limits
- Logging of every invocation (inputs, outputs, latency), including node tool calls
- Node tool providers are auto‑registered in the manifest, namespaced by node type (e.g., `airtable.createRecord`)

## 8. Security & Sandboxing

- RBAC on tools and MCP servers
- Input/output sanitization before memory writes
- Network isolation: sandbox containers for untrusted tools
- Encryption at rest (LTM, STM snapshots)

## 9. API & Interfaces (internal)

• `POST /internal/agent-node/input` – send object to a node run
• `GET  /internal/agent-node/state` – inspect STM/LTM/MCP stats
• `POST /internal/agent-node/tool` – register/query tool or MCP client
• gRPC streams for real‑time thought trace and tool logs

## 10. Logging & Monitoring

- Structured logs for each decision, memory access, tool/MCP call
- Metrics: loop latency, memory hit‑rates, tool/MCP error rates
- Distributed tracing with request IDs across tools and MCP servers

## 11. Example Scenario

User: “Summarize yesterday’s customer support tickets about bug X.”
 1. Preprocessor normalizes text
 2. STM returns recent conversation
 3. LTM semantic search fetches ticket summaries
 4. MCP client queries index server for embeddings of bug‑X tickets
 5. Planner merges all contexts → invokes LLM
 6. Output formatted, returned
 7. STM updated; LTM/MCP optionally store the summary for future retrieval

## 12. Future Extensions

- Multi‑Agent orchestration bus
- Hierarchical memory (episodic, semantic layers)
- Automated tool/MCP discovery via schema introspection
- Reinforcement learning for memory prioritization