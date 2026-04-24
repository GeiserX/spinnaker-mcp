# Roadmap

This document outlines the planned features and improvements for Spinnaker-MCP. The Spinnaker Gate API is extensive (~150+ endpoints across 60+ controllers), and this MCP server currently covers the highest-value operations. The roadmap is organized by priority and grouped by API domain.

## Current State (v0.3.2)

**37 tools** covering the core deployment workflow, pipeline lifecycle, and infrastructure visibility:

- Applications: list, get
- Pipelines: list, get config, trigger, save, update, delete, history
- Executions: list, get, search, cancel, pause, resume, restart stage, evaluate expression
- Strategies: list, save, delete
- Infrastructure: server groups, load balancers, clusters, scaling activities, target server groups, firewalls, instances, console output, images, image tags, networks, subnets, accounts
- Tasks: get status

**Authentication**: Bearer token, basic auth, x509 client certificates

**Transports**: stdio, Streamable HTTP

---

## ~~Phase 1 — Pipeline Lifecycle (v0.3.0)~~ ✅

Complete coverage of pipeline CRUD and execution management.

### Pipeline Configuration Management

- [x] `save_pipeline` — Save/create a pipeline definition (`POST /pipelines`)
- [x] `update_pipeline` — Update an existing pipeline definition (`PUT /pipelines/{id}`)
- [x] `delete_pipeline` — Delete a pipeline definition (`DELETE /pipelines/{application}/{pipelineName}`)
- [x] `get_pipeline_history` — Get revision history for a pipeline config (`GET /pipelineConfigs/{id}/history`)

### Execution Control

- [x] `restart_stage` — Restart a failed stage within an execution (`PUT /pipelines/{id}/stages/{stageId}/restart`)
- [x] `search_executions` — Rich search across executions by trigger type, time range, status, event ID (`GET /applications/{app}/executions/search`)
- [x] `evaluate_expression` — Evaluate a SpEL pipeline expression against an execution context (`POST /pipelines/{id}/evaluateExpression`)

### Strategy Management

- [x] `list_strategies` — List deployment strategy configurations for an application
- [x] `save_strategy` — Create or update a deployment strategy
- [x] `delete_strategy` — Delete a deployment strategy

---

## ~~Phase 2 — Infrastructure Deep Dive (v0.3.0)~~ ✅

Read-only visibility into Spinnaker-managed infrastructure.

### Clusters

- [x] `list_clusters` — List cluster names grouped by account (`GET /applications/{app}/clusters`)
- [x] `get_cluster` — Get cluster details including server groups (`GET /applications/{app}/clusters/{account}/{cluster}`)
- [x] `get_scaling_activities` — Get scaling activities for a cluster (`GET .../scalingActivities`)
- [x] `get_target_server_group` — Target-based server group lookup (newest, oldest, ancestor)

### Security Groups / Firewalls

- [x] `list_firewalls` — List all firewalls/security groups across accounts
- [x] `get_firewall` — Get firewall details by account, region, and name

### Instances

- [x] `get_instance` — Get instance details (health, metadata, launch time)
- [x] `get_console_output` — Get instance console output for debugging

### Images

- [x] `find_images` — Search for machine images by tags, region, account
- [x] `get_image_tags` — List image tags for a repository

### Networks and Subnets

- [x] `list_networks` — List VPCs/networks by cloud provider
- [x] `list_subnets` — List subnets by cloud provider

### Credentials / Accounts

- [x] `list_accounts` — List all configured cloud accounts/credentials
- [x] `get_account` — Get account details and permissions

---

## Phase 3 — Protocol & Developer Experience (v0.3.3)

Full MCP protocol surface (resources, prompts) and production-readiness improvements for Kubernetes deployment. This release transforms spinnaker-mcp from a tool-only server into a complete MCP v1 implementation with first-class Kubernetes support.

### MCP Resources

Expose read-only Spinnaker data as browsable MCP resources, allowing clients to discover and read infrastructure state without calling tools. Resources use URI templates following the `spinnaker://` scheme and return JSON with `application/json` MIME type.

**Application Resources:**

- [ ] `spinnaker://applications` — All Spinnaker applications with metadata (email, cloud providers, create time)
- [ ] `spinnaker://application/{name}` — Application details including accounts, clusters, attributes, and pipeline count
- [ ] `spinnaker://application/{name}/pipelines` — All pipeline configurations for an application (name, ID, stages, triggers, parameters)
- [ ] `spinnaker://application/{name}/executions` — Recent pipeline executions (last 25) with status, timing, and trigger info

**Pipeline & Execution Resources:**

- [ ] `spinnaker://pipeline/{id}` — Single pipeline configuration by pipeline config ID
- [ ] `spinnaker://execution/{id}` — Full execution details including all stages, outputs, and timing

**Infrastructure Resources:**

- [ ] `spinnaker://accounts` — All configured cloud accounts with provider type and environment
- [ ] `spinnaker://account/{name}` — Account details with regions, permissions, and provider-specific metadata
- [ ] `spinnaker://application/{name}/clusters` — Clusters grouped by account for an application
- [ ] `spinnaker://application/{name}/server-groups` — Server groups with instance counts, image, and capacity
- [ ] `spinnaker://application/{name}/load-balancers` — Load balancers across all accounts and regions

All resources are backed by existing Gate client methods — no new API calls required. Requires `server.WithResourceCapabilities(true)` and `s.AddResourceTemplate()` for parameterized URIs.

### MCP Prompts

Pre-built prompt templates for common Spinnaker workflows. Prompts give LLMs structured starting points for multi-step operations, referencing the right tools in the right order.

- [ ] **`deploy-review`** — Review a pipeline configuration before triggering. Arguments: `application` (required), `pipeline` (required). Produces a structured checklist covering: pipeline stages and their types, trigger configuration, parameter defaults, notification setup, manual judgment gates, expected artifacts, and last 5 execution outcomes. Designed for pre-deployment sign-off.

- [ ] **`incident-response`** — Investigate a failed or stuck deployment. Arguments: `application` (required), `execution_id` (optional — if omitted, targets the most recent failed execution). Guides the LLM through: execution status and failed stage details, stage error messages and context, server group health across regions, recent scaling activities, and instance-level console output for unhealthy instances. Structured as a diagnostic runbook.

- [ ] **`pipeline-audit`** — Audit a pipeline configuration for best practices. Arguments: `application` (required), `pipeline` (required). Checks for: missing notification stages, no manual judgment gates before production deploys, hardcoded image references (vs. artifact bindings), unused parameters, missing rollback strategies, stages without timeout configuration, and overly permissive trigger settings.

- [ ] **`infra-overview`** — Summarize the complete infrastructure state for an application. Arguments: `application` (required), `account` (optional — if omitted, covers all accounts). Aggregates: server groups by cluster with instance counts and health, load balancers and their target groups, firewall/security group rules, networks and subnets in use, and current scaling policies.

- [ ] **`rollback-plan`** — Generate a rollback strategy for a deployment. Arguments: `application` (required), `cluster` (required), `account` (required), `region` (required). Uses target server group lookups (previous, ancestor) and scaling activities to propose: which server group to roll back to, expected instance count changes, load balancer re-targeting steps, and verification checks after rollback.

Each prompt returns `[]mcp.PromptMessage` with structured text and tool-call suggestions. Requires `server.WithPromptCapabilities(true)` and `s.AddPrompt()`.

### Toolsets

Enable or disable tool groups via the `--toolsets` flag or `TOOLSETS` environment variable, reducing the tool surface for clients that only need specific functionality. Useful for security-conscious deployments (read-only mode) or reducing LLM token overhead when only a subset of operations is needed.

**Tool Groups:**

| Group | Tools | Count |
|-------|-------|-------|
| `applications` | `list_applications`, `get_application` | 2 |
| `pipelines` | `list_pipelines`, `get_pipeline`, `trigger_pipeline`, `save_pipeline`, `update_pipeline`, `delete_pipeline`, `get_pipeline_history` | 7 |
| `executions` | `list_executions`, `get_execution`, `search_executions`, `cancel_execution`, `pause_execution`, `resume_execution`, `restart_stage`, `evaluate_expression` | 8 |
| `strategies` | `list_strategies`, `save_strategy`, `delete_strategy` | 3 |
| `infrastructure` | `list_server_groups`, `list_load_balancers`, `list_clusters`, `get_cluster`, `get_scaling_activities`, `get_target_server_group`, `list_firewalls`, `get_firewall`, `get_instance`, `get_console_output`, `find_images`, `get_image_tags`, `list_networks`, `list_subnets`, `list_accounts`, `get_account` | 16 |
| `tasks` | `get_task` | 1 |

**Meta-Groups:**

| Meta | Resolves to | Use case |
|------|-------------|----------|
| `all` | All 6 groups (default) | Full access |
| `readonly` | All tools annotated `readOnly` (24 tools) | Safe observation mode — no mutations possible |
| `mutating` | Tools annotated `mutating` or `destructive` (10 tools) | Action-only (pair with `applications` for context) |

**Usage examples:**

```sh
# Only pipeline and execution tools (15 tools)
spinnaker-mcp --toolsets=pipelines,executions

# Read-only mode — safe for monitoring agents (24 tools)
TOOLSETS=readonly spinnaker-mcp

# Infrastructure visibility only (16 tools)
spinnaker-mcp --toolsets=infrastructure

# Everything except strategies (34 tools)
spinnaker-mcp --toolsets=applications,pipelines,executions,infrastructure,tasks
```

Comma-separated, case-insensitive. Invalid group names produce a startup error listing valid groups. The `--toolsets` flag takes precedence over the `TOOLSETS` env var (consistent with CLI-over-env convention). When toolsets are active, only matching tools are registered via `s.AddTool()`, and resources/prompts are always registered regardless of toolset selection.

### Health Check Endpoints

HTTP health and readiness probes for Kubernetes liveness/readiness, Docker `HEALTHCHECK`, and load balancer health checks.

- [ ] **`GET /healthz`** — **Liveness probe**. Returns `200 OK` if the server process is running. No external dependency checks — if the process can respond, it's alive.
  ```json
  {"status": "ok", "version": "0.3.3"}
  ```

- [ ] **`GET /readyz`** — **Readiness probe**. Returns `200 OK` if the Gate API is reachable (HEAD request to `GATE_URL` with 5-second timeout). Returns `503 Service Unavailable` if Gate is down or unreachable.
  ```json
  {"status": "ready", "gate_url": "http://spin-gate:8084", "gate_reachable": true}
  ```

HTTP transport only — stdio mode has no HTTP surface (the parent process manages lifecycle directly). Probes run on the same `MCP_PORT` (default 8085) using a custom HTTP handler that wraps `StreamableHTTPServer` and adds health routes. Probes are unauthenticated — no Gate credentials required to check health.

**Docker HEALTHCHECK example:**

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8085/healthz || exit 1
```

### Helm Chart

Kubernetes-native deployment as a standalone service or sidecar container, packaged as a Helm chart for simple, repeatable installation.

**Chart structure:**

```
helm/spinnaker-mcp/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── serviceaccount.yaml
│   └── ingress.yaml
└── README.md
```

**Deployment modes:**

1. **Standalone** (default) — Dedicated pod running spinnaker-mcp with HTTP transport, exposed via ClusterIP Service. Ideal for shared access by multiple agents or services within the cluster.
2. **Sidecar** — Container spec injected into an existing pod (e.g., alongside an LLM agent container), communicating via localhost HTTP or stdio. Enabled via `sidecar.enabled: true` in values.

**Key values:**

```yaml
gate:
  url: "http://spin-gate:8084"
  auth:
    type: token                       # token | basic | cert | none
    token: ""                         # or use existingSecret
    user: ""
    pass: ""
  tls:
    insecure: false
    certSecret: ""                    # Secret containing client cert + key

transport: http                       # http | stdio (stdio for sidecar mode)
port: 8085
bindAddr: "0.0.0.0"
toolsets: "all"

image:
  repository: drumsergio/spinnaker-mcp
  tag: ""                             # Defaults to chart appVersion
  pullPolicy: IfNotPresent

replicaCount: 1

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 128Mi

probes:
  liveness:
    path: /healthz
    initialDelaySeconds: 5
    periodSeconds: 10
  readiness:
    path: /readyz
    initialDelaySeconds: 5
    periodSeconds: 10

ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts: []
  tls: []

serviceAccount:
  create: true
  name: ""
  annotations: {}

sidecar:
  enabled: false
```

**Distribution:**

- OCI registry via CI: `helm push` to `oci://ghcr.io/geiserx/spinnaker-mcp`
- GitHub Pages: `gh-pages` branch with `helm repo index` for traditional `helm repo add`
- ArtifactHub listing for discoverability

---

## Phase 4 — Tasks and Manual Judgment

High-value additions for interactive deployment workflows. These would let an LLM approve/reject manual judgment stages and manage ad-hoc operations.

### Task Management

- [ ] `create_task` — Create an ad-hoc orchestration task (scale, destroy, rollback, etc.) (`POST /tasks`)
- [ ] `cancel_task` — Cancel a running task (`PUT /tasks/{id}/cancel`)

### Manual Judgment

- [ ] `approve_stage` — Approve a manual judgment stage (via `PATCH /pipelines/{id}/stages/{stageId}`)
- [ ] `reject_stage` — Reject a manual judgment stage

---

## Future — Community-Driven Expansion

The following areas cover the remaining ~110 Gate API endpoints. They are documented here as a reference for contributors. PRs are welcome — open an issue first for larger features.

### Canary Analysis (Kayenta)

- [ ] `list_canary_configs`, `get_canary_config`, `save_canary_config`, `delete_canary_config`
- [ ] `start_canary`, `get_canary_result`, `list_canary_results`

### Artifacts and Builds

- [ ] `list_artifact_accounts`, `list_artifact_versions`, `fetch_artifact`
- [ ] `list_build_masters`, `get_build`, `trigger_webhook`

### Managed Delivery (Keel)

- [ ] Delivery configs CRUD, resource management, environment controls (pin/veto/approve)

### Kubernetes Native

- [ ] `get_manifest`, `deploy_manifest`, `scale_manifest`, `rollback_manifest`

### Pipeline Templates (v2)

- [ ] Template CRUD, plan/preview, dependency listing

### Observability

- [ ] Projects, entity tags, notifications, global search, system info

---

## Cross-Cutting Improvements

### Authentication

- [x] Bearer token, basic auth, x509 client certificates
- [ ] OAuth2/OIDC token refresh — Automatic token refresh for OAuth-based auth
- [ ] Cookie-based auth — Support session cookies from browser SSO flows

### Transport

- [x] stdio and Streamable HTTP transports

### Developer Experience

- [x] `--version` / `--help` CLI flags
- [x] Tool annotations (readOnly, mutating, destructive hints)
- [ ] Structured logging — JSON log output with configurable levels (`LOG_LEVEL` env var) for production observability
- [ ] Prometheus metrics — `/metrics` endpoint exposing tool call counts, Gate API latency histograms, and error rates
- [ ] Response caching — Optional short-lived cache for read-only Gate responses (`CACHE_TTL` env var, default off). Applications and accounts change rarely; caching reduces Gate load for polling agents.

### Distribution

- [x] npm — `npx spinnaker-mcp` or `npm install -g spinnaker-mcp`
- [x] Docker Hub — `drumsergio/spinnaker-mcp`
- [ ] Homebrew tap — `brew install GeiserX/tap/spinnaker-mcp`

### Registry Listings

- [x] MCP Official Registry — Published via `mcp-publisher` CLI
- [x] Glama — Listed with AAA score, 37 tools detected
- [x] awesome-mcp-servers — PR to `punkpeye/awesome-mcp-servers` (Cloud Platforms section)
- [x] ToolSDK Registry — PR to `toolsdk-ai/toolsdk-mcp-registry` (cloud-platforms)
- [x] awesome-devops-mcp-servers — PR to `rohitg00/awesome-devops-mcp-servers`
- [x] mcpservers.org — Listed
- [x] mcp.so — Auto-indexed from GitHub
- [ ] `appcypher/awesome-mcp-servers` — PR to second-largest MCP list

---

## Non-Goals

These are explicitly out of scope for this MCP server:

- **Replacing the Spinnaker UI (Deck)** — This server provides programmatic access for LLMs, not a full management interface
- **Direct cloud provider operations** — Spinnaker abstracts cloud providers; this server exposes Spinnaker's abstractions, not raw AWS/GCP/Azure APIs
- **Spinnaker installation or configuration** — Use Halyard/Operator/kleat for Spinnaker setup; this server consumes a running Gate
- **Multi-Gate routing** — Each MCP server instance connects to one Gate endpoint; run multiple instances for multiple Spinnaker deployments

---

## Contributing

Contributions toward any roadmap item are welcome. If you're planning to work on a larger feature (Phase 4+), please open an issue first to discuss the approach.

Priority is given to features that benefit the most common Spinnaker workflows: pipeline management, deployment monitoring, and canary analysis.
