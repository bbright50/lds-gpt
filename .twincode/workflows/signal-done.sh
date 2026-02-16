#!/bin/bash
# signal-done.sh — Signal phase completion to the workflow orchestrator
set -euo pipefail

STATE_DIR=".twincode"
phase="${1:-unknown}"

printf '{"status":"complete","phase":"%s"}\n' "$phase" > "$STATE_DIR/signal"
