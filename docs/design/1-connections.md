# Connections & Integrations – Design Document (v1)

_Extends **0‑agents.md**. We generalise the earlier “provider connections” idea so that a **Connection** can represent **any external system** the platform talks to: LLM providers, data sources (Airtable, Postgres), communication channels (Gmail, Slack), etc._

## 1. Goals

1. Single, unified abstraction (**Connection**) regardless of backend type.  
2. Support **Bring‑Your‑Own‑Credentials** and platform‑hosted defaults.  
3. Manage connections via both **UI** and **Public API**.  
4. Make capabilities discoverable so the agent builder can restrict selectable skills (e.g. only show “Send Mail” if a Gmail connection exists).  
5. Store secrets securely and track per‑connection usage/costs where relevant.

## 2. Core Concepts

| Term              | Description |
|-------------------|-------------|
| **Integration**   | Static descriptor maintained by the platform that defines how to talk to a service (endpoints, auth scheme, supported capabilities). Eg. “openai”, “gmail”, “airtable”. |
| **Connection**    | User‑scoped instance of an integration containing secrets (API key, OAuth tokens, etc.) and configuration (workspace id, base url). |
| **Hosted Pool**   | Platform‑owned connection available as a default for certain integrations (e.g. OpenAI). |
| **Capability**    | Fine‑grained feature exposed by an integration (chat, embeddings, read_rows, send_email). |


## 3. Data‑model (Postgres)

We refactor the earlier specialised tables into a generic set.

1. `integrations` (replaces prior `providers` table, extends one from 0‑agents)
```
id 🔑 UUID
name TEXT UNIQUE NOT NULL            -- "openai", "gmail", "airtable"
category TEXT CHECK (category IN ('llm_provider','database','communication','storage','custom'))
auth_type TEXT CHECK (auth_type IN ('api_key','oauth2','none'))
base_url TEXT
logo_url TEXT
capabilities TEXT[]                  -- e.g. {'chat','embeddings'} for OpenAI, {'send_email'} for Gmail
created_at TIMESTAMPTZ DEFAULT now()
```

2. `connections`
```
id 🔑 UUID
user_id 🔗 UUID REFERENCES users(id) ON DELETE CASCADE
integration_id 🔗 UUID REFERENCES integrations(id)
name TEXT NOT NULL                  -- "Personal OpenAI", "Work Gmail"
secret JSONB ENCRYPTED              -- opaque structure interpreted by integration driver
config JSONB                        -- non‑secret fields (region, workspace id, base, etc.)
usage_limit_month INTEGER NULL      -- optional user cap (tokens, rows, emails etc.)
is_default BOOLEAN DEFAULT FALSE    -- one default per integration per user
created_at TIMESTAMPTZ DEFAULT now()
last_validated TIMESTAMPTZ NULL
status TEXT CHECK (status IN ('valid','invalid','expired')) DEFAULT 'valid'
```

3. `connection_usage` (superset of provider_usage)
```
id 🔑 UUID
connection_id 🔗 UUID REFERENCES connections(id)
run_id 🔗 UUID
metric JSONB                        -- e.g. {"prompt_tokens":123,"rows_read":456}
cost_cents INTEGER NULL             -- optional billing field
timestamp TIMESTAMPTZ DEFAULT now()
```

Backward‑compat advantage: existing `credentials` table from v0 can be migrated into `connections` 1‑to‑1.


## 4. Capability Discovery

When the agent builder UI loads:
1. `GET /api/integrations` – returns master list of integrations & capabilities.  
2. `GET /api/connections` – returns user’s concrete connections.  
3. UI combines the two: if user lacks a connection providing capability _X_, the corresponding skill is disabled or the builder shows a prompt “Add Gmail connection to enable”.

Programmatic agents can perform the same check via API.


## 5. UI Flows

### 5.1 Connections Dashboard

• Table groups connections by integration category (LLM, Communication, Data).  
• Badge for default connection.  
• Usage progress bar if `usage_limit_month` set.  
• “Add Connection” button opens wizard.

### 5.2 Add Connection Wizard

1. **Choose Integration** (searchable).  
2. **Authentication** step adapts based on `auth_type`:  
   • _API Key_ – paste key, optional org id.  
   • _OAuth2_ – redirect to provider, handle PKCE.  
3. **Configure** – provider specific (e.g. Airtable base id).  
4. **Validate** – lightweight test call.  
5. **Finish** – optionally mark as default.

### 5.3 Hosted Pool Opt‑in & Monetisation Strategy

Some integrations (primarily LLM providers) offer a **platform‑hosted pool** key that end‑users can “rent” from us instead of bringing their own.  Beyond convenience, this is a revenue lever: we purchase usage at wholesale rates and resell with a transparent markup.

#### 5.3.1 Schema Additions

Extend `integrations` with:

```
hosted_available BOOLEAN DEFAULT FALSE,
hosted_pricing JSONB  -- {"unit":"1k_tokens","base_cost_usd":0.002,"markup_pct":35}
```

Optionally, a separate `integration_pricing_tiers` table for volume discounts:

```
id 🔑 UUID
integration_id 🔗 UUID
threshold INTEGER          -- in billable units (e.g. tokens)
price_per_unit_usd NUMERIC -- after markup; lower at higher tiers
```

#### 5.3.2 UX Flow

1. **Default Choice Dialog** – When user adds their first agent that requires capability *chat* and has **no** personal connection, modal appears:
   • “Bring my own key” (recommended for heavy usage – link to doc).  
   • “Use hosted key at $0.30 / 1k tokens” (example).  The UI fetches `hosted_pricing` to display current rate and includes a cost‑calculator slider.

2. **Toggle per Agent** – In builder’s right‑side panel, under “Execution settings”, a **Connection selector** defaulting to:
   • user’s default connection (if exists) or  
   • **“Platform Hosted (estimate $X / run)**”.  Tooltip shows: “Running GPT‑4 @ $0.09 / 1k prompt tokens + $0.12 / 1k completion tokens”.

3. **Billing Transparency** – The account billing page shows two buckets:
   • **Pass‑through usage** – runs executed on user‑supplied keys (cost = $0).  
   • **Hosted pool usage** – runs on platform key billed at markup price; breakdown by model.

4. **Upsell Nudges** – If user exceeds $50/month on hosted pool, banner proposes “Add your own key and save ~X %”.

#### 5.3.3 Pricing / Arbitrage Mechanics

1. **Wholesale Contracts** – We negotiate discounted token rates with providers (e.g. 20 % below list price).  
2. **Markup** – Default 35 % margin above *list* price, leaving headroom even if provider price drops.  Markup configurable per integration in `hosted_pricing`.
3. **Dynamic Repricing** – Nightly cron syncs provider price changes and recomputes hosted margin keeping absolute cents / 1k tokens target.  UI pulls latest on page load.
4. **Volume Tiering** – Large customers can be routed to custom tier rows in `integration_pricing_tiers`—exposed via UI after contacting sales.

#### 5.3.4 Runtime Resolution (updated)

When runtime falls back to hosted pool (#7 algorithm step 2b):
1. The execution context is annotated `billing_mode = "hosted"`.
2. After run completes, usage is stored in `connection_usage` with `connection_id = NULL` **and** `hosted_integration_id` populated for attribution.
3. Billing engine multiplies usage × hosted price at time of usage.

#### 5.3.5 Safeguards

• **Daily Spend Caps** – Platform key enforces per‑user soft cap (e.g. $10/day) to contain abuse.  UI shows progress bar; user can request lift.  
• **Quality of Service** – Separate rate‑limit bucket so hosted pool users don’t starve paying BYOK customers.  
• **Fallback** – If platform hits wholesale quota, temporarily disable hosted option with status banner.


## 6. Public API

Base path remains `/api`.

| Method | Path                              | Description |
|--------|-----------------------------------|-------------|
| GET    | `/integrations`                   | List all integrations & metadata. |
| GET    | `/connections`                    | List user connections. |
| POST   | `/connections`                    | Create new connection. Body ⇒ `{integration_id, name, secret, config}` |
| GET    | `/connections/{id}`               | Retrieve single connection. |
| PATCH  | `/connections/{id}`               | Update fields (rename, default, usage cap…). |
| DELETE | `/connections/{id}`               | Soft delete. |
| POST   | `/connections/{id}/test`          | Validation probe (returns latency + provider‑specific info). |

The API surface is integration‑agnostic: secret & config are opaque JSON to the client, validated server‑side by integration drivers.


## 7. Runtime Usage Resolution

1. **Skill Execution** requires capability _C_.  
2. Agent’s skill spec may explicitly reference a `connection_id`. If missing, runtime selects:  
   a. User’s `connections` with that capability and `is_default=true`.  
   b. If none, integration’s hosted pool (if available).  
3. Usage/cost is recorded to `connection_usage` with normalised metric keys (provider driver maps raw response ➜ standard fields: `prompt_tokens`, `rows_read`, `emails_sent`, …).


## 8. Security & Compliance

• Secrets encrypted with `pgcrypto`; application decrypts with KMS key.  
• OAuth refresh tokens stored, access tokens cached in Redis with TTL.  
• Row‑Level Security isolates `connections` per tenant.  
• Audit log for all create/update/delete, including validation results.  
• GDPR export bundles `connections` minus secrets (masked).


## 9. Non‑functional Targets

| Metric                  | Target |
|-------------------------|--------|
| Connection validation P95 | ≤ 800 ms |
| Max secret size         | 10 KB |
| Hosted pool availability | 99.95 % |
| Capability sync lag     | ≤ 60 sec |


## 10. Open Questions

1. Should we allow **org‑level connections** vs user‑level only? Requires additional ACL column.  
2. How to redact / masking strategy in payloads returned to API (show last 4 chars?).  
3. Rate‑limiting per‑integration (e.g. Gmail daily send quota) – central throttling layer?  
4. Support for **events/webhooks** originating from integrations (e.g. “new Airtable row” trigger). Might require `connection_webhooks` table to persist callback URL + secret.


---
_Last updated: 2025‑04‑17_
