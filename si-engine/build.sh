#!/bin/sh
# Builds the web UI into web/admin/static (embedded in the binary) and then si-engine.
# Set SKIP_UI=1 to build only the Go binary (the UI will then not be available).
set -eu
cd "$(dirname "$0")"

GIT_COMMIT=$(git rev-list --abbrev-commit -1 HEAD)
GIT_VERSION=$(git describe --tags --dirty --always)

if [ "${SKIP_UI:-0}" != "1" ]; then
    find web/admin/static -mindepth 1 ! -name .gitkeep -delete
    (
        cd ui
        npm ci --no-audit --no-fund
        # webpack 4 needs the legacy OpenSSL provider on Node >= 17
        NODE_OPTIONS=--openssl-legacy-provider npx vue-cli-service build --no-clean --dest ../web/admin/static
    )
fi

CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"
