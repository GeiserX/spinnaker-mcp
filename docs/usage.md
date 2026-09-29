# Tools

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

Everything is exposed over JSON-RPC. LLMs and agents can: `initialize` -> `listTools` -> `callTool` and interact with your Spinnaker deployments.
