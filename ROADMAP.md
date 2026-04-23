# Roadmap

This document outlines the planned features and improvements for Spinnaker-MCP. The Spinnaker Gate API is extensive (~150+ endpoints across 60+ controllers), and this MCP server currently covers the highest-value operations. The roadmap is organized by priority and grouped by API domain.

## Current State (v0.1.0)

**13 tools** covering the core deployment workflow:

- Applications: list, get
- Pipelines: list, get config, trigger
- Executions: list, get, cancel, pause, resume
- Infrastructure: server groups, load balancers
- Tasks: get status

**Authentication**: Bearer token, basic auth, x509 client certificates

**Transports**: stdio, Streamable HTTP

---

## Phase 1 — Pipeline Lifecycle (v0.2.0)

Complete coverage of pipeline CRUD and execution management.

### Pipeline Configuration Management

- [ ] `save_pipeline` — Save/create a pipeline definition (`POST /pipelines`)
- [ ] `update_pipeline` — Update an existing pipeline definition (`PUT /pipelines/{id}`)
- [ ] `delete_pipeline` — Delete a pipeline definition (`DELETE /pipelines/{application}/{pipelineName}`)
- [ ] `get_pipeline_history` — Get revision history for a pipeline config (`GET /pipelineConfigs/{id}/history`)

### Execution Control

- [ ] `restart_stage` — Restart a failed stage within an execution (`PUT /pipelines/{id}/stages/{stageId}/restart`)
- [ ] `search_executions` — Rich search across executions by trigger type, time range, status, event ID (`GET /applications/{app}/executions/search`)
- [ ] `evaluate_expression` — Evaluate a SpEL pipeline expression against an execution context (`POST /pipelines/{id}/evaluate/{expression}`)

### Strategy Management

- [ ] `list_strategies` — List deployment strategy configurations for an application
- [ ] `save_strategy` — Create or update a deployment strategy
- [ ] `delete_strategy` — Delete a deployment strategy

---

## Phase 2 — Infrastructure Deep Dive (v0.3.0)

Read-only visibility into Spinnaker-managed infrastructure.

### Clusters

- [ ] `list_clusters` — List cluster names grouped by account (`GET /applications/{app}/clusters`)
- [ ] `get_cluster` — Get cluster details including server groups (`GET /applications/{app}/clusters/{account}/{cluster}`)
- [ ] `get_scaling_activities` — Get scaling activities for a cluster (`GET .../scalingActivities`)
- [ ] `get_target_server_group` — Target-based server group lookup (newest, oldest, ancestor)

### Security Groups / Firewalls

- [ ] `list_firewalls` — List all firewalls/security groups across accounts
- [ ] `get_firewall` — Get firewall details by account, region, and name

### Instances

- [ ] `get_instance` — Get instance details (health, metadata, launch time)
- [ ] `get_console_output` — Get instance console output for debugging

### Images

- [ ] `find_images` — Search for machine images by tags, region, account
- [ ] `get_image_tags` — List image tags for a repository

### Networks and Subnets

- [ ] `list_networks` — List VPCs/networks by cloud provider
- [ ] `list_subnets` — List subnets by cloud provider

### Credentials / Accounts

- [ ] `list_accounts` — List all configured cloud accounts/credentials
- [ ] `get_account` — Get account details and permissions

---

## Phase 3 — Tasks and Operations (v0.4.0)

Full task management and operational actions.

### Task Management

- [ ] `create_task` — Create an ad-hoc orchestration task (scale, destroy, rollback, etc.) (`POST /tasks`)
- [ ] `cancel_task` — Cancel a running task (`PUT /tasks/{id}/cancel`)
- [ ] `list_application_tasks` — List all tasks for an application with status filtering
- [ ] `get_task_details` — Get detailed task execution information

### Manual Judgment

- [ ] `approve_stage` — Approve a manual judgment stage (via `PATCH /pipelines/{id}/stages/{stageId}`)
- [ ] `reject_stage` — Reject a manual judgment stage

This is critical for interactive pipeline workflows where an LLM assists in deployment decisions.

---

## Phase 4 — Canary Analysis (v0.5.0)

Integration with Kayenta for automated canary analysis.

### Canary Configs

- [ ] `list_canary_configs` — List all canary configuration templates
- [ ] `get_canary_config` — Get a specific canary config
- [ ] `save_canary_config` — Create or update a canary config
- [ ] `delete_canary_config` — Delete a canary config

### Canary Executions

- [ ] `start_canary` — Start a canary analysis execution (`POST /v2/canaries/canary/{configId}`)
- [ ] `get_canary_result` — Get canary analysis results with metric comparisons
- [ ] `list_canary_results` — List canary results for an application

### Canary Metadata

- [ ] `list_canary_judges` — List configured canary judges
- [ ] `list_metrics_services` — List available metrics service metadata
- [ ] `list_canary_accounts` — List Kayenta account integrations

---

## Phase 5 — Artifacts and Builds (v0.6.0)

Artifact management and CI/CD build integration.

### Artifacts

- [ ] `list_artifact_accounts` — List configured artifact sources (Docker, GCS, S3, Maven, etc.)
- [ ] `list_artifact_names` — List artifact names for an account
- [ ] `list_artifact_versions` — List available versions for an artifact
- [ ] `fetch_artifact` — Fetch artifact contents (streaming)

### Builds (v3 API)

- [ ] `list_build_masters` — List CI integrations (Jenkins, Travis, Wercker, Concourse, Google Cloud Build)
- [ ] `list_build_jobs` — List jobs for a build master
- [ ] `get_build` — Get a specific build result
- [ ] `list_builds` — List builds for a job

### Webhooks

- [ ] `trigger_webhook` — Post a webhook to trigger pipelines (`POST /webhooks/webhook/{source}`)
- [ ] `list_preconfigured_webhooks` — List preconfigured webhook stage types

---

## Phase 6 — Managed Delivery / Keel (v0.7.0)

Integration with Spinnaker's declarative delivery system.

### Delivery Configs

- [ ] `get_delivery_config` — Get a delivery config definition
- [ ] `save_delivery_config` — Create or update a delivery config
- [ ] `delete_delivery_config` — Delete a delivery config
- [ ] `validate_delivery_config` — Validate a delivery config before saving
- [ ] `diff_delivery_config` — Diff changes to a delivery config

### Resource Management

- [ ] `get_managed_resource` — Get a managed resource by ID
- [ ] `get_resource_status` — Get current resource status
- [ ] `pause_resource` — Pause management of a resource
- [ ] `resume_resource` — Resume management of a resource
- [ ] `export_resource` — Export a resource definition from running infrastructure

### Environment Control

- [ ] `pin_artifact` — Pin an artifact version in an environment
- [ ] `unpin_artifact` — Remove an artifact pin
- [ ] `veto_artifact` — Veto an artifact version from promotion
- [ ] `mark_artifact_bad` — Mark an artifact version as bad
- [ ] `mark_artifact_good` — Mark an artifact version as good
- [ ] `list_constraints` — List constraint states for an environment
- [ ] `update_constraint` — Approve or reject an environment constraint

### Application-Level Control

- [ ] `get_managed_application` — Get managed delivery status for an application
- [ ] `pause_managed_application` — Pause all managed delivery for an application
- [ ] `resume_managed_application` — Resume managed delivery for an application

---

## Phase 7 — Kubernetes Native (v0.8.0)

Kubernetes-specific operations for the most common Spinnaker cloud provider.

### Manifests

- [ ] `get_manifest` — Get a Kubernetes manifest by account, namespace, and name
- [ ] `deploy_manifest` — Deploy a manifest via ad-hoc task
- [ ] `delete_manifest` — Delete a manifest via ad-hoc task
- [ ] `patch_manifest` — Patch a manifest via ad-hoc task
- [ ] `scale_manifest` — Scale a manifest (replicas) via ad-hoc task
- [ ] `rollback_manifest` — Rollback to a previous manifest version
- [ ] `undo_rollout` — Undo a Kubernetes rollout

### Server Group Managers

- [ ] `list_server_group_managers` — List Kubernetes Deployments/StatefulSets/ReplicaSets

---

## Phase 8 — Pipeline Templates (v0.9.0)

Reusable pipeline templates for standardized workflows.

### V2 Templates

- [ ] `list_pipeline_templates` — List all pipeline templates
- [ ] `get_pipeline_template` — Get a specific template with version
- [ ] `create_pipeline_template` — Create a new pipeline template
- [ ] `update_pipeline_template` — Update an existing template
- [ ] `delete_pipeline_template` — Delete a template
- [ ] `plan_pipeline_template` — Plan/preview a template config before saving
- [ ] `list_template_dependents` — List pipelines that depend on a template

---

## Phase 9 — Observability and Metadata (v0.10.0)

### Projects

- [ ] `list_projects` — List all Spinnaker projects (groups of applications)
- [ ] `get_project` — Get project details with associated clusters and pipelines

### Entity Tags

- [ ] `list_entity_tags` — List entity tags with rich filtering
- [ ] `create_entity_tags` — Create or update entity tags
- [ ] `delete_entity_tags` — Delete specific tags from an entity

### Notifications

- [ ] `get_notification_config` — Get notification preferences for an application
- [ ] `save_notification_config` — Save notification preferences
- [ ] `list_notification_types` — List available notification types (Slack, email, PagerDuty, etc.)

### Search

- [ ] `search_infrastructure` — Global search across all Spinnaker resources (applications, server groups, instances, load balancers)

### System

- [ ] `get_gate_version` — Get the Spinnaker Gate version
- [ ] `get_current_user` — Get the authenticated user's info and roles
- [ ] `validate_cron` — Validate a cron expression for pipeline triggers

---

## Cross-Cutting Improvements

### Authentication

- [ ] OAuth2/OIDC token refresh — Automatic token refresh for OAuth-based auth
- [ ] Cookie-based auth — Support session cookies from browser SSO flows
- [ ] IAP (Identity-Aware Proxy) — Google Cloud IAP service account auth
- [ ] SAML assertion forwarding — For environments using SAML-based SSO

### Transport and Protocol

- [ ] MCP Resources — Expose read-only data as MCP resources (e.g., `spinnaker://application/{name}`, `spinnaker://execution/{id}`)
- [ ] MCP Prompts — Pre-built prompt templates for common workflows (deploy review, incident response, canary analysis)
- [ ] Session-aware tools — Stateful context tracking for multi-turn deployment conversations
- [ ] Rate limit awareness — Parse `X-RateLimit-*` headers and expose remaining capacity

### Developer Experience

- [ ] `--toolsets` CLI flag — Enable/disable tool groups (e.g., `--toolsets=pipelines,executions`)
- [ ] `--tools` CLI flag — Enable specific tools by name
- [ ] OpenTelemetry integration — Trace and metric export for tool call volume and latency
- [ ] Structured logging — JSON log output with configurable verbosity
- [ ] Health check endpoint — `/healthz` for container orchestration

### Distribution

- [ ] Homebrew tap — `brew install GeiserX/tap/spinnaker-mcp`
- [ ] AUR package — Arch Linux user repository
- [ ] Helm chart — Deploy as a sidecar or standalone service in Kubernetes
- [ ] Nix flake — Reproducible builds for Nix users

### Ecosystem Integrations

- [ ] n8n community node (`n8n-nodes-spinnaker-mcp`) — Workflow automation for Spinnaker operations
- [ ] Home Assistant integration (`spinnaker-ha`) — Dashboard cards for deployment status
- [ ] Unraid CA template — Community Applications template for self-hosted users
- [ ] Portainer community template — One-click deploy via Portainer
- [ ] CasaOS app — Self-hosted platform integration
- [ ] Runtipi app — Custom app store entry

### Registry Listings

- [ ] MCP Official Registry — `mcp-publisher` CLI submission
- [ ] Glama — AAA score listing
- [ ] mcpservers.org — Web form submission
- [ ] awesome-mcp-servers — PR to curated list
- [ ] mcp.so — Listing

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
