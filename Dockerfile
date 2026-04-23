# ───────────────────────────────────────────────
# Stage 1 – build the Go binary
# ───────────────────────────────────────────────
FROM golang:1.25 AS builder
LABEL maintainer="9169332+GeiserX@users.noreply.github.com"

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/spinnaker-mcp ./cmd/server

# ───────────────────────────────────────────────
# Stage 2 – tiny runtime image
# ───────────────────────────────────────────────
FROM alpine:3.23
LABEL io.modelcontextprotocol.server.name="io.github.GeiserX/spinnaker-mcp"

RUN addgroup -S mcp && adduser -S mcp -G mcp
COPY --from=builder /out/spinnaker-mcp /usr/local/bin/spinnaker-mcp

ENV GATE_URL=http://spin-gate:8084
ENV TRANSPORT=stdio

USER mcp
ENTRYPOINT ["/usr/local/bin/spinnaker-mcp"]
