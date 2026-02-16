#!/bin/bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
trap handle_interrupt INT

FEATURE_DESC="${1:-}"

init_workflow
echo "Starting feature workflow"

RESUME_PHASE=$(get_resume_phase)
case "$RESUME_PHASE" in
  ""|"discuss") START_STEP=1 ;;
  "spec-update") START_STEP=2 ;;
  "scope") START_STEP=3 ;;
  "test-red") START_STEP=4 ;;
  "implement") START_STEP=5 ;;
  "review") START_STEP=6 ;;
  "summarize-changes") START_STEP=7 ;;
  "synthesize") START_STEP=8 ;;
  "complete")
    echo "Workflow already complete. Delete .twincode/current-phase.json to restart."
    exit 0 ;;
  *) START_STEP=1 ;;
esac

if [[ $START_STEP -gt 1 ]]; then
  echo "Resuming from phase: $RESUME_PHASE (step $START_STEP)"
fi

# ── Phase 1: Discussion ──
if [[ $START_STEP -le 1 ]]; then
  run_interactive_phase "discuss" \
    "The developer's feature request: $FEATURE_DESC"
  advance_phase "spec-update"
fi

# ── Phase 2: Spec Update ──
if [[ $START_STEP -le 2 ]]; then
  run_interactive_phase "spec-update"
  advance_phase "scope"
fi

# ── Phase 3: Scope ──
if [[ $START_STEP -le 3 ]]; then
  run_interactive_phase "scope"
  advance_phase "test-red"
fi

# ── Phase 4: Test RED ──
if [[ $START_STEP -le 4 ]]; then
  run_interactive_phase "test-red"
  advance_phase "implement"
fi

# ── Phase 5: Implement GREEN ──
if [[ $START_STEP -le 5 ]]; then
  run_interactive_phase "implement"
  advance_phase "review"
fi

# ── Phase 6: Review ──
if [[ $START_STEP -le 6 ]]; then
  run_interactive_phase "review"
  advance_phase "summarize-changes"
fi

# ── Phase 7: Summarize Changes ──
if [[ $START_STEP -le 7 ]]; then
  run_interactive_phase "summarize-changes"
  advance_phase "synthesize"
fi

# ── Phase 8: Synthesis ──
if [[ $START_STEP -le 8 ]]; then
  run_interactive_phase "synthesize"
  advance_phase "complete"
fi

rm -f "$STATE_DIR/current-phase.json"
echo ""
echo "Feature workflow complete."
