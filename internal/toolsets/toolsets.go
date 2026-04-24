package toolsets

import (
	"fmt"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolEntry pairs a tool definition with its handler.
type ToolEntry struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// Group names for individual tool categories.
const (
	GroupApplications   = "applications"
	GroupPipelines      = "pipelines"
	GroupExecutions     = "executions"
	GroupStrategies     = "strategies"
	GroupInfrastructure = "infrastructure"
	GroupTasks          = "tasks"
)

// Meta-group names that resolve to multiple groups or annotation filters.
const (
	MetaAll      = "all"
	MetaReadonly = "readonly"
	MetaMutating = "mutating"
)

// AllGroups lists every valid individual group name.
var AllGroups = []string{
	GroupApplications,
	GroupPipelines,
	GroupExecutions,
	GroupStrategies,
	GroupInfrastructure,
	GroupTasks,
}

// allValidNames lists every valid name (groups + meta-groups) for error messages.
var allValidNames = []string{
	MetaAll, MetaReadonly, MetaMutating,
	GroupApplications, GroupPipelines, GroupExecutions,
	GroupStrategies, GroupInfrastructure, GroupTasks,
}

// BuildTools returns all tool entries for the given gate client, organized by group.
func BuildTools(gate *client.GateClient) map[string][]ToolEntry {
	reg := make(map[string][]ToolEntry)

	add := func(group string, fn func(*client.GateClient) (mcp.Tool, server.ToolHandlerFunc)) {
		t, h := fn(gate)
		reg[group] = append(reg[group], ToolEntry{Tool: t, Handler: h})
	}

	// Applications
	add(GroupApplications, tools.NewListApplications)
	add(GroupApplications, tools.NewGetApplication)

	// Pipelines
	add(GroupPipelines, tools.NewListPipelines)
	add(GroupPipelines, tools.NewGetPipeline)
	add(GroupPipelines, tools.NewTriggerPipeline)
	add(GroupPipelines, tools.NewSavePipeline)
	add(GroupPipelines, tools.NewUpdatePipeline)
	add(GroupPipelines, tools.NewDeletePipeline)
	add(GroupPipelines, tools.NewGetPipelineHistory)

	// Executions
	add(GroupExecutions, tools.NewGetExecution)
	add(GroupExecutions, tools.NewListExecutions)
	add(GroupExecutions, tools.NewSearchExecutions)
	add(GroupExecutions, tools.NewCancelExecution)
	add(GroupExecutions, tools.NewPauseExecution)
	add(GroupExecutions, tools.NewResumeExecution)
	add(GroupExecutions, tools.NewRestartStage)
	add(GroupExecutions, tools.NewEvaluateExpression)

	// Strategies
	add(GroupStrategies, tools.NewListStrategies)
	add(GroupStrategies, tools.NewSaveStrategy)
	add(GroupStrategies, tools.NewDeleteStrategy)

	// Infrastructure
	add(GroupInfrastructure, tools.NewListServerGroups)
	add(GroupInfrastructure, tools.NewListLoadBalancers)
	add(GroupInfrastructure, tools.NewListClusters)
	add(GroupInfrastructure, tools.NewGetCluster)
	add(GroupInfrastructure, tools.NewGetScalingActivities)
	add(GroupInfrastructure, tools.NewGetTargetServerGroup)
	add(GroupInfrastructure, tools.NewListFirewalls)
	add(GroupInfrastructure, tools.NewGetFirewall)
	add(GroupInfrastructure, tools.NewGetInstance)
	add(GroupInfrastructure, tools.NewGetConsoleOutput)
	add(GroupInfrastructure, tools.NewFindImages)
	add(GroupInfrastructure, tools.NewGetImageTags)
	add(GroupInfrastructure, tools.NewListNetworks)
	add(GroupInfrastructure, tools.NewListSubnets)
	add(GroupInfrastructure, tools.NewListAccounts)
	add(GroupInfrastructure, tools.NewGetAccount)

	// Tasks
	add(GroupTasks, tools.NewGetTask)

	return reg
}

// isReadOnly returns true if the tool annotation marks it as read-only.
func isReadOnly(t mcp.Tool) bool {
	if t.Annotations.ReadOnlyHint != nil {
		return *t.Annotations.ReadOnlyHint
	}
	return false
}

// Resolve takes the raw toolsets string (comma-separated) and returns the
// filtered list of ToolEntry to register. Returns an error if any name is invalid.
func Resolve(raw string, allTools map[string][]ToolEntry) ([]ToolEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, MetaAll) {
		return flatten(allTools, AllGroups), nil
	}

	names := strings.Split(raw, ",")
	for i := range names {
		names[i] = strings.TrimSpace(strings.ToLower(names[i]))
	}

	// Check for meta-groups first (they can't be mixed with regular groups)
	for _, name := range names {
		switch name {
		case MetaReadonly:
			return filterByAnnotation(allTools, true), nil
		case MetaMutating:
			return filterByAnnotation(allTools, false), nil
		case MetaAll:
			return flatten(allTools, AllGroups), nil
		}
	}

	// Validate all names
	for _, name := range names {
		if !isValidGroup(name) {
			return nil, fmt.Errorf("unknown toolset %q; valid values: %s", name, strings.Join(allValidNames, ", "))
		}
	}

	return flatten(allTools, names), nil
}

func isValidGroup(name string) bool {
	for _, g := range AllGroups {
		if g == name {
			return true
		}
	}
	return false
}

func flatten(allTools map[string][]ToolEntry, groups []string) []ToolEntry {
	seen := make(map[string]bool)
	var result []ToolEntry
	for _, g := range groups {
		for _, entry := range allTools[g] {
			if !seen[entry.Tool.Name] {
				seen[entry.Tool.Name] = true
				result = append(result, entry)
			}
		}
	}
	return result
}

func filterByAnnotation(allTools map[string][]ToolEntry, wantReadOnly bool) []ToolEntry {
	var result []ToolEntry
	for _, g := range AllGroups {
		for _, entry := range allTools[g] {
			if isReadOnly(entry.Tool) == wantReadOnly {
				result = append(result, entry)
			}
		}
	}
	return result
}
