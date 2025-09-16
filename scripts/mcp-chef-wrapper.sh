#!/usr/bin/env bash
# Auto-rebuilding wrapper for the chef MCP server.
# Rebuilds the mcp-chef binary (unless NO_REBUILD=1) then execs it for MCP stdio use.
# Honors existing environment variables:
#   CHEF_USER, CHEF_KEY_PATH, CHEF_SERVER_URL, KNIFE_FALLBACK
# Optional:
#   NO_REBUILD=1   -> skip rebuilding
#   VERBOSE=1      -> show build steps
#   SKIP_VERSION_INJECT=1 -> build without dynamic git describe (falls back to Makefile logic)
set -euo pipefail

export PATH=$PATH:/usr/local/go/bin

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist"
BINARY="${DIST_DIR}/mcp-chef"

if [[ "${VERBOSE:-0}" == "1" ]]; then
  set -x
fi

if [[ "${NO_REBUILD:-0}" != "1" ]]; then
  # Build only the MCP binary for speed.
  make -C "${REPO_ROOT}" build-mcp >/dev/null
fi

if [[ ! -x "${BINARY}" ]]; then
  echo "Error: binary not found at ${BINARY}" >&2
  exit 1
fi

# Exec replaces this process so MCP host sees a stable stdio lifecycle.
exec "${BINARY}"

