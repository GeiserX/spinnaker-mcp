package toolsets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
)

func TestBuildTools_ReturnsAllGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	tools := BuildTools(gate)

	for _, group := range AllGroups {
		entries, ok := tools[group]
		if !ok {
			t.Errorf("missing group %q in BuildTools output", group)
			continue
		}
		if len(entries) == 0 {
			t.Errorf("group %q has no tools", group)
		}
	}
}

func TestBuildTools_ToolCounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	tools := BuildTools(gate)

	expected := map[string]int{
		GroupApplications:   2,
		GroupPipelines:      7,
		GroupExecutions:     9,
		GroupStrategies:     3,
		GroupInfrastructure: 16,
		GroupTasks:          1,
	}

	total := 0
	for group, wantCount := range expected {
		got := len(tools[group])
		if got != wantCount {
			t.Errorf("group %q: got %d tools, want %d", group, got, wantCount)
		}
		total += got
	}

	if total != 38 {
		t.Errorf("total tools = %d, want 38", total)
	}
}

func TestBuildTools_AllToolsHaveHandlers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	tools := BuildTools(gate)
	for group, entries := range tools {
		for _, entry := range entries {
			if entry.Tool.Name == "" {
				t.Errorf("group %q has tool with empty name", group)
			}
			if entry.Handler == nil {
				t.Errorf("group %q tool %q has nil handler", group, entry.Tool.Name)
			}
		}
	}
}

// hints is what each tool must declare: readOnly, destructive, idempotent. Clients
// auto-approve or confirm based on these, so every tool is listed and a new tool fails
// the test until it is added here on purpose.
var wantHints = map[string][3]bool{
	"list_applications": {true, false, true}, "get_application": {true, false, true},
	"list_pipelines": {true, false, true}, "get_pipeline": {true, false, true},
	"get_pipeline_history": {true, false, true},
	"trigger_pipeline":     {false, true, false},
	"save_pipeline":        {false, true, false},
	"update_pipeline":      {false, true, true},
	"delete_pipeline":      {false, true, true},
	"get_execution":        {true, false, true}, "list_executions": {true, false, true},
	"search_executions": {true, false, true}, "evaluate_expression": {true, false, true},
	"cancel_execution":   {false, true, true},
	"pause_execution":    {false, false, true},
	"resume_execution":   {false, false, true},
	"restart_stage":      {false, true, false},
	"judge_stage":        {false, true, false},
	"list_strategies":    {true, false, true},
	"save_strategy":      {false, true, true},
	"delete_strategy":    {false, true, true},
	"list_server_groups": {true, false, true}, "list_load_balancers": {true, false, true},
	"list_clusters": {true, false, true}, "get_cluster": {true, false, true},
	"get_scaling_activities": {true, false, true}, "get_target_server_group": {true, false, true},
	"list_firewalls": {true, false, true}, "get_firewall": {true, false, true},
	"get_instance": {true, false, true}, "get_console_output": {true, false, true},
	"find_images": {true, false, true}, "get_image_tags": {true, false, true},
	"list_networks": {true, false, true}, "list_subnets": {true, false, true},
	"list_accounts": {true, false, true}, "get_account": {true, false, true},
	"get_task": {true, false, true},
}

func TestBuildTools_AnnotationsPerTool(t *testing.T) {
	gate, err := client.NewGate(client.GateOptions{BaseURL: "http://localhost:8084"})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	seen := map[string]bool{}
	for _, entries := range BuildTools(gate) {
		for _, e := range entries {
			name, a := e.Tool.Name, e.Tool.Annotations
			seen[name] = true
			want, ok := wantHints[name]
			if !ok {
				t.Errorf("tool %s has no entry in wantHints", name)
				continue
			}
			if a.ReadOnlyHint == nil || a.DestructiveHint == nil || a.IdempotentHint == nil || a.OpenWorldHint == nil {
				t.Errorf("%s: a hint is unset: %+v", name, a)
				continue
			}
			got := [3]bool{*a.ReadOnlyHint, *a.DestructiveHint, *a.IdempotentHint}
			if got != want {
				t.Errorf("%s: readOnly/destructive/idempotent = %v, want %v", name, got, want)
			}
			if !*a.OpenWorldHint {
				t.Errorf("%s: openWorldHint false, but every tool calls Gate", name)
			}
			if e.Tool.Title == "" || a.Title != e.Tool.Title {
				t.Errorf("%s: title %q / annotation title %q", name, e.Tool.Title, a.Title)
			}
			// Read tools relay Gate JSON and declare an outputSchema for it.
			if *a.ReadOnlyHint && e.Tool.RawOutputSchema == nil {
				t.Errorf("%s: read tool without outputSchema", name)
			}
		}
	}
	for name := range wantHints {
		if !seen[name] {
			t.Errorf("wantHints lists %s, which is not registered", name)
		}
	}
}

func TestBuildTools_ResolveIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	gate, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}

	tools := BuildTools(gate)

	// Test "all" resolves to every tool
	all, err := Resolve("all", tools)
	if err != nil {
		t.Fatalf("Resolve all: %v", err)
	}
	if len(all) != 38 {
		t.Errorf("all = %d tools, want 38", len(all))
	}

	// Test "readonly" + "mutating" = "all"
	ro, err := Resolve("readonly", tools)
	if err != nil {
		t.Fatalf("Resolve readonly: %v", err)
	}
	mut, err := Resolve("mutating", tools)
	if err != nil {
		t.Fatalf("Resolve mutating: %v", err)
	}
	if len(ro) != 27 || len(mut) != 11 {
		t.Errorf("readonly = %d, mutating = %d, want 27 and 11", len(ro), len(mut))
	}

	// Test single group
	pipes, err := Resolve("pipelines", tools)
	if err != nil {
		t.Fatalf("Resolve pipelines: %v", err)
	}
	if len(pipes) != 7 {
		t.Errorf("pipelines = %d tools, want 7", len(pipes))
	}
}
