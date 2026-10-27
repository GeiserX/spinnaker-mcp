package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
	"github.com/mark3labs/mcp-go/server"
)

// maxRequestBody caps one POST to /mcp. The largest legitimate request is a 1 MiB
// pipeline_json for save_pipeline or update_pipeline plus JSON-RPC framing.
const maxRequestBody = 4 << 20

type httpConfig struct {
	port     string
	bindAddr string
	// token, when set, must arrive as "Authorization: Bearer <token>" on every /mcp request.
	token string
	// allowedOrigins are browser origins (scheme://host[:port]) allowed to call /mcp cross-origin.
	allowedOrigins []string
}

// loadHTTPConfig reads the HTTP transport settings and refuses an unauthenticated
// listener on anything but loopback: /mcp acts with the server's Gate credential.
func loadHTTPConfig() (httpConfig, error) {
	c := httpConfig{
		port:     config.Getenv("MCP_PORT"),
		bindAddr: config.Getenv("MCP_BIND_ADDR"),
		token:    strings.TrimSpace(config.Getenv("MCP_AUTH_TOKEN")),
	}
	if c.port == "" {
		c.port = "8085"
	}
	if p, err := strconv.Atoi(c.port); err != nil || p < 1 || p > 65535 {
		return c, fmt.Errorf("invalid MCP_PORT %q: must be 1-65535", c.port)
	}
	if c.bindAddr == "" {
		c.bindAddr = "127.0.0.1"
	}
	for _, o := range strings.Split(config.Getenv("MCP_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.allowedOrigins = append(c.allowedOrigins, o)
		}
	}
	if c.token == "" && !isLoopback(c.bindAddr) {
		return c, fmt.Errorf("MCP_BIND_ADDR=%s is not a loopback address and MCP_AUTH_TOKEN is not set: "+
			"anyone who can reach the port would act with this server's Gate credential. "+
			"Set MCP_AUTH_TOKEN (for example: openssl rand -hex 32) or bind to 127.0.0.1", c.bindAddr)
	}
	return c, nil
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// newHTTPHandler mounts the stateless Streamable HTTP endpoint at /mcp behind a bearer
// check (when a token is configured), cross-origin protection and a body cap, next to the
// unauthenticated /healthz and /readyz probes, which reveal nothing about Gate.
func newHTTPHandler(s *server.MCPServer, gate *client.GateClient, c httpConfig) (http.Handler, error) {
	mcpHandler := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
		// No per-session state: every tool is global, and 2026-07-28 has no sessions.
		server.WithStateLess(true),
	)

	cop := http.NewCrossOriginProtection()
	for _, o := range c.allowedOrigins {
		if err := cop.AddTrustedOrigin(o); err != nil {
			return nil, fmt.Errorf("MCP_ALLOWED_ORIGINS: %w", err)
		}
	}

	var h http.Handler = http.MaxBytesHandler(mcpHandler, maxRequestBody)
	h = cop.Handler(h)
	if c.token != "" {
		h = requireBearer(c.token, h)
	}

	mux := http.NewServeMux()
	mux.Handle("/mcp", h)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler(gate))
	return mux, nil
}

// requireBearer accepts only "Authorization: Bearer <token>". The comparison runs on
// SHA-256 digests in constant time, so neither the token nor its length leaks by timing.
func requireBearer(token string, next http.Handler) http.Handler {
	want := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := bearerToken(r.Header.Get("Authorization"))
		sum := sha256.Sum256([]byte(got))
		if !ok || subtle.ConstantTimeCompare(sum[:], want[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="spinnaker-mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	t := strings.TrimSpace(header[len(prefix):])
	return t, t != ""
}
