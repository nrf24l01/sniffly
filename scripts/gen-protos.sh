#!/usr/bin/env bash
set -euo pipefail

# Portable script to generate Go protobuf and gRPC code for all .proto files in the repository.
# Usage (from repository root):
#   ./scripts/gen-protos.sh
# It does not hardcode any user paths. It will install protoc-gen-go/protoc-gen-go-grpc
# using 'go install' if they are missing from PATH.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "Repo root: $REPO_ROOT"

command -v protoc >/dev/null 2>&1 || {
  echo "protoc not found in PATH. Please install protoc (libprotoc)." >&2
  exit 1
}

ensure_plugin() {
  local bin="$1"
  local install_cmd="$2"
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "$bin not found in PATH — installing via: $install_cmd"
    # Use go install to put binary into $(go env GOPATH)/bin or $GOBIN
    eval "$install_cmd"
    if ! command -v "$bin" >/dev/null 2>&1; then
      echo "Failed to install $bin. Make sure GOPATH/GOBIN is configured and in PATH." >&2
      exit 1
    fi
  fi
}

# Ensure protoc-gen-go and protoc-gen-go-grpc are available
ensure_plugin protoc-gen-go 'go install google.golang.org/protobuf/cmd/protoc-gen-go@latest'
ensure_plugin protoc-gen-go-grpc 'go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest'

PROTO_ROOT="$REPO_ROOT"

echo "Searching for .proto files..."

found=0
while IFS= read -r -d '' proto; do
  found=1
  echo "\nGenerating for: $proto"
  dir=$(dirname "$proto")
  base=$(basename "$proto")
  # Run protoc from the proto's directory and use that directory as proto_path
  # This avoids duplication when using paths=source_relative
  (cd "$dir" && \
    protoc \
      --proto_path="." \
      --go_out=paths=source_relative:. \
      --go-grpc_out=paths=source_relative:. \
      "$base")
done < <(find "$REPO_ROOT" -name '*.proto' -print0)

if [ "$found" -eq 0 ]; then
  echo "No .proto files found in repository."
fi

echo "Done. Generated files are next to their .proto sources."
