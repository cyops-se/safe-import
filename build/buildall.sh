#!/bin/sh
# Builds all safe-import components. Run from the repository root.
set -eu
cd "$(dirname "$0")/.."
ROOT=$(pwd)

GIT_COMMIT=$(git rev-list --abbrev-commit -1 HEAD)
GIT_VERSION=$(git describe --tags --dirty --always)
LDFLAGS="-s -w -X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"

for c in si-outer si-inner si-gatekeeper; do
    echo "Building $c"
    (cd "$ROOT/$c" && CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS")
done

echo "Building si-engine (including web UI)"
"$ROOT/si-engine/build.sh"
