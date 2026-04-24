package toolsets

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func boolPtr(b bool) *bool { return &b }

func fakeTools() map[string][]ToolEntry {
	noopHandler := server.ToolHandlerFunc(func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, nil
	})

	return map[string][]ToolEntry{
		GroupApplications: {
			{Tool: mcp.Tool{Name: "list_applications", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
			{Tool: mcp.Tool{Name: "get_application", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
		},
		GroupPipelines: {
			{Tool: mcp.Tool{Name: "list_pipelines", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
			{Tool: mcp.Tool{Name: "trigger_pipeline", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(false)}}, Handler: noopHandler},
		},
		GroupExecutions: {
			{Tool: mcp.Tool{Name: "get_execution", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
			{Tool: mcp.Tool{Name: "cancel_execution", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(false)}}, Handler: noopHandler},
		},
		GroupStrategies: {
			{Tool: mcp.Tool{Name: "list_strategies", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
		},
		GroupInfrastructure: {
			{Tool: mcp.Tool{Name: "list_server_groups", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
		},
		GroupTasks: {
			{Tool: mcp.Tool{Name: "get_task", Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, Handler: noopHandler},
		},
	}
}

func TestResolve_EmptyDefaultsToAll(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 9 {
		t.Errorf("expected 9 tools, got %d", len(result))
	}
}

func TestResolve_AllMeta(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("all", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 9 {
		t.Errorf("expected 9 tools, got %d", len(result))
	}
}

func TestResolve_SingleGroup(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("applications", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result))
	}
	for _, entry := range result {
		if entry.Tool.Name != "list_applications" && entry.Tool.Name != "get_application" {
			t.Errorf("unexpected tool: %s", entry.Tool.Name)
		}
	}
}

func TestResolve_MultipleGroups(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("applications,tasks", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 tools, got %d", len(result))
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("Applications,TASKS", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 tools, got %d", len(result))
	}
}

func TestResolve_ReadonlyMeta(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("readonly", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, entry := range result {
		if entry.Tool.Annotations.ReadOnlyHint == nil || !*entry.Tool.Annotations.ReadOnlyHint {
			t.Errorf("tool %s should be read-only", entry.Tool.Name)
		}
	}
	if len(result) != 7 {
		t.Errorf("expected 7 readonly tools, got %d", len(result))
	}
}

func TestResolve_MutatingMeta(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("mutating", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, entry := range result {
		if entry.Tool.Annotations.ReadOnlyHint == nil || *entry.Tool.Annotations.ReadOnlyHint {
			t.Errorf("tool %s should be mutating", entry.Tool.Name)
		}
	}
	if len(result) != 2 {
		t.Errorf("expected 2 mutating tools, got %d", len(result))
	}
}

func TestResolve_InvalidGroup(t *testing.T) {
	tools := fakeTools()
	_, err := Resolve("bogus", tools)
	if err == nil {
		t.Fatal("expected error for invalid group name")
	}
}

func TestResolve_TrimsWhitespace(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("  applications , tasks  ", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 tools, got %d", len(result))
	}
}

func TestResolve_NoDuplicates(t *testing.T) {
	tools := fakeTools()
	result, err := Resolve("applications,applications", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 tools (no duplicates), got %d", len(result))
	}
}

func TestIsReadOnly(t *testing.T) {
	tests := []struct {
		name string
		tool mcp.Tool
		want bool
	}{
		{"read-only true", mcp.Tool{Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(true)}}, true},
		{"read-only false", mcp.Tool{Annotations: mcp.ToolAnnotation{ReadOnlyHint: boolPtr(false)}}, false},
		{"nil hint", mcp.Tool{Annotations: mcp.ToolAnnotation{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isReadOnly(tt.tool); got != tt.want {
				t.Errorf("isReadOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTools_AllGroupsPresent(t *testing.T) {
	// BuildTools requires a real GateClient which we can't create without a server.
	// Instead, verify that AllGroups covers expected names.
	expected := map[string]bool{
		"applications":   true,
		"pipelines":      true,
		"executions":     true,
		"strategies":     true,
		"infrastructure": true,
		"tasks":          true,
	}
	for _, g := range AllGroups {
		if !expected[g] {
			t.Errorf("unexpected group in AllGroups: %s", g)
		}
		delete(expected, g)
	}
	for g := range expected {
		t.Errorf("missing group in AllGroups: %s", g)
	}
}
