# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-22
**Commit:** 6937a2e+
**Branch:** main

## OVERVIEW

MCP (Model Context Protocol) server providing AI assistants read-only access to Chef Server resources. Go binary, stdio JSON-RPC transport, no HTTP. Wraps `go-chef` client with multi-organization support. Read-only enforced at three layers: MCP annotations, interface narrowing, HTTP transport blocking.

## STRUCTURE

```
.
├── cmd/mcp-chef/main.go            # Entry point: config init, 18 MCP tool registrations, signal handling
├── internal/
│   ├── chefapi/chefapi.go          # Chef API wrapper — ChefReader interface, per-org client caching, 20 public methods
│   ├── chefapi/chefapi_test.go     # Interface compliance, constructor, concurrency tests
│   ├── config/config.go            # Env var loader — credentials, org aliases, backward-compat URL parsing
│   ├── config/config_test.go       # Env loading, alias parsing, org resolution tests
│   ├── safeclient/roundtripper.go  # Read-only HTTP transport — blocks POST/PUT/DELETE/PATCH
│   ├── safeclient/roundtripper_test.go # HTTP method allow/block tests
│   └── version/version.go          # Build-time version injection via ldflags
├── scripts/mcp-chef-wrapper.sh     # Dev wrapper: auto-rebuild + exec for MCP stdio lifecycle
├── Dockerfile                      # Multi-stage: golang:1.25 → distroless (prod) / alpine (debug)
├── Makefile                        # build, test, docker, smoke targets
└── .github/
    ├── workflows/ci.yml            # CI: test on push; multi-arch Docker + release on tags
    └── copilot-instructions.md     # Legacy architecture notes (partially stale)
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add new Chef resource tool | `cmd/mcp-chef/main.go` | Add Input/Output structs + `mcp.AddTool()` with `annotations`; add method to `ChefReader` interface |
| Add new Chef API method | `internal/chefapi/chefapi.go` | Add to `ChefReader` interface AND `ChefAPI` struct; use `getClientForOrg()` |
| Read-only enforcement | `internal/safeclient/roundtripper.go` | HTTP transport layer; blocks non-GET methods at runtime |
| Change env var handling | `internal/config/config.go` | `LoadFromEnv()` + `ResolveOrganization()` |
| Docker build issues | `Dockerfile` | Multi-stage; CGO_ENABLED=0 for static binary |
| CI/release pipeline | `.github/workflows/ci.yml` | Multi-arch: amd64, arm64, arm/v7 |
| Local dev workflow | `scripts/mcp-chef-wrapper.sh` | Set `NO_REBUILD=1` to skip recompilation |
| Version bump | `internal/version/version.go` | Never hardcode — injected via `-ldflags` at build |

## CONVENTIONS

- **All tools are read-only** — enforced via `ReadOnlyTransport` (HTTP layer), `ChefReader` interface (compile-time), and `ToolAnnotations.ReadOnlyHint` (MCP protocol)
- **Every tool has `Annotations: readOnlyAnnotations()`** — new tools MUST include this
- **Every tool accepts optional `organization` param** — resolved via: explicit arg → alias → default → error
- **Organization resolution order**: per-request param → `CHEF_ORG_ALIASES` lookup → `CHEF_DEFAULT_ORG` → fail
- **Backward compat**: `CHEF_SERVER_URL` containing `/organizations/<org>` auto-extracts org and trims URL
- **No global state** — config passed explicitly, ChefAPI clients cached per-org with `sync.RWMutex`
- **Logging to stderr only** — stdout reserved for MCP stdio transport
- **Build-time version**: `make build` injects via `-ldflags -X .../version.Version=$(git describe)`
- **New read methods**: must be added to BOTH `ChefReader` interface AND `ChefAPI` struct (compile-time enforced)

## ANTI-PATTERNS (THIS PROJECT)

- **NEVER log private key contents** — security rule
- **NEVER run without all 3 credentials**: `CHEF_USER`, `CHEF_KEY_PATH`, `CHEF_SERVER_URL` — startup fatals
- **NEVER hardcode version strings** — always ldflags injection
- **NEVER add write methods to ChefAPI** — `ReadOnlyTransport` blocks POST/PUT/DELETE at HTTP layer; `ChefReader` interface limits compile-time surface
- **DO NOT write to Chef Server** — this is a read-only MCP server enforced at 3 layers
- **DO NOT use local filesystem paths for `CHEF_KEY_PATH` in Docker** — must be container-relative (e.g., `/chef/user.pem`)
- **DO NOT hit real Chef servers in unit tests** — use mocked `ChefAPI` with dependency injection
- **DO NOT register tools without `readOnlyAnnotations()`** — all tools must declare read-only via MCP protocol

## UNIQUE STYLES

- **Monolithic main.go** (~600 lines): all Input/Output type defs + all 18 tool registrations live in one file
- **Defense-in-depth read-only**: `safeclient.ReadOnlyTransport` (runtime) + `ChefReader` interface (compile-time) + MCP `ToolAnnotations` (protocol)
- **Org alias parsing** accepts both JSON (`{"qa":"qa1"}`) and simple format (`qa=qa1,prod=production`)
- **`getClientForOrg()` double-checked locking** — RWMutex with fast RLock path, slow Lock+create path

## COMMANDS

```bash
make build        # Build dist/mcp-chef binary
make test         # go test -race ./...
make vet          # go vet static analysis
make ci           # tidy + lint + test + build (full local CI)
make docker-build # Docker image (single-arch)
make mcp-smoke    # Smoke test: send JSON-RPC over stdio
make clean        # Remove dist/ and coverage.out
```

## NOTES

- `copilot-instructions.md` references `internal/server` and knife fallback — both removed. Treat as aspirational/historical
- `internal/knife` package was removed — dead code from knife fallback era
- Go module requires 1.23+ but CI and Dockerfile use 1.25.x
- MCP SDK: `github.com/modelcontextprotocol/go-sdk v0.5.0` — tool registration via `mcp.AddTool(server, &mcp.Tool{...}, handlerFunc)`
- Chef client: `github.com/go-chef/chef v0.30.1` — types like `chef.Node`, `chef.Role`, `chef.CookbookListResult` come from here
- Graceful shutdown: SIGINT/SIGTERM → context cancellation → server.Run returns
- `chef.Config.Client` field used to inject `safeclient.NewHTTPClient()` — this is the runtime read-only enforcement point
