#!/usr/bin/env bash
set -euo pipefail
psql "postgres://app:app@localhost:5432/commerce?sslmode=disable" -f migrations/001_init.sql
echo "migrations applied"
