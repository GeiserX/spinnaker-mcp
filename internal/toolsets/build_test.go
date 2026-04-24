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
		GroupExecutions:     8,
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

	if total != 37 {
		t.Errorf("total tools = %d, want 37", total)
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

func TestBuildTools_AllToolsHaveAnnotations(t *testing.T) {
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
			if entry.Tool.Annotations.ReadOnlyHint == nil {
				t.Errorf("group %q tool %q has nil ReadOnlyHint annotation", group, entry.Tool.Name)
			}
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

	// Test "all" resolves to 37
	all, err := Resolve("", tools)
	if err != nil {
		t.Fatalf("Resolve all: %v", err)
	}
	if len(all) != 37 {
		t.Errorf("all = %d tools, want 37", len(all))
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
	if len(ro)+len(mut) != 37 {
		t.Errorf("readonly(%d) + mutating(%d) = %d, want 37", len(ro), len(mut), len(ro)+len(mut))
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
