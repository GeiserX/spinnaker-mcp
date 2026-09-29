# Getting started

Needs a reachable Spinnaker Gate URL and, for the npm package, Node 18 or newer.

## npm (stdio transport)

```sh
npx spinnaker-mcp
```

Or install globally:

```sh
npm install -g spinnaker-mcp
spinnaker-mcp
```

This downloads the pre-built Go binary for your platform and runs it with stdio transport.

## Docker

```sh
docker run -i --rm -e GATE_URL=http://spin-gate:8084 -e TRANSPORT=stdio drumsergio/spinnaker-mcp:v0.3.3
```

Image tags carry the `v` prefix (`v0.3.3`). `-i` keeps stdin open, which the stdio transport needs.

## Local build

```sh
git clone https://github.com/GeiserX/spinnaker-mcp
cd spinnaker-mcp

cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

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

See [Configuration](configuration.md) for the environment variables.
