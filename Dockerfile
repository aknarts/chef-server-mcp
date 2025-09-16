# Minimal image for mcp-chef (HTTP server removed)
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/aknarts/chef-server-mcp/internal/version.Version=${VERSION}" -o /out/mcp-chef ./cmd/mcp-chef

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/mcp-chef /mcp-chef
USER nonroot
# No exposed port (stdio protocol)
ENTRYPOINT ["/mcp-chef"]
