package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register adds all Spinnaker prompt templates to the MCP server.
func Register(s *server.MCPServer) {
	s.AddPrompt(deployReview())
	s.AddPrompt(incidentResponse())
	s.AddPrompt(pipelineAudit())
	s.AddPrompt(infraOverview())
	s.AddPrompt(rollbackPlan())
}

func deployReview() (mcp.Prompt, server.PromptHandlerFunc) {
	prompt := mcp.NewPrompt("deploy-review",
		mcp.WithPromptDescription("Review a pipeline configuration before triggering a deployment"),
		mcp.WithArgument("application",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Spinnaker application name"),
		),
		mcp.WithArgument("pipeline",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Pipeline name to review"),
		),
	)

	handler := func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		app := req.Params.Arguments["application"]
		pipeline := req.Params.Arguments["pipeline"]

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Deploy review for %s/%s", app, pipeline),
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(`Review the pipeline "%s" in application "%s" before deployment.

Steps:
1. Use get_pipeline to fetch the full pipeline configuration for application "%s", pipeline "%s"
2. Use list_executions (application "%s", limit 5) to check recent execution history
3. Review the pipeline and report:
   - Pipeline stages and their types (deploy, wait, manual judgment, etc.)
   - Trigger configuration (manual, webhook, cron, etc.)
   - Parameter definitions and their defaults
   - Whether manual judgment gates exist before production stages
   - Notification configuration (email, Slack, etc.)
   - Any hardcoded image references (should use artifact bindings instead)
   - Recent execution success/failure rate
4. Provide a go/no-go recommendation with specific concerns if any`, pipeline, app, app, pipeline, app),
					},
				},
			},
		}, nil
	}

	return prompt, handler
}

func incidentResponse() (mcp.Prompt, server.PromptHandlerFunc) {
	prompt := mcp.NewPrompt("incident-response",
		mcp.WithPromptDescription("Investigate a failed or stuck deployment"),
		mcp.WithArgument("application",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Spinnaker application name"),
		),
		mcp.WithArgument("execution_id",
			mcp.ArgumentDescription("Execution ID to investigate (if omitted, targets the most recent failed execution)"),
		),
	)

	handler := func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		app := req.Params.Arguments["application"]
		execID := req.Params.Arguments["execution_id"]

		var step1 string
		if execID != "" {
			step1 = fmt.Sprintf("1. Use get_execution with execution_id \"%s\" to get the full execution details", execID)
		} else {
			step1 = fmt.Sprintf("1. Use search_executions for application \"%s\" with statuses \"TERMINAL\" to find the most recent failed execution, then use get_execution on it", app)
		}

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Incident response for %s", app),
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(`Investigate a failed or stuck deployment for application "%s".

Steps:
%s
2. Identify the failed stage and extract:
   - Stage type and name
   - Error message and exception details
   - Stage execution time and timeout configuration
   - Upstream/downstream stage dependencies
3. Use list_server_groups for application "%s" to check current infrastructure health
4. If instances are unhealthy, use get_instance and get_console_output to get details
5. Use get_scaling_activities if relevant to check for scaling failures
6. Provide a diagnosis:
   - Root cause analysis
   - Affected infrastructure
   - Recommended remediation (restart stage, rollback, manual fix)
   - Whether it's safe to retry`, app, step1, app),
					},
				},
			},
		}, nil
	}

	return prompt, handler
}

func pipelineAudit() (mcp.Prompt, server.PromptHandlerFunc) {
	prompt := mcp.NewPrompt("pipeline-audit",
		mcp.WithPromptDescription("Audit a pipeline configuration for best practices"),
		mcp.WithArgument("application",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Spinnaker application name"),
		),
		mcp.WithArgument("pipeline",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Pipeline name to audit"),
		),
	)

	handler := func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		app := req.Params.Arguments["application"]
		pipeline := req.Params.Arguments["pipeline"]

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Pipeline audit for %s/%s", app, pipeline),
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(`Audit the pipeline "%s" in application "%s" for best practices.

Steps:
1. Use get_pipeline to fetch the full configuration for application "%s", pipeline "%s"
2. Check for these issues and report findings:
   - Missing notification stages (no alerts on failure)
   - No manual judgment gates before production deploy stages
   - Hardcoded image references instead of artifact bindings
   - Unused pipeline parameters (defined but never referenced in SpEL)
   - Missing rollback strategies (no rollback stage after deploy)
   - Stages without timeout configuration (risk of stuck pipelines)
   - Overly permissive triggers (e.g., triggering on every commit without filters)
   - Missing expected artifacts declarations
   - Deploy stages without a preceding bake/find image stage
   - Concurrent execution settings (parallel vs sequential)
3. Rate each finding as: CRITICAL, WARNING, or INFO
4. Provide specific remediation steps for each issue`, pipeline, app, app, pipeline),
					},
				},
			},
		}, nil
	}

	return prompt, handler
}

func infraOverview() (mcp.Prompt, server.PromptHandlerFunc) {
	prompt := mcp.NewPrompt("infra-overview",
		mcp.WithPromptDescription("Summarize the complete infrastructure state for an application"),
		mcp.WithArgument("application",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Spinnaker application name"),
		),
		mcp.WithArgument("account",
			mcp.ArgumentDescription("Cloud account to filter by (if omitted, covers all accounts)"),
		),
	)

	handler := func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		app := req.Params.Arguments["application"]
		account := req.Params.Arguments["account"]

		accountFilter := ""
		if account != "" {
			accountFilter = fmt.Sprintf(" in account \"%s\"", account)
		}

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Infrastructure overview for %s%s", app, accountFilter),
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(`Provide a comprehensive infrastructure overview for application "%s"%s.

Steps:
1. Use list_server_groups for application "%s" to get all deployment targets
2. Use list_clusters for application "%s" to see cluster organization
3. Use list_load_balancers for application "%s" to see traffic routing
4. Summarize:
   - Server groups by cluster: name, region, instance count, health status, image/build
   - Load balancers: name, type, target groups, health check configuration
   - Cluster organization: which clusters exist in which accounts/regions
   - Overall health: total instances, healthy vs unhealthy ratio
   - Current capacity: min/max/desired for each server group
5. Flag any concerns:
   - Server groups with zero healthy instances
   - Clusters with only one server group (no rollback target)
   - Mismatched instance counts between regions (asymmetric deployment)`, app, accountFilter, app, app, app),
					},
				},
			},
		}, nil
	}

	return prompt, handler
}

func rollbackPlan() (mcp.Prompt, server.PromptHandlerFunc) {
	prompt := mcp.NewPrompt("rollback-plan",
		mcp.WithPromptDescription("Generate a rollback strategy for a deployment"),
		mcp.WithArgument("application",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Spinnaker application name"),
		),
		mcp.WithArgument("cluster",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Cluster name to roll back"),
		),
		mcp.WithArgument("account",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Cloud account name"),
		),
		mcp.WithArgument("region",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Cloud region or scope (e.g., us-east-1)"),
		),
	)

	handler := func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		app := req.Params.Arguments["application"]
		cluster := req.Params.Arguments["cluster"]
		account := req.Params.Arguments["account"]
		region := req.Params.Arguments["region"]

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Rollback plan for %s/%s in %s/%s", app, cluster, account, region),
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(`Generate a rollback strategy for cluster "%s" in application "%s" (account: "%s", region: "%s").

Steps:
1. Use get_cluster (application "%s", account "%s", cluster "%s") to see all server groups
2. Use get_target_server_group with target "current" (cloud_provider from cluster, scope "%s") to identify the active deployment
3. Use get_target_server_group with target "ancestor" to identify the rollback target
4. Use get_scaling_activities (application "%s", account "%s", cluster "%s") to understand recent changes
5. Build the rollback plan:
   - Current server group: name, image, instance count, health
   - Rollback target: name, image, instance count, when it was last active
   - Expected capacity changes after rollback
   - Load balancer re-targeting steps (if applicable)
   - Estimated rollback time
6. Provide verification steps:
   - Health checks to confirm after rollback
   - Metrics to monitor
   - How to confirm traffic is flowing to rolled-back server group`, cluster, app, account, region, app, account, cluster, region, app, account, cluster),
					},
				},
			},
		}, nil
	}

	return prompt, handler
}
