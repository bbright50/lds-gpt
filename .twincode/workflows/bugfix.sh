#!/bin/bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
trap handle_interrupt INT

BUG_DESC="${1:-}"

init_workflow
echo "Starting bugfix workflow"

RESUME_PHASE=$(get_resume_phase)
case "$RESUME_PHASE" in
  ""|"discuss") START_STEP=1 ;;
  "test-red") START_STEP=2 ;;
  "implement") START_STEP=3 ;;
  "spec-update") START_STEP=4 ;;
  "review") START_STEP=5 ;;
  "complete")
    echo "Workflow already complete. Delete .twincode/current-phase.json to restart."
    exit 0 ;;
  *) START_STEP=1 ;;
esac

if [[ $START_STEP -gt 1 ]]; then
  echo "Resuming from phase: $RESUME_PHASE (step $START_STEP)"
fi

# ── Phase 1: Discussion (understand the bug) ──
if [[ $START_STEP -le 1 ]]; then
  run_interactive_phase "discuss" \
    "The developer's bug report: $BUG_DESC. Focus on understanding: what is happening, what should happen, how to reproduce. Identify root cause or narrow to candidates."
  advance_phase "test-red"
fi

# ── Phase 2: Regression Test RED (retry loop) ──
if [[ $START_STEP -le 2 ]]; then
  for attempt in 1 2 3; do
    run_phase "test-red" \
      "Write a regression test that reproduces the bug. The test MUST fail, confirming the bug exists." \
      "Read,Write,Edit,Bash,Grep,Glob" \
      $DEFAULT_MAX_TURNS \
      "tests_fail"

    if [[ "$(check_phase_result tests_fail)" == "true" ]]; then
      echo "Regression test fails as expected (RED). Proceeding."
      break
    fi

    if [[ $attempt -eq 3 ]]; then
      echo "Regression test did not fail after 3 attempts."
      gate "Regression Test — needs intervention"
    fi
  done
  advance_phase "implement"
fi

# ── Phase 3: Fix GREEN (retry loop) ──
if [[ $START_STEP -le 3 ]]; then
  for attempt in 1 2 3; do
    run_phase "implement" \
      "Implement the minimal fix to make the regression test pass. Run the full test suite for the affected area." \
      "Read,Write,Edit,Bash,Grep,Glob" \
      $DEFAULT_MAX_TURNS \
      "tests_pass"

    if [[ "$(check_phase_result tests_pass)" == "true" ]]; then
      echo "All tests pass (GREEN). Proceeding."
      break
    fi

    if [[ $attempt -eq 3 ]]; then
      echo "Tests did not pass after 3 attempts."
      gate "Bug Fix — needs intervention"
    fi
  done
  advance_phase "spec-update"
fi

# ── Phase 4: Spec Update (if needed) ──
if [[ $START_STEP -le 4 ]]; then
  run_phase "spec-update" \
    "If the bug revealed a gap in the spec, update the relevant spec file and append to .twincode/spec/CHANGELOG.md. If no spec update is needed, state that and move on." \
    "Read,Write,Edit,Grep,Glob"
  advance_phase "review"
fi

# ── Phase 5: Review ──
if [[ $START_STEP -le 5 ]]; then
  run_phase "review" \
    "Review the fix. Flag related areas that might have the same bug." \
    "Read,Write,Edit,Bash,Grep,Glob"
  advance_phase "complete"
fi

rm -f "$STATE_DIR/current-phase.json"
echo ""
echo "Bugfix workflow complete."
