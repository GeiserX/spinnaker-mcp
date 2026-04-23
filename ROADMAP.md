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

## Phase 3 — Tasks and Manual Judgment

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

### Transport and Protocol

- [x] stdio and Streamable HTTP transports
- [ ] MCP Resources — Expose read-only data as MCP resources (e.g., `spinnaker://application/{name}`)
- [ ] MCP Prompts — Pre-built prompt templates for common workflows (deploy review, incident response)

### Developer Experience

- [x] `--version` / `--help` CLI flags
- [x] Tool annotations (readOnly, mutating, destructive hints)
- [ ] `--toolsets` CLI flag — Enable/disable tool groups (e.g., `--toolsets=pipelines,executions`)
- [ ] Health check endpoint — `/healthz` for container orchestration

### Distribution

- [x] npm — `npx spinnaker-mcp` or `npm install -g spinnaker-mcp`
- [x] Docker Hub — `drumsergio/spinnaker-mcp`
- [ ] Homebrew tap — `brew install GeiserX/tap/spinnaker-mcp`
- [ ] AUR package — Arch Linux user repository
- [ ] Helm chart — Deploy as a sidecar or standalone service in Kubernetes

### Registry Listings

- [x] MCP Official Registry — Published via `mcp-publisher` CLI
- [x] Glama — Listed with release, 37 tools detected
- [x] awesome-mcp-servers — PR to `punkpeye/awesome-mcp-servers` (Cloud Platforms section)
- [x] ToolSDK Registry — PR to `toolsdk-ai/toolsdk-mcp-registry` (cloud-platforms)
- [x] awesome-devops-mcp-servers — PR to `rohitg00/awesome-devops-mcp-servers`
- [ ] mcpservers.org — Web form submission
- [ ] mcp.so — Auto-indexed from GitHub
- [ ] `appcypher/awesome-mcp-servers` — PR to second-largest MCP list
- [ ] `wong2/awesome-mcp-servers` — PR to third-largest MCP list

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
