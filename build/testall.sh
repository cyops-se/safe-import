#!/bin/sh
# Starts all components locally for manual testing (development only).
# For a production like setup use docker compose, see README.md.
# Run build/buildall.sh first. Stop everything with Ctrl+C.
set -eu
cd "$(dirname "$0")/.."
ROOT=$(pwd)

trap 'kill 0' INT TERM EXIT

nats-server &
sleep 1
(cd "$ROOT/si-outer" && ./si-outer) &
sleep 1
(cd "$ROOT/si-inner" && sudo --preserve-env ./si-inner) &
sleep 1
(cd "$ROOT/si-gatekeeper" && ./si-gatekeeper) &
sleep 1
(cd "$ROOT/si-engine" && ./si-engine)
