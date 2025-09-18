# Multi-stage build for mcp-chef with production and debug variants
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags "-s -w -X github.com/aknarts/chef-server-mcp/internal/version.Version=${VERSION}" -o /out/mcp-chef ./cmd/mcp-chef

# Production image (minimal, no shell)
FROM gcr.io/distroless/static:nonroot AS production
COPY --from=build /out/mcp-chef /mcp-chef
USER nonroot
ENTRYPOINT ["/mcp-chef"]

# Debug image (includes bash and common debugging tools)
FROM alpine:latest AS debug
RUN apk add --no-cache bash curl wget ca-certificates
COPY --from=build /out/mcp-chef /mcp-chef
RUN adduser -D -s /bin/bash nonroot
USER nonroot
ENTRYPOINT ["/mcp-chef"]

# Default to production image
FROM production
