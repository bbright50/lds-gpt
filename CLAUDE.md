# TwinCode: Guided Enterprise Development System

This project uses a specification-driven development workflow
defined in `.twincode/WORKFLOW.md`. Read it before performing development work.

## Session Start Protocol
1. Read `.twincode/spec/INDEX.md` for system architecture
2. Read `.twincode/spec/SESSION_STATE.md` for incomplete work
3. Read `.twincode/TASKS.md` for in-progress tasks
4. Drift detection: compare spec against code, flag discrepancies

If `.twincode/spec/INDEX.md` is empty, prompt the developer to run `/bootstrap`.

## Hard Gates
You MUST refuse to write implementation code unless:
1. Discussion phase completed — requirements understood and agreed
2. Test phase completed — tests written and verified to FAIL

See .twincode/WORKFLOW.md for the escape hatch (simple changes).

## Regression Guard
Every bug fix MUST include a regression test. No exceptions.

## Multi-Developer Awareness
- Each developer works independently
- Spec conflicts resolved via Git merge
- Always read latest spec from disk, never rely on session memory
