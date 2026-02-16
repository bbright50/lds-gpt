#!/bin/bash
# lib.sh — shared TwinCode workflow functions

STATE_DIR=".twincode"
SIGNAL_FILE="$STATE_DIR/signal"
readonly DEFAULT_MAX_TURNS=20
_TWINCODE_PID=""
_TAIL_PID=""
_FORMATTER_PID=""
_STREAM_FILE=""
_STREAM_FIFO=""

# Debug logging — enabled via TWINCODE_DEBUG=1
debug() {
  if [[ "${TWINCODE_DEBUG:-}" == "1" ]]; then
    echo "[DEBUG] $*" >&2
  fi
}

# Clean up streaming processes and temp files.
_cleanup_stream() {
  for pid in "$_TWINCODE_PID" "$_TAIL_PID" "$_FORMATTER_PID"; do
    if [[ -n "$pid" ]]; then
      kill "$pid" 2>/dev/null || true
    fi
  done
  for pid in "$_TWINCODE_PID" "$_TAIL_PID" "$_FORMATTER_PID"; do
    if [[ -n "$pid" ]]; then
      wait "$pid" 2>/dev/null || true
    fi
  done
  _TWINCODE_PID=""
  _TAIL_PID=""
  _FORMATTER_PID=""
  rm -f "$_STREAM_FILE" "$_STREAM_FIFO"
  _STREAM_FILE=""
  _STREAM_FIFO=""
}

# Interrupt handler — kills the active claude process and exits.
handle_interrupt() {
  _cleanup_stream
  restore_terminal
  echo ""
  echo "Workflow interrupted."
  exit 130
}

init_workflow() {
  debug "init_workflow: STATE_DIR=$STATE_DIR"
  mkdir -p "$STATE_DIR"
  rm -f "$STATE_DIR/phase-result.json"
  rm -f "$SIGNAL_FILE"
  # current-phase.json is preserved for resume across sessions
}

# Read the resume phase from current-phase.json (empty string if missing).
get_resume_phase() {
  local phase_file="$STATE_DIR/current-phase.json"
  if [[ -f "$phase_file" ]]; then
    local phase
    phase=$(jq -r '.phase // empty' "$phase_file" 2>/dev/null || true)
    debug "get_resume_phase: found '$phase' in $phase_file"
    echo "$phase"
  else
    debug "get_resume_phase: no phase file found"
  fi
}

# Restore terminal to canonical (cooked) mode after an interactive TUI session.
# Claude Code uses raw mode for its TUI; if it exits without fully restoring
# settings, select/read will not process line-buffered input correctly.
restore_terminal() {
  stty sane 2>/dev/null || true
}

# Record which phase should run next. Called after each phase completes
# (before its gate) so that a future `twincode run` can resume here.
advance_phase() {
  local next_phase="$1"
  debug "advance_phase: $next_phase"
  printf '{"phase":"%s"}\n' "$next_phase" > "$STATE_DIR/current-phase.json"
}

# Format stream-json JSONL into human-readable output.
# Shows assistant text and tool call summaries as they happen.
_format_stream() {
  jq --unbuffered -r '
    if .type == "assistant" then
      (.message.content[]? |
        if .type == "text" then .text
        elif .type == "tool_use" then
          "→ " + .name + (
            if .name == "Read" then " " + .input.file_path
            elif .name == "Edit" then " " + .input.file_path
            elif .name == "Write" then " " + .input.file_path
            elif .name == "Bash" then " $ " + (.input.command | .[0:120])
            elif .name == "Glob" then " " + .input.pattern
            elif .name == "Grep" then " " + .input.pattern
            else ""
            end)
        else empty end)
    else empty end
  ' 2>/dev/null
}

# Run a phase as a fresh, stateless session.
# Claude signals completion by writing .twincode/signal (via Write or Bash tool).
run_phase() {
  local phase_name="$1"
  local extra_prompt="$2"
  local allowed_tools="$3"
  local max_turns="${4:-$DEFAULT_MAX_TURNS}"
  local result_key="${5:-}"

  echo ""
  echo "── Phase: $phase_name ──"
  echo ""

  debug "Journaling current phase: $phase_name"
  printf '{"phase": "%s"}\n' "$phase_name" > "$STATE_DIR/current-phase.json"
  rm -f "$STATE_DIR/phase-result.json"
  rm -f "$SIGNAL_FILE"

  local signal_hint
  if [[ -n "$result_key" ]]; then
    signal_hint="IMPORTANT: When you have completed all phase criteria, signal completion by writing .twincode/phase-result.json with the key \"$result_key\" set to true. For example: {\"$result_key\": true}. Write this file as your LAST action."
  else
    signal_hint="IMPORTANT: When you have completed all phase criteria, signal completion by writing the file .twincode/signal with the exact content: PHASE_COMPLETE"
  fi

  local prompt
  prompt="You are executing the TwinCode '$phase_name' phase. Read .twincode/WORKFLOW.md and follow the instructions under '## Phases > ### $phase_name'. $extra_prompt $signal_hint"

  local -a args=(-p "$prompt" --verbose --output-format stream-json --max-turns "$max_turns")

  if [[ -n "$allowed_tools" ]]; then
    args+=(--allowedTools "$allowed_tools")
  fi

  debug "run_phase: claude ${args[*]}"

  # Stream architecture: claude → file → tail -f → FIFO → jq → stdout
  # The file + tail -f layer avoids pipe buffering that suppresses
  # real-time output from claude's stream-json mode. The FIFO lets us
  # track tail and jq as separate PIDs for clean shutdown.
  _STREAM_FILE=$(mktemp)
  _STREAM_FIFO=$(mktemp -u)
  mkfifo "$_STREAM_FIFO"

  tail -f "$_STREAM_FILE" > "$_STREAM_FIFO" &
  _TAIL_PID=$!

  _format_stream < "$_STREAM_FIFO" &
  _FORMATTER_PID=$!

  claude "${args[@]}" > "$_STREAM_FILE" 2>/dev/null &
  _TWINCODE_PID=$!
  debug "run_phase: claude=$_TWINCODE_PID tail=$_TAIL_PID fmt=$_FORMATTER_PID"

  wait "$_TWINCODE_PID" || true
  _TWINCODE_PID=""

  # Give formatter a moment to drain, then tear down
  sleep 0.5
  _cleanup_stream
  debug "run_phase: phase '$phase_name' complete"
}

# Run a phase interactively. The developer chats with Claude directly.
# Context comes from the markdown files on disk.
run_interactive_phase() {
  local phase_name="$1"
  local extra_prompt="${2:-}"

  echo ""
  echo "── Phase: $phase_name (interactive) ──"
  echo ""

  # Write current phase for the Stop hook
  debug "Journaling current phase: $phase_name"
  printf '{"phase": "%s"}\n' "$phase_name" > "$STATE_DIR/current-phase.json"
  rm -f "$STATE_DIR/phase-result.json"
  rm -f "$SIGNAL_FILE"

  local prompt
  prompt="You are executing the TwinCode '$phase_name' phase. Read .twincode/WORKFLOW.md and follow the instructions under '## Phases > ### $phase_name'. $extra_prompt

PHASE TRANSITION: Track the phase criteria listed in WORKFLOW.md for this phase. When the developer indicates they want to move on (e.g., 'next phase', 'move on', 'continue'), check all criteria. If all criteria are met, confirm and tell the developer to exit the session (Ctrl+C or /exit) to advance to the next phase. If criteria are unmet, list what remains and ask the developer if they want to address it or skip."

  debug "run_interactive_phase: claude --append-system-prompt (phase=$phase_name)"
  echo "Starting interactive session for phase: $phase_name... (Ctrl+C to advance)"
  claude --append-system-prompt "$prompt"
  restore_terminal
  debug "run_interactive_phase: claude exited for phase '$phase_name'"
}

# Gate: developer approval with interactive discussion option
gate() {
  local gate_name="$1"
  while true; do
    echo ""
    echo "════════════════════════════════════════"
    echo "  Gate: $gate_name"
    echo "════════════════════════════════════════"
    echo ""
    select choice in "Accept — continue to next phase" \
                     "Discuss — open interactive session" \
                     "Abort — stop the workflow"; do
      case $REPLY in
        1) return 0 ;;
        2)
          claude
          restore_terminal
          break ;;  # re-show menu after interactive exit
        3)
          echo "Workflow aborted."
          exit 1 ;;
        *)
          echo "Invalid choice. Please enter 1, 2, or 3."
          break ;;
      esac
    done
  done
}

# Check phase criteria (read hook result file)
check_phase_result() {
  local key="$1"
  if [[ -f "$STATE_DIR/phase-result.json" ]]; then
    local val
    val=$(jq -r ".$key // false" "$STATE_DIR/phase-result.json")
    debug "check_phase_result: $key=$val"
    echo "$val"
  else
    debug "check_phase_result: no phase-result.json found"
    echo "false"
  fi
}
