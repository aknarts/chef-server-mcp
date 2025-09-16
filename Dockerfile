# Multi-stage build for chef-mcp
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Inject version metadata if provided
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/aknarts/chef-server-mcp/internal/version.Version=${VERSION}" -o /out/chef-mcp ./cmd/chef-mcp

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/chef-mcp /chef-mcp
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/chef-mcp"]
