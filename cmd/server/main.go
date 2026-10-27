package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/geiserx/spinnaker-mcp/config"
	"github.com/geiserx/spinnaker-mcp/internal/toolsets"
	"github.com/geiserx/spinnaker-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var toolsetsFlag string

	if len(os.Args) > 1 && os.Args[1] == "login" {
		os.Exit(runLogin(os.Args[2:], os.Stdin, os.Stderr))
	}

	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Println("spinnaker-mcp " + version.String())
			return
		case arg == "--help" || arg == "-h":
			printHelp()
			return
		case strings.HasPrefix(arg, "--toolsets="):
			toolsetsFlag = strings.TrimPrefix(arg, "--toolsets=")
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\nRun 'spinnaker-mcp --help' for usage.\n", arg)
				os.Exit(1)
			}
		}
	}

	log.Printf("Spinnaker MCP %s starting…", version.String())

	cfg := config.LoadGateConfig()
	gate, err := newGate(cfg)
	if err != nil {
		log.Fatalf("Failed to create Gate client: %v", err)
	}
	if cfg.Token == "" && cfg.User == "" && cfg.Cookie != "" {
		log.Println("Gate auth: session cookie")
	}

	// Resolve toolsets
	toolsetsFlag = toolsetsSpec(toolsetsFlag)

	allTools := toolsets.BuildTools(gate)
	selectedTools, err := toolsets.Resolve(toolsetsFlag, allTools)
	if err != nil {
		log.Fatalf("Invalid toolsets: %v", err)
	}

	s := newMCPServer(gate, selectedTools)
	log.Printf("Registered %d tools (toolsets: %s)", len(selectedTools), toolsetsLabel(toolsetsFlag))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	transport := strings.ToLower(config.Getenv("TRANSPORT"))
	if transport == "stdio" {
		stdioSrv := server.NewStdioServer(s)
		log.Println("Spinnaker MCP running on stdio")
		if err := stdioSrv.Listen(ctx, os.Stdin, os.Stdout); err != nil && ctx.Err() == nil {
			log.Fatalf("stdio server error: %v", err)
		}
		return
	}

	httpCfg, err := loadHTTPConfig()
	if err != nil {
		log.Fatal(err)
	}
	handler, err := newHTTPHandler(s, gate, httpCfg)
	if err != nil {
		log.Fatalf("HTTP setup: %v", err)
	}
	addr := net.JoinHostPort(httpCfg.bindAddr, httpCfg.port)
	auth := "no bearer token (loopback only)"
	if httpCfg.token != "" {
		auth = "bearer token required"
	}
	log.Printf("Spinnaker MCP listening on %s, %s (health: /healthz, /readyz)", addr, auth)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// Longer than the 30s Gate client timeout, so a slow Gate call still gets its answer out.
		WriteTimeout:   90 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 64 << 10,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case <-ctx.Done():
		log.Println("Shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}
}

func newGate(cfg config.GateConfig) (*client.GateClient, error) {
	return client.NewGate(client.GateOptions{
		BaseURL:  cfg.BaseURL,
		Token:    cfg.Token,
		User:     cfg.User,
		Pass:     cfg.Pass,
		CertFile: cfg.CertFile,
		KeyFile:  cfg.KeyFile,
		Insecure: cfg.Insecure,
		Cookie:   cfg.Cookie,
	})
}

// toolsetsSpec returns the --toolsets flag if given, else TOOLSETS from the environment.
func toolsetsSpec(flag string) string {
	if flag != "" {
		return flag
	}
	return config.Getenv("TOOLSETS")
}

func toolsetsLabel(raw string) string {
	if raw == "" {
		return toolsets.MetaReadonly + " (default)"
	}
	return raw
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": version.Version,
	})
}

func readyzHandler(gate *client.GateClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := gate.Ping(ctx)
		reachable := err == nil

		status := "ready"
		code := http.StatusOK
		if !reachable {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}

		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"status":         status,
			"gate_reachable": reachable,
		})
	}
}

func printHelp() {
	fmt.Println("spinnaker-mcp " + version.String())
	fmt.Println()
	fmt.Println("Usage: spinnaker-mcp [flags]")
	fmt.Println("       spinnaker-mcp login [--gate=URL] [--cookie=VALUE] [--cookie-file=PATH] [--no-browser]")
	fmt.Println("                              Sign in to a Gate behind SSO and store its session cookie")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v          Print version and exit")
	fmt.Println("  --help, -h             Show this help message")
	fmt.Println("  --toolsets=GROUPS      Comma-separated tool groups to enable (default: all)")
	fmt.Println()
	fmt.Println("Toolset groups:")
	fmt.Println("  applications           Application listing and details (2 tools)")
	fmt.Println("  pipelines              Pipeline CRUD, trigger, history (7 tools)")
	fmt.Println("  executions             Execution lifecycle and search (8 tools)")
	fmt.Println("  strategies             Deployment strategy management (3 tools)")
	fmt.Println("  infrastructure         Server groups, clusters, LBs, firewalls, images, networks, accounts (16 tools)")
	fmt.Println("  tasks                  Task status (1 tool)")
	fmt.Println()
	fmt.Println("Meta-groups:")
	fmt.Println("  all                    All tool groups (default)")
	fmt.Println("  readonly               Only read-only tools (27 tools)")
	fmt.Println("  mutating               Only mutating and destructive tools (10 tools)")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  GATE_URL               Spinnaker Gate API URL (default: http://localhost:8084)")
	fmt.Println("  GATE_TOKEN             Bearer token for Gate authentication")
	fmt.Println("  GATE_USER              Basic auth username (alternative to token)")
	fmt.Println("  GATE_PASS              Basic auth password")
	fmt.Println("  GATE_CERT_FILE         Path to x509 client certificate (PEM)")
	fmt.Println("  GATE_KEY_FILE          Path to x509 client key (PEM)")
	fmt.Println("  GATE_INSECURE          Skip TLS certificate verification (default: false)")
	fmt.Println("  GATE_COOKIE            Session cookie for a Gate behind SSO (SESSION=... or the bare value)")
	fmt.Println("  GATE_COOKIE_FILE       File holding that cookie (default: $XDG_CONFIG_HOME/spinnaker-mcp/cookies/<host>)")
	fmt.Println("  TRANSPORT              'stdio' for stdio transport (default: HTTP)")
	fmt.Println("  MCP_PORT               HTTP listen port (default: 8085)")
	fmt.Println("  MCP_BIND_ADDR          HTTP bind address (default: 127.0.0.1)")
	fmt.Println("  TOOLSETS               Same as --toolsets flag (flag takes precedence)")
}
