package tools_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/geiserx/spinnaker-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// These tests pin what each tool sends to Gate: path, query and body. The client tests
// cover the HTTP layer; these catch a tool passing the wrong names or values to it.

type seen struct {
	method, path, query, body string
	n                         int
}

func recordGate(t *testing.T, reply string) (*seen, func(w http.ResponseWriter, r *http.Request)) {
	s := &seen{}
	return s, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		s.method, s.path, s.query, s.body = r.Method, r.URL.EscapedPath(), r.URL.RawQuery, string(b)
		s.n++
		w.Write([]byte(reply))
	}
}

func call(t *testing.T, h func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := h(context.Background(), makeRequest(args))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return res
}

func text(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestWire_SearchExecutionsUsesGateNames(t *testing.T) {
	s, h := recordGate(t, `[]`)
	_, handler := tools.NewSearchExecutions(newGate(t, h))
	res := call(t, handler, map[string]any{
		"application":   "myapp",
		"pipeline_name": "Deploy to Prod",
		"trigger_type":  "webhook,cron",
		"statuses":      "TERMINAL",
		"start_time":    "2026-07-28T00:00:00Z",
		"end_time":      "2026-07-29",
		"size":          float64(5),
		"start_index":   float64(10),
	})
	if res.IsError {
		t.Fatalf("tool error: %s", text(res))
	}
	if s.path != "/applications/myapp/executions/search" {
		t.Errorf("path = %s", s.path)
	}
	start := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC).UnixMilli()
	end := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC).UnixMilli()
	want := map[string]string{
		"pipelineName": "Deploy to Prod", "triggerTypes": "webhook,cron", "statuses": "TERMINAL",
		"triggerTimeStartBoundary": strconv.FormatInt(start, 10),
		"triggerTimeEndBoundary":   strconv.FormatInt(end, 10),
		"size":                     "5", "startIndex": "10",
	}
	q := parseQuery(t, s.query)
	for k, v := range want {
		if q[k] != v {
			t.Errorf("query %s = %q, want %q (query %s)", k, q[k], v, s.query)
		}
	}
	for _, bad := range []string{"triggerType", "startTime", "endTime"} {
		if _, ok := q[bad]; ok {
			t.Errorf("query carries %s, which Gate ignores", bad)
		}
	}

	res = call(t, handler, map[string]any{"application": "myapp", "start_time": "yesterday"})
	if !res.IsError {
		t.Error("an unparseable start_time was sent to Gate")
	}
}

func parseQuery(t *testing.T, raw string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, kv := range strings.Split(raw, "&") {
		k, v, _ := strings.Cut(kv, "=")
		v = strings.ReplaceAll(v, "+", " ")
		v = strings.ReplaceAll(v, "%2C", ",")
		out[k] = v
	}
	return out
}

func TestWire_TargetServerGroupSelectors(t *testing.T) {
	s, h := recordGate(t, `{}`)
	_, handler := tools.NewGetTargetServerGroup(newGate(t, h))
	base := map[string]any{"application": "app", "account": "prod", "cluster_name": "app-main", "cloud_provider": "aws", "scope": "us-east-1"}
	for _, target := range []string{"current", "previous", "fail", "newest", "ancestor"} {
		base["target"] = target
		if res := call(t, handler, base); res.IsError {
			t.Errorf("target %s rejected: %s", target, text(res))
		}
		if !strings.HasSuffix(s.path, "/serverGroups/target/"+target) {
			t.Errorf("target %s: path %s", target, s.path)
		}
	}
	base["target"] = "smallest"
	before := s.n
	if res := call(t, handler, base); !res.IsError || s.n != before {
		t.Error("target smallest, which Clouddriver rejects, was sent")
	}
}

func TestWire_ImageTagsDefaultsToDockerRegistry(t *testing.T) {
	s, h := recordGate(t, `["v1"]`)
	_, handler := tools.NewGetImageTags(newGate(t, h))
	call(t, handler, map[string]any{"account": "dockerhub", "repository": "library/nginx"})
	if q := parseQuery(t, s.query); q["provider"] != "dockerRegistry" {
		t.Errorf("provider = %q, want dockerRegistry (query %s)", q["provider"], s.query)
	}
}

func TestWire_GetFirewallUsesFirewallsWithProvider(t *testing.T) {
	s, h := recordGate(t, `{}`)
	_, handler := tools.NewGetFirewall(newGate(t, h))
	call(t, handler, map[string]any{"account": "gcp-prod", "region": "global", "name": "allow-web", "provider": "gce"})
	if s.path != "/firewalls/gcp-prod/global/allow-web" || s.query != "provider=gce" {
		t.Errorf("got %s?%s", s.path, s.query)
	}
}

func TestWire_PipelineNameWithSpaces(t *testing.T) {
	s, h := recordGate(t, `{}`)
	_, handler := tools.NewGetPipeline(newGate(t, h))
	call(t, handler, map[string]any{"application": "app", "pipeline_name": "Deploy to Prod"})
	if s.path != "/applications/app/pipelineConfigs/Deploy%20to%20Prod" {
		t.Errorf("path = %s", s.path)
	}
}

func TestWire_CancelForce(t *testing.T) {
	s, h := recordGate(t, ``)
	_, handler := tools.NewCancelExecution(newGate(t, h))
	call(t, handler, map[string]any{"execution_id": "e1", "force": true})
	if s.method != "PUT" || s.path != "/pipelines/e1/cancel" || s.query != "force=true" {
		t.Errorf("got %s %s?%s", s.method, s.path, s.query)
	}
}

func TestWire_TriggerPipelineParameters(t *testing.T) {
	s, h := recordGate(t, `{"ref":"/pipelines/01JEXEC"}`)
	_, handler := tools.NewTriggerPipeline(newGate(t, h))
	for _, params := range []any{
		map[string]any{"tag": "v1.2.3"},
		`{"tag":"v1.2.3"}`,
	} {
		res := call(t, handler, map[string]any{"application": "app", "pipeline_name": "Deploy to Prod", "parameters": params})
		if res.IsError {
			t.Fatalf("tool error: %s", text(res))
		}
		if s.path != "/pipelines/v2/app/Deploy%20to%20Prod" {
			t.Errorf("path = %s", s.path)
		}
		var body map[string]any
		_ = json.Unmarshal([]byte(s.body), &body)
		if p, _ := body["parameters"].(map[string]any); p["tag"] != "v1.2.3" || body["type"] != "manual" {
			t.Errorf("body = %s", s.body)
		}
		b, _ := json.Marshal(res.StructuredContent)
		if !strings.Contains(string(b), `"executionId":"01JEXEC"`) {
			t.Errorf("structuredContent = %s, want executionId", b)
		}
	}
	if res := call(t, handler, map[string]any{"application": "app", "pipeline_name": "p", "parameters": "[1,2]"}); !res.IsError {
		t.Error("a JSON array was accepted as parameters")
	}
}

const executionFixture = `{
  "id": "01JEXEC", "name": "Deploy to Prod", "application": "app", "status": "RUNNING",
  "startTime": 1000, "trigger": {"type": "manual", "user": "sergio", "parameters": {"tag": "v1"}},
  "stages": [
    {"id": "s1", "refId": "1", "name": "Bake", "type": "bake", "status": "SUCCEEDED", "startTime": 1000, "endTime": 4000, "context": {"huge": "` + "x" + `"}},
    {"id": "s2", "refId": "2", "name": "Canary", "type": "deploy", "status": "TERMINAL", "startTime": 4000, "endTime": 5000,
     "context": {"exception": {"details": {"errors": ["Insufficient capacity", "retry later"]}}}},
    {"id": "s3", "refId": "3", "name": "Approve prod", "type": "manualJudgment", "status": "RUNNING", "startTime": 5000, "context": {}},
    {"id": "s4", "refId": "4", "name": "Soak", "type": "wait", "status": "RUNNING", "startTime": 5000, "context": {}}
  ]
}`

func TestGetExecution_Summary(t *testing.T) {
	_, h := recordGate(t, executionFixture)
	_, handler := tools.NewGetExecution(newGate(t, h))

	res := call(t, handler, map[string]any{"execution_id": "01JEXEC"})
	sum, ok := res.StructuredContent.(tools.ExecutionSummary)
	if !ok {
		t.Fatalf("structuredContent is %T, want ExecutionSummary", res.StructuredContent)
	}
	if sum.Status != "RUNNING" || sum.TriggeredBy != "sergio" || len(sum.Stages) != 4 || len(sum.Running) != 2 {
		t.Errorf("summary = %+v", sum)
	}
	if sum.Error != "Canary: Insufficient capacity; retry later" {
		t.Errorf("error = %q", sum.Error)
	}
	if len(sum.Judgments) != 1 || sum.Judgments[0].ID != "s3" {
		t.Errorf("awaitingJudgment = %+v", sum.Judgments)
	}
	if sum.Stages[0].DurationMs != 3000 {
		t.Errorf("Bake duration = %d", sum.Stages[0].DurationMs)
	}
	if strings.Contains(text(res), "huge") {
		t.Error("summary text carries stage context")
	}

	full := call(t, handler, map[string]any{"execution_id": "01JEXEC", "summary": false})
	if !strings.Contains(text(full), `"huge"`) {
		t.Error("summary=false does not return the full execution")
	}
}

func TestJudgeStage_OnlyRunningManualJudgment(t *testing.T) {
	var patches []string
	gate := newGate(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			b, _ := io.ReadAll(r.Body)
			patches = append(patches, r.URL.Path+" "+string(b))
			w.Write([]byte(`{}`))
			return
		}
		w.Write([]byte(executionFixture))
	})
	_, handler := tools.NewJudgeStage(gate)

	for _, tc := range []map[string]any{
		{"execution_id": "01JEXEC", "stage_id": "s2", "decision": "continue"},   // deploy stage
		{"execution_id": "01JEXEC", "stage_id": "s4", "decision": "continue"},   // running, not a judgment
		{"execution_id": "01JEXEC", "stage_id": "nope", "decision": "continue"}, // unknown stage
		{"execution_id": "01JEXEC", "stage_id": "s3", "decision": "maybe"},      // bad decision
	} {
		if res := call(t, handler, tc); !res.IsError {
			t.Errorf("%v: accepted", tc)
		}
	}
	if len(patches) != 0 {
		t.Fatalf("PATCH sent for a refused judgment: %v", patches)
	}

	res := call(t, handler, map[string]any{"execution_id": "01JEXEC", "stage_id": "s3", "decision": "stop", "input": "rollback"})
	if res.IsError {
		t.Fatalf("tool error: %s", text(res))
	}
	if len(patches) != 1 || patches[0] != `/pipelines/01JEXEC/stages/s3 {"judgmentInput":"rollback","judgmentStatus":"stop"}` {
		t.Errorf("patches = %v", patches)
	}
}

func TestEvaluateExpression_Guard(t *testing.T) {
	s, h := recordGate(t, `{"result":"ok"}`)
	_, handler := tools.NewEvaluateExpression(newGate(t, h))
	for _, expr := range []string{
		"${new\tjava.io.File('/etc').list()}",
		"${new\tcom.example.Exploit()}",
		"${NEW java.util.Date()}",
		"${new\njava.net.URL('http://169.254.169.254/').openStream()}",
		"${T (java.lang.Runtime).getRuntime()}",
		"${''.getClass().forName('x')}",
		"${java.nio.file.Files}",
	} {
		before := s.n
		if res := call(t, handler, map[string]any{"execution_id": "e", "expression": expr}); !res.IsError || s.n != before {
			t.Errorf("%q reached Gate", expr)
		}
	}
	for _, expr := range []string{"${parameters.runtimeVersion}", "${#stage('Deploy').context.threadCount}", "${trigger.parameters.tag}"} {
		if res := call(t, handler, map[string]any{"execution_id": "e", "expression": expr}); res.IsError {
			t.Errorf("%q rejected: %s", expr, text(res))
		}
	}
}

func TestStructuredContent_ListAndRecord(t *testing.T) {
	_, h := recordGate(t, `[{"name":"a"},{"name":"b"}]`)
	_, handler := tools.NewListApplications(newGate(t, h))
	res := call(t, handler, nil)
	m, _ := res.StructuredContent.(map[string]any)
	if m["count"] != 2 {
		t.Errorf("list structuredContent = %v", res.StructuredContent)
	}
	if text(res) == "" {
		t.Error("text content dropped")
	}

	_, h = recordGate(t, `{"name":"myapp","attributes":{"email":"x"}}`)
	_, handler = tools.NewGetApplication(newGate(t, h))
	res = call(t, handler, map[string]any{"application": "myapp"})
	m, _ = res.StructuredContent.(map[string]any)
	if m["name"] != "myapp" {
		t.Errorf("record structuredContent = %v", res.StructuredContent)
	}
}
