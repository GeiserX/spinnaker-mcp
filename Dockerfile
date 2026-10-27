# ───────────────────────────────────────────────
# Stage 1 – build the Go binary
# ───────────────────────────────────────────────
# Pinned by digest; dependabot's docker ecosystem proposes updates. 1.27.3 matches
# the go.mod toolchain, which rejects the compromised eMudhra intermediates.
FROM golang:1.27.3@sha256:9048d5e7751706200eb1719c7dae0436950d362f29a1da91fa05e5c63867c60f AS builder
LABEL maintainer="9169332+GeiserX@users.noreply.github.com"

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# goreleaser passes the release version, so the image reports it instead of "dev".
ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/geiserx/spinnaker-mcp/version.Version=${VERSION} -X github.com/geiserx/spinnaker-mcp/version.Commit=${COMMIT}" \
      -o /out/spinnaker-mcp ./cmd/server

# ───────────────────────────────────────────────
# Stage 2 – tiny runtime image
# ───────────────────────────────────────────────
FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
# The base image can lag the CA bundle; take the current one, which drops
# the compromised eMudhra (emSign) intermediates.
RUN apk upgrade --no-cache ca-certificates-bundle
LABEL io.modelcontextprotocol.server.name="io.github.GeiserX/spinnaker-mcp"

RUN addgroup -S mcp && adduser -S mcp -G mcp
COPY --from=builder /out/spinnaker-mcp /usr/local/bin/spinnaker-mcp

ENV GATE_URL=http://spin-gate:8084
ENV TRANSPORT=stdio

USER mcp

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD if [ "$TRANSPORT" = "http" ] || [ "$TRANSPORT" = "HTTP" ]; then wget -qO- "http://127.0.0.1:${MCP_PORT:-8085}/healthz" || exit 1; else exit 0; fi

ENTRYPOINT ["/usr/local/bin/spinnaker-mcp"]
