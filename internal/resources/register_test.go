package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newGateWithHandler(t *testing.T, handler http.HandlerFunc) (*client.GateClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	gate, err := client.NewGate(client.GateOptions{BaseURL: srv.URL})
	if err != nil {
		srv.Close()
		t.Fatalf("NewGate: %v", err)
	}
	return gate, srv
}

func makeReq(uri string) mcp.ReadResourceRequest {
	req := mcp.ReadResourceRequest{}
	req.Params.URI = uri
	return req
}

func TestRegister_DoesNotPanic(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})
	defer srv.Close()

	s := server.NewMCPServer("test", "0.0.0",
		server.WithResourceCapabilities(false, false),
	)
	Register(s, gate)
}

func TestHandleListApplications_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications" {
			w.Write([]byte(`[{"name":"app1"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListApplications(gate)
	result, err := h(context.Background(), makeReq("spinnaker://applications"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"name":"app1"}]` {
		t.Errorf("text = %q, want %q", text, `[{"name":"app1"}]`)
	}
}

func TestHandleListApplications_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListApplications(gate)
	_, err := h(context.Background(), makeReq("spinnaker://applications"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListAccounts_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/credentials" {
			w.Write([]byte(`[{"name":"prod"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListAccounts(gate)
	result, err := h(context.Background(), makeReq("spinnaker://accounts"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"name":"prod"}]` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListAccounts_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListAccounts(gate)
	_, err := h(context.Background(), makeReq("spinnaker://accounts"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleGetApplication_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp" {
			w.Write([]byte(`{"name":"myapp"}`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleGetApplication(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `{"name":"myapp"}` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleGetApplication_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleGetApplication(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListPipelines_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp/pipelineConfigs" {
			w.Write([]byte(`[{"name":"deploy"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListPipelines(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp/pipelines"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"name":"deploy"}]` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListPipelines_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListPipelines(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp/pipelines"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListExecutions_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp/pipelines" {
			w.Write([]byte(`[{"id":"e1"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListExecutions(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp/executions"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"id":"e1"}]` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListExecutions_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListExecutions(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp/executions"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListClusters_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp/clusters" {
			w.Write([]byte(`{"prod":["c1"]}`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListClusters(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp/clusters"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `{"prod":["c1"]}` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListClusters_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListClusters(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp/clusters"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListServerGroups_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp/serverGroups" {
			w.Write([]byte(`[{"name":"v001"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListServerGroups(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp/server-groups"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"name":"v001"}]` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListServerGroups_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListServerGroups(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp/server-groups"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleListLoadBalancers_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/applications/myapp/loadBalancers" {
			w.Write([]byte(`[{"name":"lb1"}]`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleListLoadBalancers(gate)
	result, err := h(context.Background(), makeReq("spinnaker://application/myapp/load-balancers"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `[{"name":"lb1"}]` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleListLoadBalancers_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleListLoadBalancers(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/myapp/load-balancers"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleGetExecution_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pipelines/exec-1" {
			w.Write([]byte(`{"id":"exec-1"}`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleGetExecution(gate)
	result, err := h(context.Background(), makeReq("spinnaker://execution/exec-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `{"id":"exec-1"}` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleGetExecution_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleGetExecution(gate)
	_, err := h(context.Background(), makeReq("spinnaker://execution/exec-1"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleGetAccount_Happy(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/credentials/prod" {
			w.Write([]byte(`{"name":"prod"}`))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	})
	defer srv.Close()

	h := handleGetAccount(gate)
	result, err := h(context.Background(), makeReq("spinnaker://account/prod"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result[0].(mcp.TextResourceContents).Text
	if text != `{"name":"prod"}` {
		t.Errorf("text = %q", text)
	}
}

func TestHandleGetAccount_Error(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	})
	defer srv.Close()

	h := handleGetAccount(gate)
	_, err := h(context.Background(), makeReq("spinnaker://account/prod"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractParam_NonStringArgument(t *testing.T) {
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "spinnaker://application/fallback"
	req.Params.Arguments = map[string]any{"name": 42}
	got := extractParam(req, "name")
	if got != "fallback" {
		t.Errorf("extractParam() = %q, want %q (URI fallback)", got, "fallback")
	}
}

func TestExtractParam_EmptyArguments(t *testing.T) {
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "spinnaker://account/staging"
	req.Params.Arguments = map[string]any{}
	got := extractParam(req, "name")
	if got != "staging" {
		t.Errorf("extractParam() = %q, want %q", got, "staging")
	}
}

func TestParseURIParam_PipelineID(t *testing.T) {
	got := parseURIParam("spinnaker://pipeline/pipe-abc", "id")
	if got != "pipe-abc" {
		t.Errorf("parseURIParam() = %q, want %q", got, "pipe-abc")
	}
}

func TestParseURIParam_UnknownKey(t *testing.T) {
	got := parseURIParam("spinnaker://application/myapp", "unknown")
	if got != "" {
		t.Errorf("parseURIParam() = %q, want empty", got)
	}
}

func TestParseURIParam_ShortURI(t *testing.T) {
	got := parseURIParam("spinnaker://application", "name")
	if got != "" {
		t.Errorf("parseURIParam() = %q, want empty", got)
	}
}

func TestHandleGetApplication_EmptyName(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	h := handleGetApplication(gate)
	_, err := h(context.Background(), makeReq("spinnaker://application/"))
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestHandleGetExecution_EmptyID(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	h := handleGetExecution(gate)
	_, err := h(context.Background(), makeReq("spinnaker://execution/"))
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestHandleGetAccount_EmptyName(t *testing.T) {
	gate, srv := newGateWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	h := handleGetAccount(gate)
	_, err := h(context.Background(), makeReq("spinnaker://account/"))
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}
