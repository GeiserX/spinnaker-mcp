package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/internal/toolsets"
	"github.com/mark3labs/mcp-go/server"
)

// stdioPeer drives the real MCP server over the stdio transport, one JSON-RPC line at a time.
type stdioPeer struct {
	t   *testing.T
	in  *io.PipeWriter
	out *bufio.Reader
}

func newStdioPeer(t *testing.T, toolsetSpec string) (*stdioPeer, int) {
	t.Helper()
	gate, err := client.NewGate(client.GateOptions{BaseURL: "http://127.0.0.1:1"})
	if err != nil {
		t.Fatalf("NewGate: %v", err)
	}
	selected, err := toolsets.Resolve(toolsetSpec, toolsets.BuildTools(gate))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	s := newMCPServer(gate, selected)

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.NewStdioServer(s).Listen(ctx, inR, outW)
		outW.Close()
	}()
	t.Cleanup(func() {
		cancel()
		inW.Close()
		<-done
	})
	return &stdioPeer{t: t, in: inW, out: bufio.NewReader(outR)}, len(selected)
}

func (p *stdioPeer) send(msg string) {
	p.t.Helper()
	if _, err := io.WriteString(p.in, msg+"\n"); err != nil {
		p.t.Fatalf("write: %v", err)
	}
}

// call sends a request and returns its decoded response.
func (p *stdioPeer) call(msg string) map[string]any {
	p.t.Helper()
	p.send(msg)
	type res struct {
		line string
		err  error
	}
	ch := make(chan res, 1)
	go func() {
		l, err := p.out.ReadString('\n')
		ch <- res{l, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			p.t.Fatalf("read: %v", r.err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(r.line), &m); err != nil {
			p.t.Fatalf("bad JSON-RPC line %q: %v", r.line, err)
		}
		return m
	case <-time.After(10 * time.Second):
		p.t.Fatalf("no response to %s", msg)
	}
	return nil
}

const modernMeta = `"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"protocol-test","version":"0"}}`

func result(t *testing.T, resp map[string]any) map[string]any {
	t.Helper()
	r, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("no result in %v", resp)
	}
	return r
}

// A 2026-07-28 client never sends initialize: it calls server/discover and then sends
// every request with its version and capabilities in _meta.
func TestProtocol_ModernStdio(t *testing.T) {
	p, nTools := newStdioPeer(t, "")

	disc := result(t, p.call(`{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{`+modernMeta+`}}`))
	versions, _ := disc["supportedVersions"].([]any)
	var got []string
	for _, v := range versions {
		got = append(got, v.(string))
	}
	for _, want := range []string{"2026-07-28", "2025-11-25", "2024-11-05"} {
		if !slices.Contains(got, want) {
			t.Errorf("server/discover supportedVersions = %v, missing %s", got, want)
		}
	}
	if disc["instructions"] == "" || disc["instructions"] == nil {
		t.Error("server/discover carries no instructions")
	}

	list := result(t, p.call(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{`+modernMeta+`}}`))
	if list["resultType"] != "complete" {
		t.Errorf("tools/list resultType = %v, want complete", list["resultType"])
	}
	if ttl, _ := list["ttlMs"].(float64); ttl != listCacheTTLMs {
		t.Errorf("tools/list ttlMs = %v, want %d", list["ttlMs"], listCacheTTLMs)
	}
	if list["cacheScope"] != "public" {
		t.Errorf("tools/list cacheScope = %v, want public", list["cacheScope"])
	}
	tools, _ := list["tools"].([]any)
	if len(tools) != nTools {
		t.Errorf("tools/list returned %d tools, want %d", len(tools), nTools)
	}

	// ping was removed in 2026-07-28; a modern request for it is an unknown method.
	ping := p.call(`{"jsonrpc":"2.0","id":3,"method":"ping","params":{` + modernMeta + `}}`)
	errObj, ok := ping["error"].(map[string]any)
	if !ok {
		t.Fatalf("modern ping answered %v, want an error", ping)
	}
	if code, _ := errObj["code"].(float64); code != -32601 {
		t.Errorf("modern ping error code = %v, want -32601", errObj["code"])
	}
}

// Claude Code, Claude Desktop, Cursor and VS Code still open with initialize.
func TestProtocol_LegacyStdio(t *testing.T) {
	p, nTools := newStdioPeer(t, "all")

	init := result(t, p.call(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"legacy-test","version":"0"}}}`))
	if init["protocolVersion"] != "2025-11-25" {
		t.Errorf("initialize protocolVersion = %v, want 2025-11-25", init["protocolVersion"])
	}
	info, _ := init["serverInfo"].(map[string]any)
	if info["title"] != "Spinnaker" || info["websiteUrl"] == nil {
		t.Errorf("serverInfo = %v, want title and websiteUrl", info)
	}
	p.send(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)

	list := result(t, p.call(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	tools, _ := list["tools"].([]any)
	if len(tools) != nTools {
		t.Errorf("legacy tools/list returned %d tools, want %d", len(tools), nTools)
	}
	// Tools come back sorted by name, so the list is stable for prompt caches.
	var names []string
	for _, tl := range tools {
		names = append(names, tl.(map[string]any)["name"].(string))
	}
	if !slices.IsSorted(names) {
		t.Errorf("tools/list order is not deterministic: %v", names)
	}
	// The cache fields are 2026-07-28 only; a legacy client gets the old shape.
	if _, ok := list["ttlMs"]; ok {
		t.Error("legacy tools/list carries ttlMs")
	}

	pong := p.call(`{"jsonrpc":"2.0","id":3,"method":"ping"}`)
	if _, ok := pong["result"]; !ok {
		t.Errorf("legacy ping = %v, want a result", pong)
	}
}
