#!/bin/bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
trap handle_interrupt INT

REFACTOR_DESC="${1:-}"

init_workflow
echo "Starting refactor workflow"

RESUME_PHASE=$(get_resume_phase)
case "$RESUME_PHASE" in
  ""|"discuss") START_STEP=1 ;;
  "test-red") START_STEP=2 ;;
  "implement") START_STEP=3 ;;
  "review") START_STEP=4 ;;
  "synthesize") START_STEP=5 ;;
  "complete")
    echo "Workflow already complete. Delete .twincode/current-phase.json to restart."
    exit 0 ;;
  *) START_STEP=1 ;;
esac

if [[ $START_STEP -gt 1 ]]; then
  echo "Resuming from phase: $RESUME_PHASE (step $START_STEP)"
fi

# ── Phase 1: Discussion (scope the refactor) ──
if [[ $START_STEP -le 1 ]]; then
  run_interactive_phase "discuss" \
    "The developer's refactoring request: $REFACTOR_DESC. Focus on: what to refactor, why (performance, complexity, testability, future feature, duplication), proposed approach, files affected, confirming external behavior unchanged."
  advance_phase "test-red"
fi

# ── Phase 2: Safety Net (ensure test coverage) ──
if [[ $START_STEP -le 2 ]]; then
  run_phase "test-red" \
    "Identify existing tests covering the code being refactored. Run them — they must all pass. If coverage is insufficient, write additional tests to lock in current behavior. This is a safety net, not TDD red phase." \
    "Read,Write,Edit,Bash,Grep,Glob"
  advance_phase "implement"
fi

# ── Phase 3: Refactor ──
if [[ $START_STEP -le 3 ]]; then
  run_phase "implement" \
    "Make structural changes in small, verifiable steps. After each significant change, run the test suite. If a test breaks, determine if the refactoring introduced a bug (fix it) or if the test was testing implementation details (update test but flag to developer)." \
    "Read,Write,Edit,Bash,Grep,Glob" \
    30
  advance_phase "review"
fi

# ── Phase 4: Review + Final Verification ──
if [[ $START_STEP -le 4 ]]; then
  run_phase "review" \
    "Run the full test suite one final time. Confirm all tests pass and no behavioral changes introduced. Review code quality." \
    "Read,Write,Edit,Bash,Grep,Glob"
  advance_phase "synthesize"
fi

# ── Phase 5: Spec Sync ──
if [[ $START_STEP -le 5 ]]; then
  run_phase "synthesize" \
    "If the refactoring changed internal architecture affecting spec (module boundaries, new internal interfaces), update relevant spec files and append to .twincode/spec/CHANGELOG.md." \
    "Read,Write,Edit,Grep,Glob"
  advance_phase "complete"
fi

rm -f "$STATE_DIR/current-phase.json"
echo ""
echo "Refactor workflow complete."
