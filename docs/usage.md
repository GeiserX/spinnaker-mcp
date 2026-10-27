# Usage

The server exposes 38 tools, 10 resources and 5 prompts. [Toolsets](configuration.md#toolsets) choose which
tools are registered, and by default only the 27 read tools are; resources and prompts are always there.

## Tools

| Category | Tool | Description |
|----------|------|-------------|
| **Applications** | `list_applications` | List all Spinnaker applications |
| | `get_application` | Get detailed application info (accounts, clusters, attributes) |
| **Pipelines** | `list_pipelines` | List pipeline configurations for an application |
| | `get_pipeline` | Get a specific pipeline's full configuration |
| | `trigger_pipeline` | Trigger a pipeline with optional parameters |
| | `save_pipeline` | Save/create a pipeline definition |
| | `update_pipeline` | Update an existing pipeline definition |
| | `delete_pipeline` | Delete a pipeline definition |
| | `get_pipeline_history` | Get revision history for a pipeline config |
| **Executions** | `list_executions` | List recent executions, filterable by status |
| | `get_execution` | Compact summary by default (status, per-stage timing and first error, judgments waiting); `summary=false` for the full JSON |
| | `search_executions` | Search by pipeline name, trigger type, status and trigger time, with paging |
| | `cancel_execution` | Cancel a running execution with optional reason, or `force` it |
| | `pause_execution` | Pause a running execution at the current stage |
| | `resume_execution` | Resume a paused execution |
| | `restart_stage` | Restart a failed stage within an execution |
| | `evaluate_expression` | Evaluate a SpEL expression against an execution |
| | `judge_stage` | Approve (`continue`) or reject (`stop`) a running Manual Judgment stage |
| **Strategies** | `list_strategies` | List deployment strategy configurations |
| | `save_strategy` | Create or update a deployment strategy |
| | `delete_strategy` | Delete a deployment strategy |
| **Infrastructure** | `list_server_groups` | List server groups (deployment targets) with instance counts |
| | `list_load_balancers` | List load balancers across all accounts and regions |
| | `list_clusters` | List cluster names grouped by account |
| | `get_cluster` | Get cluster details including server groups |
| | `get_scaling_activities` | Get scaling activities for a cluster |
| | `get_target_server_group` | Server group by selector: current, previous, oldest, largest, fail |
| | `list_firewalls` | List all firewalls/security groups across accounts |
| | `get_firewall` | Get firewall details by account, region and name, for any provider |
| | `get_instance` | Get instance details (health, metadata, launch time) |
| | `get_console_output` | Get instance console output for debugging |
| | `find_images` | Search for machine images by tags, region, account |
| | `get_image_tags` | List image tags for a repository (Docker registry by default) |
| | `list_networks` | List VPCs/networks by cloud provider |
| | `list_subnets` | List subnets by cloud provider |
| | `list_accounts` | List all configured cloud accounts/credentials |
| | `get_account` | Get account details and permissions |
| **Tasks** | `get_task` | Get orchestration task status (deploy, resize, rollback) |

Eleven tools change Spinnaker, and they are registered only when `TOOLSETS` asks for them. Nine are marked
destructive, because they can deploy, stop or overwrite something: `trigger_pipeline`, `judge_stage`,
`save_pipeline`, `update_pipeline`, `delete_pipeline`, `cancel_execution`, `restart_stage`, `save_strategy`
and `delete_strategy`. `pause_execution` and `resume_execution` change state but destroy nothing. The other
27 only read. `evaluate_expression` sends the expression to Gate in a POST but changes nothing, so it is
marked read-only.

Every tool has a title and sets all four MCP hints (`readOnlyHint`, `destructiveHint`, `idempotentHint`,
`openWorldHint`) explicitly, so a client that reads them can auto-approve the reads and ask before the
destructive nine. Read tools declare an `outputSchema` and return the Gate data as `structuredContent` as
well as text: a JSON array from Gate comes back as `{items, count}`, an object as is.

## Resources

All return JSON and only read.

| URI | Returns |
|-----|---------|
| `spinnaker://applications` | All applications with metadata |
| `spinnaker://accounts` | All cloud accounts with provider type and environment |
| `spinnaker://application/{name}` | Application details: accounts, clusters, attributes |
| `spinnaker://application/{name}/pipelines` | All pipeline configurations of the application |
| `spinnaker://application/{name}/executions` | Recent pipeline executions with status and timing |
| `spinnaker://application/{name}/clusters` | Clusters grouped by account |
| `spinnaker://application/{name}/server-groups` | Server groups with instance counts, image and capacity |
| `spinnaker://application/{name}/load-balancers` | Load balancers across all accounts and regions |
| `spinnaker://execution/{id}` | Full execution details: stages, outputs, timing |
| `spinnaker://account/{name}` | Account details: regions, permissions, provider metadata |

## Prompts

A prompt returns instructions for the assistant; it calls nothing by itself.

| Prompt | Arguments | Asks the assistant to |
|--------|-----------|-----------------------|
| `deploy-review` | `application`, `pipeline` | Review a pipeline configuration before triggering a deployment |
| `incident-response` | `application`, `execution_id` (optional; the most recent failed execution if omitted) | Investigate a failed or stuck deployment |
| `pipeline-audit` | `application`, `pipeline` | Audit a pipeline configuration for best practices |
| `infra-overview` | `application`, `account` (optional; all accounts if omitted) | Summarize the complete infrastructure state for an application |
| `rollback-plan` | `application`, `cluster`, `account`, `region` | Generate a rollback strategy for a deployment |

## On the wire

Everything is JSON-RPC 2.0, and the server speaks both protocol eras on stdio and HTTP:

- **2026-07-28**: the client calls `server/discover`, then sends every request with its protocol version
  and capabilities in `_meta`. There is no handshake and no session. List results carry `ttlMs` and
  `cacheScope`.
- **2025-11-25 down to 2024-11-05**: the client sends `initialize`, then `tools/list`, `resources/list`,
  `prompts/list`, `tools/call`, `resources/read` and `prompts/get`. Claude Code, Claude Desktop, Cursor and
  VS Code connect this way today.
