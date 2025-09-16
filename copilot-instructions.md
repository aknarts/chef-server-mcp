# Internal Notes / Copilot Instructions

This file collects project-specific guidance for contributors and AI assistants.

## Architecture Overview
- cmd/chef-mcp: Entry point (wires config, version injection, graceful shutdown)
- internal/config: Environment configuration loader
- internal/chefapi: Thin wrapper around go-chef client (add more API calls here)
- internal/server: HTTP routes and fallback logic
- internal/knife: Execution wrapper for `knife` CLI
- internal/version: Build-time version variable

## Design Goals
1. Prefer Chef API; only use knife when unavoidable or API fails and fallback is enabled.
2. Keep server logic simple; push complexity (parsing, adaptation) into dedicated packages.
3. Avoid global state; allow dependency injection for testing (see NodeProvider + runKnife variable override).
4. Graceful shutdown & future extensibility (metrics, auth, tracing) kept in mind.

## Conventions
- New endpoints go in internal/server with minimal logic; delegate outward.
- Add interfaces when mocking is desirable (e.g., NodeProvider).
- Return JSON arrays/objects; no plain text except healthz.
- Log operationally relevant events (fallback usage, errors) at standard logger for now.

## Adding a New Resource Endpoint (Example: /cookbooks)
1. Extend chefapi with ListCookbooks().
2. Add interface method if mock/testing needed.
3. Add handler in server.go (pattern: try API -> fallback to knife `cookbook list`).
4. Write tests covering API success, API fail + knife success, both fail.

## Testing Strategy
- Unit test handlers with injected mocks & overridden runKnife.
- Avoid hitting real Chef in unit tests.
- (Future) Integration test suite could spin a disposable Chef server (container) if needed.

## Versioning
- Use semantic version tags (v0.x while unstable).
- CI injects version with: -ldflags "-X chef-mcp/internal/version.Version=$GIT_TAG".

## Knife Fallback Guidance
- Keep JSON normalization consistent (e.g., strip blank lines, sort if needed later).
- Instrument fallback usage (future metrics counter).

## Potential Enhancements (Low Risk)
- Introduce structured logging (zap or slog).
- Add /readyz separate from /healthz once dependencies expand.
- Add middleware chain (auth, request ID, recover, logging) in server.New().

## AI Assistant Tips
- Before adding a new file, check if similar functionality exists to avoid duplication.
- When editing server.go, preserve existing handlers—append new ones.
- Prefer small, composable functions; avoid large monolithic handlers.
- Ensure tests fail meaningfully (clear error messages).</n
## Security Considerations
- Never log private key contents.
- Validate and sanitize any future user-supplied inputs (filters, queries).
- Consider rate limiting & auth before external exposure.

## Release Checklist
- All tests pass.
- go vet / staticcheck (if added) clean.
- Docker image builds successfully.
- CHANGELOG (if introduced) updated.


