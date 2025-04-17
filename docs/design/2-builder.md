# Agent Builder & Extensibility – Design Document

This document augments **0‑agents.md** and **1‑connections.md**, focusing on **how users create, debug, and extend agents** via a modern, node‑based interface similar to Make.com, n8n, Node‑RED, or Kestrel.

## 1. Goals

1. **Low‑floor, high‑ceiling** – non‑programmers can click‐together flows; power users can drop down to code.
2. **Multi‑platform** – builder works on desktop web, tablet, and (lightweight) mobile; same JSON spec behind the UI to allow programmatic authoring via API/IaC.
3. **Extensible** – custom skills, scripts, inbound/outbound webhooks, and marketplace asset sharing.
4. **Observable & debuggable** – real‑time logs, step‑wise replay, version control.


## 2. Builder UX Overview

### 2.1 Canvas

• Node‑based DAG canvas with pan/zoom and mini‑map.  
• Nodes represent **Skills** (LLM call, HTTP request, Transform, Code, Wait, etc.).  
• Edges represent data flow; conditional edges (success/failure) supported.

### 2.2 Inspector Panel

• Shows selected node properties (model, prompt, retry policy).  
• Supports **expression editor** with TypeScript‑like syntax (`{{input.summary}}`).  
• Live validation against JSONSchema.

### 2.3 Trigger Bar

• Configure **Triggers**: schedule, webhook, cron, external events (Slack slash‑command).  
• Generates unique webhook URLs & secrets stored in `triggers` table.

### 2.4 Debug Sidebar

• “Test Run” executes agent with sample inputs, streaming logs & token usage.  
• Step selector to inspect inputs/outputs per node.  
• Ability to pin example runs as regression tests.


## 3. Internal Representation

Builder serialises to the **`agent_versions.graph` JSONB** already defined. The draft schema:

```jsonc
{
  "nodes": [
    {
      "id": "node-1",
      "type": "llm",
      "skill_id": "skill-llm-chat",
      "params": {
        "model": "gpt-4o",
        "prompt": "Write a reply to: {{input.email_body}}"
      }
    }
  ],
  "edges": [
    { "from": "node-1", "to": "node-2", "condition": "success" }
  ]
}
```

It intentionally mirrors the builder UI so that **agents can be created in three ways**:
1. Drag‑and‑drop builder.  
2. REST API: `POST /api/agents` with JSON payload.  
3. IaC (Terraform provider or YAML) which submits the same JSON.


## 4. Extensibility Mechanisms

### 4.1 Custom Code Node (“Script”)

• Run JavaScript (Deno runtime) or Python (Pyodide) in a **server‑side sandbox** – no egress by default except allowed `fetch`.  
• Script code stored in `skills.spec.code` and snapshotted per version.  
• Time‑limit (e.g. 5 s) & memory cap (128 MB).  
• Dependency whitelist (lodash, dayjs, etc.) – downloaded at runtime.

### 4.2 Plugin Marketplace

• Users can package skills (code, icon, config schema) and publish to marketplace.  
• Installed plugins appear in builder palette and create new `skills` rows with `owner_user_id = NULL` but `is_third_party = TRUE`.

### 4.3 Webhooks

Inbound:  
• “Webhook Trigger” node creates unique URL; events POSTed are injected as `input` for agent run.  

Outbound:  
• “HTTP Request” node or **Outgoing Webhook** node uses `connections` with `category='communication'` or generic HTTP model.  
• Response mapping UI to transform and save outputs.

Security: CSRF secret in query param or header; HMAC option for verification.


## 5. Multi‑platform Rendering

We reuse a single **React** code‑base:

• Desktop web: full canvas, keyboard shortcuts.  
• Tablet: gestures; toolbar collapses to icons.  
• Mobile: read‑only + basic edits (rename node, trigger run).  
Underlying state managed with **Zustand** or **Redux Toolkit**; view transforms via **react‑flow** library.


## 6. API Extensions

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/agents/{id}/runs/test` | Executes test run, returns streamed logs. |
| POST | `/skills` | Create/update custom skill (used by marketplace & code node). |
| GET  | `/webhooks/{id}` | Retrieve metadata & sample payload. |


## 7. Versioning & Source Control

• Every “Save” in builder creates a new row in `agent_versions`.  
• Users can view diff between versions using **json‑patch** visualisation.  
• Optional GitHub sync: push/pull builder JSON in repo under `/agents/agent‑name.json` enabling PR workflows.


## 8. Security Concerns

• Sandbox code execution using **gVisor** containers.  
• Rate limit inbound webhooks per agent.  
• Secrets injected at runtime, never stored directly in node params.


## 9. Open Questions

1. Is visual branching sufficient or do we need fully‑fledged **state machines**?  
2. Should we support **long‑running steps** (hours / days)? Might require persistence of node state and resumption.  
3. Permissioning: can organisation admins restrict use of Script node?  
4. Marketplace revenue share model.


---
_Last updated: 2025‑04‑17_
