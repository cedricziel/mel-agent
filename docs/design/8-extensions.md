# Extensions & Composability – Design Document

This document augments **0‑agents.md**, **1‑connections.md**, **2‑builder.md**, **4‑execution‑runtime.md**, **5‑triggers.md**, and **7‑agent‑node.md**, defining a unified plugin architecture to make the system open, composable, and community‑driven.

## 1. Goals

1. **Open Ecosystem** – support two streams of extensibility: GoPlugin modules (compiled Go binaries) and external Model Context Protocol (MCP) servers.
2. **Plugin Types** – unify implementations into two categories: GoPlugin modules for in-process extensions, and MCPServer integrations for remote model & tool resources.
3. **Dynamic Discovery** – load GoPlugin binaries on startup (or via API) and discover external MCP servers through user-supplied configurations.
4. **Consistent Interfaces** – define Go interfaces for GoPlugin modules and leverage the standard MCP protocol for external servers.
5. **UI Extensibility** – builder UI picks up GoPlugin metadata and dynamically incorporates nodes for connected MCP servers (models, tools, resources).
6. **Security & Isolation** – sandbox or isolate GoPlugin code, enforce plugin version compatibility, and secure communications with MCP servers.

## 2. Motivation

As usage scenarios grow (custom data transforms, niche integrations, domain‑specific triggers), core will bloat. A plugin model enables:
- Lean core orchestration and persistence.
- Ecosystem contributions without core churn.
- Language‑agnostic or multi‑language plugin support (Go, WASM, external processes).
- Treat external tools and resources (e.g., APIs, cloud services) as first‑class nodes via NodePlugin wrappers.

## 3. High‑Level Architecture

```text
+------------+     +----------------+     +-------------+
|   Core     |<--->| Plugin Registry|<--->| Plugin Host |
| (API, DB,  |     | (catalog of    |     | (loader,     |
|  Scheduler)|     |  metadata)     |     |  isolation)  |
+------------+     +----------------+     +-------------+
       |                                       ^   ^
       |                                       |   |
       v                                       |   |
  +----------+   instantiates   +------------+ |   |
  | Extension|----------------->| Node/Trigger|-----+
  |  Points  |                  | Connection  | plugin
  +----------+                  +------------+
```

## 4. Extension Categories

1. **NodePlugin**: encapsulates a unit of work (LLM call, HTTP, transform, script).
2. **TriggerPlugin**: defines inbound events, polling or webhook scheduling.
3. **ConnectionPlugin**: manages external integrations, credential flows, capability discovery.
4. **AgentProtocolPlugin**: overrides or extends the agent control loop (memory, tool selection).
5. **ModelServerPlugin**: integrates external Model Context Protocol (MCP) servers, registering available models, tools, and resources as dynamic NodePlugins.
   - Adheres to the official MCP specification (https://modelcontextprotocol.io), for example:
     • GET /models → list available models
     • GET /models/{modelID} → model metadata and parameter schema
     • POST /models/{modelID}/generate → inference call
     • GET /tools → list server‑exposed tools/resources
     • POST /tools/{toolID}/invoke → tool invocation

Plugins may implement any combination of these categories, exporting multiple extension types from a single module.

## 5. Plugin Interface & SDK

Define Go interfaces in `internal/plugin` (or `pkg/plugin`), and allow a single plugin module to export multiple extension types:
```go
type PluginMeta struct {
  ID          string            // unique plugin identifier
  Version     semver.Version    // semantic version
  Categories  []string          // extension types provided, e.g. ["node","trigger"]
  Params      []ParamSpec       // parameter schema
  UIComponent string            // optional path to React bundle
}

type Plugin interface {
  Meta() PluginMeta
  // category‑specific methods follow (e.g. Execute for NodePlugin)
}

type NodePlugin interface {
  Plugin
  Execute(ctx context.Context, inputs map[string]any) (map[string]any, error)
}
// similar TriggerPlugin, ConnectionPlugin, AgentProtocolPlugin

```go
// ModelServerPlugin exposes external MCP servers and models for inference
type ModelServerPlugin interface {
  Plugin
  // Models returns metadata for each model supported by this server
  Models() []ModelSpec
  // InvokeModel sends input to the specified model and returns the response
  InvokeModel(ctx context.Context, modelID string, input any) (any, error)
}

// ModelSpec describes an available model on a ModelServerPlugin
type ModelSpec struct {
  ID          string       // unique model identifier
  Description string       // human‑readable summary
  Params      []ParamSpec  // model invocation parameters (e.g. temperature)
}
// Internally, the host maps Models() and InvokeModel() to the standard MCP HTTP API
// (e.g. GET /models, GET /models/{id}, POST /models/{id}/generate).
```

Use `mel-server` as the host binary that loads plugins, and `mel` as the CLI binary for plugin management.
Provide a CLI scaffold (`mel plugin init`) via the `mel` binary to create new modules, generate boilerplate code, tests, and a manifest.

## 6. Registry & Discovery

- **Static Registration**: plugin modules call `plugin.Register(...)` for each extension implementation; a single module can register NodePlugins, TriggerPlugins, ConnectionPlugins, etc.
- **Manifest Files**: optional JSON/YAML under `plugins/` for metadata, version pins.
- **Database Catalog**: persist installed plugin records, lock versions, audit installs.

## 7. UI Integration

- Expose `/api/extensions` endpoint returning catalog of PluginMeta.
- Builder fetches metadata, dynamically imports or lazy‑loads React bundles per plugin.
- Fallback to generic property editor using ParamSpec when no custom component is provided.

## 8. Versioning & Compatibility

- Follow Semantic Versioning for plugin releases.
- Define major/minor compatibility contracts in Go interfaces.
- Implement compatibility checks on registration; reject mismatched plugin versions.

## 9. Migration Roadmap

1. Abstract the built‑in LLM node into a ModelServerPlugin; migrate core LLM integration into `plugins/model/`.
2. Extract existing nodes (HTTP, If, Trigger) into `plugins/node/` modules.
3. Wire up registry in `internal/runs` and `internal/api/nodes` to use dynamic lookup.
4. Move Trigger definitions into `plugins/trigger/`; update scheduler to load plugins.
5. Implement `/api/extensions` and update front‑end builder to consume new catalog.
6. Publish first Plugin SDK and CLI; onboard community to build sample plugins.
7. Extend to Connection and AgentProtocol plugins.

## 10. Open Questions & Next Steps

- Sandbox model: Go plugin vs. WASM vs. external process?
- Security policy for untrusted plugins: code signing, review process.
- Marketplace or registry service: hosted vs. self‑serve.
- Plugin dependency management and governance.