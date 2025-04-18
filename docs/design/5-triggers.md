 # Triggers & Eventing – Design Document

 This document augments **0‑agents.md**, **1‑connections.md**, and **2‑builder.md**, focusing on how external and internal events automatically trigger agent runs (see **4‑execution‑runtime.md** for execution pipeline).

 > _Last updated: 2025‑04‑18_

 ## 1. Goals

 1. Support **automated runs** when external or internal events occur.
 2. Provide a **flexible, extensible** framework for multiple trigger types: webhooks, polling, websockets, SSE, cron, pub/sub, file watchers, etc.
 3. Enable **user-defined conditions** to filter and route events to workflows.
 4. Ensure **secure**, **scalable**, and **observable** triggering.
 5. Simplify **management** and **monitoring** via UI and public API.

 ## 2. Use Cases

 - **Webhook**: GitHub push, Stripe payment, PagerDuty alert.
 - **Polling**: Poll REST API or database table for changes.
 - **Pub/Sub**: Consume messages from Kafka, Redis Streams, Cloud Pub/Sub.
 - **WebSocket / SSE**: Long-lived connections to real-time streams.
 - **Cron / Schedule**: Time-based triggers with cron expressions.
 - **File / Storage**: Watch filesystem (local, S3, GCS) for new or modified files.
 - **Custom**: User-implemented hooks into arbitrary systems.

 ## 3. Core Concepts

 | Term                 | Description |
 |----------------------|-------------|
 | **Trigger**          | Definition of an event source + parameters that can start a run. |
 | **Trigger Provider** | Plugin or adapter implementing a trigger type (webhook, cron, etc.). |
| **Trigger Instance** | User-configured trigger with parameters (URL path, schedule, provider-specific filters). |
| **Event**            | Payload emitted by a provider, delivered to the workflow runner. |
| **Filter Node**      | Graph node that evaluates an expression to filter out unwanted events/items. |
| **Runner**           | Component that starts the workflow run when a trigger fires. |

 ## 4. Architecture Overview

```text
 +------------------+     +--------------+     +----------+
 | Trigger Providers| --> | Event Broker | --> | Runner   |
 | (Webhook, Cron,  |     | (Queue / Bus)|     | (Engine) |
 |  Poller, Pub/Sub)|     +--------------+     +----------+
 +------------------+                              |
                                                  v
                                    Workflow Graph (with optional Filter nodes)
``` 

- Trigger Providers listen or poll for events and push raw events into a durable Event Broker.
- The Runner consumes raw events and starts workflows with the event as initial payload.
- For fine-grained filtering beyond provider-level parameters, users add a dedicated Filter (If) node downstream in their workflow graph to drop unwanted events/items.

 ## 5. Trigger Providers

 ### 5.1 Webhook Provider
 - Expose HTTP endpoints (`POST /webhooks/:provider/:id`) secured with HMAC or API keys.
 - Validate signature, parse payload, enqueue event.

 ### 5.2 Polling Provider
 - Configurable polling interval, backoff, rate limits.
 - Query REST APIs, databases, or other endpoints.
 - Track last-seen watermark or cursor in storage.

 ### 5.3 Pub/Sub Provider
 - Support protocols: Kafka, Redis Streams, Google Pub/Sub, AWS SNS/SQS.
 - Use client libraries for subscription and offset management.

 ### 5.4 WebSocket / SSE Provider
 - Maintain long-lived connection(s) with auto-reconnect and heartbeats.
 - Stream events into broker.

 ### 5.5 Cron / Scheduler Provider
 - Use cron expressions or interval definitions.
 - Leverage scheduler library (e.g., [cron](https://pkg.go.dev/github.com/robfig/cron) in Go).
 - Emit timer events on schedule.

 ### 5.6 File Watcher Provider
 - Support inotify or polling for local FS, S3, GCS.
 - Detect create/update/delete, emit file metadata or content.

 ## 6. Configuration & Management

 - **API**: `/api/triggers` for CRUD operations on Trigger Instances.
 - **UI**: Builder page for visual trigger setup; form fields vary by provider type.
 - **Schema**: TriggerInstance record:

 ```sql
 CREATE TABLE triggers (
   id UUID PRIMARY KEY,
   user_id UUID REFERENCES users(id),
   provider TEXT NOT NULL,      -- e.g. 'webhook', 'cron'
   config JSONB NOT NULL,       -- provider-specific parameters
   enabled BOOLEAN DEFAULT TRUE,
   last_checked TIMESTAMPTZ,
   created_at TIMESTAMPTZ DEFAULT now(),
   updated_at TIMESTAMPTZ DEFAULT now()
 );
 ```

## 7. Downstream Filtering (Filter/If Node)

- Complex or dynamic filtering is performed in the workflow graph via a dedicated Filter or If node (e.g., `IfNode`).
- Users configure the node immediately after a trigger to evaluate expressions or provider-specific filters in a friendly UI or DSL.
- This keeps the trigger layer focused on event ingestion and basic provider parameters, while the workflow graph handles all conditional logic.

 ## 8. Security & Governance

 - **Access Control**: Owners only manage their triggers; RBAC for shared workflows.
 - **Secrets**: HMAC secrets, API keys stored encrypted.
 - **Rate Limiting**: Protect HTTP endpoints and consumers.
 - **Validation**: Schema and payload validation per provider.

 ## 9. Observability & Metrics

 - **Logs**: Structured logs for event receipt, evaluation, and run invocation.
 - **Metrics**: Counters and histograms:
   - `triggers_events_received_total{provider}`
   - `triggers_events_filtered_total{provider,result}`
   - `triggers_runs_started_total{provider}`
   - `triggers_latency_seconds{provider}` (end-to-end)
 - **Tracing**: Link event ingestion spans to workflow run spans.

 ## 10. Scalability & Reliability

 - **Horizontal Scaling**: Multiple provider instances and evaluators; event broker partitions.
 - **At-Least-Once Delivery**: Deduplication on Event ID or fingerprint to avoid duplicate runs.
 - **Backpressure & Retries**: Exponential backoff on failures; poison queue for bad messages.

 ## 11. Extensibility

 - Define a TriggerProvider interface:

 ```go
 type TriggerProvider interface {
   Init(ctx context.Context, config map[string]interface{}) error
   Run(ctx context.Context, out chan<- Event)    // listen or poll, send events
   Shutdown(ctx context.Context) error
 }
 ```

 - Register new providers via plugin or code registration.

## Graph ↔ Trigger Persistence ↔ Scheduler Lifecycle

This section describes how trigger nodes defined in an agent graph flow through persistence into the active scheduler and back to workflow execution.

1. Graph Authoring
   - In the visual builder, users drag a Trigger node (cron, webhook, poller, etc.) onto the canvas and configure its parameters.
   - The node has an internal `nodeId`, and its properties include provider type and `config` JSON.

2. Persisting on Version Save
   - When a new agent version is saved via `POST /agents/{agentID}/versions`, the server stores the full graph in `agent_versions.graph`.
   - A post-save hook inspects the graph JSON, extracts all Trigger nodes (identified by type and `nodeId`).

3. Trigger Sync
   - For each Trigger node in the latest graph:  
     • If no matching `triggers` row exists (by `agent_id` + `node_id`), INSERT a new record.  
     • If the node’s `config` or `enabled` flag changed, UPDATE the corresponding row.  
   - Any existing `triggers` rows for this `agent_id` whose `node_id` is no longer present are DELETED (or marked `enabled=false`).

4. Scheduling Service
   - A long‑running Trigger Engine (separate process or Go routine) polls or subscribes to the `triggers` table for active rows.
   - For each active trigger it:  
     • Initializes the correct TriggerProvider (cron scheduler, HTTP webhook router, poll loop, pub/sub consumer, etc.) with the stored `config`.  
     • Registers callback handlers that will fire when an event is received.

5. Event Dispatch & Run Creation
   - When a provider detects an event, it invokes the Runner API (`POST /agents/{agentID}/runs`) with:  
     • `versionId` (latest version),  
     • `startNodeId` (the trigger’s nodeId),  
     • `input` payload (event data).
   - The run service creates a new Run record, and the execution engine begins at that trigger node, passing the event envelope into the graph.

6. Cleanup & Updates
   - When the user updates the graph (adds/removes trigger nodes) and publishes a new version, the sync logic enables/disables or recreates trigger rows automatically.  
   - The Trigger Engine watches for changes (via notifications or periodic reload) and will start/stop provider instances accordingly.

## 12. Open Questions & Next Steps

 1. DSL choice for conditions: JSONPath vs CEL vs Lua vs JavaScript.
 2. Multi-tenancy: shared vs per-tenant event topics.
 3. Transactionality: start run only after successful commit in polling scenarios.
 4. UI / UX patterns for visual condition builder.
 5. Support multi-step triggers (e.g. wait for multiple events).
 6. Versioning TriggerInstance schema for backward compatibility.