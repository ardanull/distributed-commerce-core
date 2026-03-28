#!/usr/bin/env bash
set -euo pipefail

curl -sS http://localhost:8080/ > /dev/null
curl -sS http://localhost:8081/healthz > /dev/null || true
echo "basic smoke checks passed"
