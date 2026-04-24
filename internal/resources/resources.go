package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register adds all Spinnaker resource templates to the MCP server.
func Register(s *server.MCPServer, gate *client.GateClient) {
	// Static resources (no URI parameters)
	s.AddResource(
		mcp.NewResource(
			"spinnaker://applications",
			"All Applications",
			mcp.WithResourceDescription("List of all Spinnaker applications with metadata"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			data, err := gate.ListApplications(ctx)
			if err != nil {
				return nil, fmt.Errorf("listing applications: %w", err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResource(
		mcp.NewResource(
			"spinnaker://accounts",
			"All Accounts",
			mcp.WithResourceDescription("All configured cloud accounts with provider type and environment"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			data, err := gate.ListAccounts(ctx)
			if err != nil {
				return nil, fmt.Errorf("listing accounts: %w", err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	// Parameterized templates
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}",
			"Application Details",
			mcp.WithTemplateDescription("Application details including accounts, clusters, and attributes"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.GetApplication(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("getting application %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}/pipelines",
			"Application Pipelines",
			mcp.WithTemplateDescription("All pipeline configurations for an application"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.ListPipelines(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("listing pipelines for %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}/executions",
			"Application Executions",
			mcp.WithTemplateDescription("Recent pipeline executions with status and timing"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.ListExecutions(ctx, name, 25, "")
			if err != nil {
				return nil, fmt.Errorf("listing executions for %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}/clusters",
			"Application Clusters",
			mcp.WithTemplateDescription("Clusters grouped by account for an application"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.ListClusters(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("listing clusters for %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}/server-groups",
			"Application Server Groups",
			mcp.WithTemplateDescription("Server groups with instance counts, image, and capacity"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.ListServerGroups(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("listing server groups for %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://application/{name}/load-balancers",
			"Application Load Balancers",
			mcp.WithTemplateDescription("Load balancers across all accounts and regions"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.ListLoadBalancers(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("listing load balancers for %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://execution/{id}",
			"Execution Details",
			mcp.WithTemplateDescription("Full execution details including all stages, outputs, and timing"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			id := extractParam(req, "id")
			data, err := gate.GetExecution(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("getting execution %q: %w", id, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"spinnaker://account/{name}",
			"Account Details",
			mcp.WithTemplateDescription("Account details with regions, permissions, and provider metadata"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name := extractParam(req, "name")
			data, err := gate.GetAccount(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("getting account %q: %w", name, err)
			}
			return []mcp.ResourceContents{textResource(req.Params.URI, data)}, nil
		},
	)
}

func textResource(uri string, data []byte) mcp.TextResourceContents {
	text := string(data)
	if text == "" {
		text = "[]"
	}
	return mcp.TextResourceContents{
		URI:      uri,
		MIMEType: "application/json",
		Text:     text,
	}
}

func extractParam(req mcp.ReadResourceRequest, key string) string {
	if req.Params.Arguments != nil {
		if v, ok := req.Params.Arguments[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	// Fallback: parse from URI
	return parseURIParam(req.Params.URI, key)
}

// parseURIParam extracts a parameter from a spinnaker:// URI by position.
func parseURIParam(uri, key string) string {
	uri = strings.TrimPrefix(uri, "spinnaker://")
	parts := strings.Split(uri, "/")

	switch key {
	case "name":
		if len(parts) >= 2 && (parts[0] == "application" || parts[0] == "account") {
			return parts[1]
		}
	case "id":
		if len(parts) >= 2 && (parts[0] == "execution" || parts[0] == "pipeline") {
			return parts[1]
		}
	}
	return ""
}
