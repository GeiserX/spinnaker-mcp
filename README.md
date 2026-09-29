<p align="center">
  <img src="https://raw.githubusercontent.com/GeiserX/spinnaker-mcp/main/docs/images/banner.svg" alt="Spinnaker MCP banner" width="900"/>
</p>

<h1 align="center">Spinnaker-MCP</h1>

<p align="center">
  <a href="https://www.npmjs.com/package/spinnaker-mcp"><img src="https://img.shields.io/npm/v/spinnaker-mcp?style=flat-square&logo=npm" alt="npm"/></a>
  <a href="https://github.com/GeiserX/spinnaker-mcp/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/GeiserX/spinnaker-mcp/ci.yml?style=flat-square&logo=github&label=CI" alt="CI"/></a>
  <a href="https://codecov.io/gh/GeiserX/spinnaker-mcp"><img src="https://img.shields.io/codecov/c/github/GeiserX/spinnaker-mcp?style=flat-square&logo=codecov&label=Coverage" alt="Coverage"/></a>
  <a href="https://hub.docker.com/r/drumsergio/spinnaker-mcp"><img src="https://img.shields.io/docker/pulls/drumsergio/spinnaker-mcp?style=flat-square&logo=docker" alt="Docker Pulls"/></a>
  <a href="https://github.com/GeiserX/spinnaker-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/GeiserX/spinnaker-mcp?style=flat-square" alt="License"/></a>
</p>

<p align="center"><strong>A bridge that exposes any Spinnaker instance as an MCP v1 server via the Gate API, written in Go.</strong></p>

It gives an LLM 37 tools to read and operate Spinnaker: applications, pipelines, executions, strategies and infrastructure.

## Features

- Applications and pipelines: list, read, trigger, save, update and delete pipelines, with revision history.
- Executions: list, search by trigger type, time range and status, cancel, pause, resume, restart a stage, evaluate SpEL.
- Deployment strategies: list, save and delete.
- Infrastructure: server groups, load balancers, clusters, firewalls, instances and console output, images, networks, subnets, accounts.
- Orchestration task status (deploy, resize, rollback).
- Bearer token, basic auth or x509 client certificate against Gate.
- stdio or HTTP transport; HTTP binds to `127.0.0.1` by default.
- Ships as an npm package (`npx spinnaker-mcp`), a Docker image and multi-arch Go binaries.

## Quick start

```sh
npx spinnaker-mcp
```

Set `GATE_URL` (and a credential such as `GATE_TOKEN`) first. Docker, local builds and a Claude Code / Claude Desktop config are in [Installation](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/installation.md).

## Documentation

- [Installation](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/installation.md): npm, Docker, local build, Claude Code / Claude Desktop
- [Configuration](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/configuration.md): environment variables and authentication order
- [Tools](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/usage.md): all 37 tools
- [Development](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/development.md): testing, contributing, credits
- [Related projects and listings](https://github.com/GeiserX/spinnaker-mcp/blob/main/docs/related.md)

## License

[GPL-3.0](https://github.com/GeiserX/spinnaker-mcp/blob/main/LICENSE)
