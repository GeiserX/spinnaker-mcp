<p align="center">
  <img src="docs/images/banner.svg" alt="Spinnaker MCP banner" width="900"/>
</p>

<h1 align="center">Spinnaker-MCP</h1>

<p align="center">
  <a href="https://www.npmjs.com/package/spinnaker-mcp"><img src="https://img.shields.io/npm/v/spinnaker-mcp?style=flat-square&logo=npm" alt="npm"/></a>
  <a href="https://github.com/GeiserX/spinnaker-mcp/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/GeiserX/spinnaker-mcp/ci.yml?style=flat-square&logo=github&label=CI" alt="CI"/></a>
  <a href="https://codecov.io/gh/GeiserX/spinnaker-mcp"><img src="https://img.shields.io/codecov/c/github/GeiserX/spinnaker-mcp?style=flat-square&logo=codecov&label=Coverage" alt="Coverage"/></a>
  <img src="https://img.shields.io/badge/Go-1.25-blue?style=flat-square&logo=go&logoColor=white" alt="Go"/>
  <a href="https://hub.docker.com/r/drumsergio/spinnaker-mcp"><img src="https://img.shields.io/docker/pulls/drumsergio/spinnaker-mcp?style=flat-square&logo=docker" alt="Docker Pulls"/></a>
  <a href="https://github.com/GeiserX/spinnaker-mcp/stargazers"><img src="https://img.shields.io/github/stars/GeiserX/spinnaker-mcp?style=flat-square&logo=github" alt="GitHub Stars"/></a>
  <a href="https://github.com/GeiserX/spinnaker-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/GeiserX/spinnaker-mcp?style=flat-square" alt="License"/></a>
</p>

<p align="center"><strong>A bridge that exposes any Spinnaker instance as an MCP v1 server via the Gate API, written in Go.</strong></p>

---

## What you get

| Category | Tool | Description |
|----------|------|-------------|
| **Applications** | `list_applications` | List all Spinnaker applications |
| | `get_application` | Get detailed application info (accounts, clusters, attributes) |
| **Pipelines** | `list_pipelines` | List pipeline configurations for an application |
| | `get_pipeline` | Get a specific pipeline's full configuration |
| | `trigger_pipeline` | Trigger a pipeline with optional parameters |
| **Executions** | `list_executions` | List recent executions, filterable by status |
| | `get_execution` | Get full execution details (stages, outputs, timing) |
| | `cancel_execution` | Cancel a running execution with optional reason |
| | `pause_execution` | Pause a running execution at the current stage |
| | `resume_execution` | Resume a paused execution |
| **Infrastructure** | `list_server_groups` | List server groups (deployment targets) with instance counts |
| | `list_load_balancers` | List load balancers across all accounts and regions |
| **Tasks** | `get_task` | Get orchestration task status (deploy, resize, rollback) |

Everything is exposed over JSON-RPC. LLMs and agents can: `initialize` -> `listTools` -> `callTool` and interact with your Spinnaker deployments.

---

## Quick-start

### npm (stdio transport)

```sh
npx spinnaker-mcp
```

Or install globally:

```sh
npm install -g spinnaker-mcp
spinnaker-mcp
```

This downloads the pre-built Go binary for your platform and runs it with stdio transport.

### Docker

```sh
docker run --rm -e GATE_URL=http://spin-gate:8084 -e TRANSPORT=stdio drumsergio/spinnaker-mcp:0.1.0
```

### Local build

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
cd spinnaker-mcp

cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GATE_URL` | `http://localhost:8084` | Spinnaker Gate API endpoint (without trailing /) |
| `GATE_TOKEN` | _(empty)_ | Bearer token for authentication |
| `GATE_USER` | _(empty)_ | Basic auth username (alternative to token) |
| `GATE_PASS` | _(empty)_ | Basic auth password |
| `GATE_CERT_FILE` | _(empty)_ | Path to x509 client certificate (PEM) |
| `GATE_KEY_FILE` | _(empty)_ | Path to x509 client key (PEM) |
| `GATE_INSECURE` | `false` | Skip TLS certificate verification |
| `TRANSPORT` | _(empty = HTTP)_ | Set to `stdio` for stdio transport |

**Authentication priority**: Bearer token > Basic auth > x509 client cert > No auth.

Put them in a `.env` file (from `.env.example`) or set them in the environment.

## Claude Code / Claude Desktop configuration

```json
{
  "mcpServers": {
    "spinnaker": {
      "command": "npx",
      "args": ["-y", "spinnaker-mcp"],
      "env": {
        "GATE_URL": "https://spin-gate.example.com",
        "GATE_TOKEN": "your-token-here"
      }
    }
  }
}
```

## Testing

```sh
go test -v -race ./...
```

Tested with [Inspector](https://modelcontextprotocol.io/docs/tools/inspector). Before making a PR, make sure this MCP server behaves well via that tool.

## Credits

[Spinnaker](https://spinnaker.io/) -- open-source continuous delivery platform

[MCP-GO](https://github.com/mark3labs/mcp-go) -- Go MCP implementation

[GoReleaser](https://goreleaser.com/) -- painless multi-arch releases

## Maintainers

[@GeiserX](https://github.com/GeiserX).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/GeiserX/spinnaker-mcp/issues/new) or submit PRs.

Spinnaker-MCP follows the [Contributor Covenant](http://contributor-covenant.org/version/2/1/) Code of Conduct.

## Other MCP Servers by GeiserX

- [genieacs-mcp](https://github.com/GeiserX/genieacs-mcp) -- TR-069 device management
- [cashpilot-mcp](https://github.com/GeiserX/cashpilot-mcp) -- Passive income monitoring
- [duplicacy-mcp](https://github.com/GeiserX/duplicacy-mcp) -- Backup health monitoring
- [lynxprompt-mcp](https://github.com/GeiserX/lynxprompt-mcp) -- AI configuration blueprints
- [pumperly-mcp](https://github.com/GeiserX/pumperly-mcp) -- Fuel and EV charging prices
- [telegram-archive-mcp](https://github.com/GeiserX/telegram-archive-mcp) -- Telegram message archive
