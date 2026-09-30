# Usage

The server exposes 37 tools, 10 resources and 5 prompts. [Toolsets](configuration.md#toolsets) choose which
tools are registered; resources and prompts are always there.

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
| | `get_execution` | Get full execution details (stages, outputs, timing) |
| | `search_executions` | Rich search by trigger type, time range, status |
| | `cancel_execution` | Cancel a running execution with optional reason |
| | `pause_execution` | Pause a running execution at the current stage |
| | `resume_execution` | Resume a paused execution |
| | `restart_stage` | Restart a failed stage within an execution |
| | `evaluate_expression` | Evaluate a SpEL expression against an execution |
| **Strategies** | `list_strategies` | List deployment strategy configurations |
| | `save_strategy` | Create or update a deployment strategy |
| | `delete_strategy` | Delete a deployment strategy |
| **Infrastructure** | `list_server_groups` | List server groups (deployment targets) with instance counts |
| | `list_load_balancers` | List load balancers across all accounts and regions |
| | `list_clusters` | List cluster names grouped by account |
| | `get_cluster` | Get cluster details including server groups |
| | `get_scaling_activities` | Get scaling activities for a cluster |
| | `get_target_server_group` | Target-based server group lookup (newest, oldest, etc.) |
| | `list_firewalls` | List all firewalls/security groups across accounts |
| | `get_firewall` | Get firewall details by account, region, and name |
| | `get_instance` | Get instance details (health, metadata, launch time) |
| | `get_console_output` | Get instance console output for debugging |
| | `find_images` | Search for machine images by tags, region, account |
| | `get_image_tags` | List image tags for a repository |
| | `list_networks` | List VPCs/networks by cloud provider |
| | `list_subnets` | List subnets by cloud provider |
| | `list_accounts` | List all configured cloud accounts/credentials |
| | `get_account` | Get account details and permissions |
| **Tasks** | `get_task` | Get orchestration task status (deploy, resize, rollback) |

Ten tools change Spinnaker: `trigger_pipeline`, `save_pipeline`, `update_pipeline`, `delete_pipeline`,
`cancel_execution`, `pause_execution`, `resume_execution`, `restart_stage`, `save_strategy` and
`delete_strategy`. The other 27 only read. `evaluate_expression` sends the expression to Gate in a POST but
changes nothing, so it is marked read-only. Every tool carries MCP annotations (`readOnlyHint`,
`destructiveHint`), so a client that reads them can ask before the ten, and treats `delete_pipeline` and
`delete_strategy` as destructive.

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

Everything is JSON-RPC 2.0. A client sends `initialize`, lists what is there with `tools/list`,
`resources/list` and `prompts/list`, then calls `tools/call`, `resources/read` or `prompts/get`.
