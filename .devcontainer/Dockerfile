# syntax=docker/dockerfile:1
# Trellis DevContainer — Go 1.26 + Node.js 22
# Based on the official Go devcontainer with Go version override.

FROM mcr.microsoft.com/devcontainers/go:1-bookworm

# Override Go version to match go.mod requirement
ARG GO_VERSION=1.26.0
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

# Verify installation
RUN go version
