# TwinCode Workflow Definition

## Global Rules

### Hard Gates
- No implementation without completed discussion + test phases
- Every bug fix must include a regression test

### Escape Hatch
- Changes that don't alter feature set or public API may use shortened loop
- Must still: explain change, get approval, update spec if behavior changes

### Gate Behavior
At every [GATE], the developer can:
- **Accept** — proceed to next phase
- **Reject** — explain consequences, offer alternatives
- **Modify** — re-enter interactive discussion

## Phases

### discuss
Read the project spec. Ask the developer clarifying questions about their
request. After each response, state your confidence level (0-100%) in
understanding what the developer wants to build. Continue asking questions
until you reach at least 85% confidence. Only then offer: "Continue
refining or proceed with sensible defaults?" End with a summary of
understood requirements.

**Criteria:**
- [ ] Requirements summarized
- [ ] Developer confirms understanding

### spec-update
Determine which spec files to create or update. Read related specs and
check for contradictions. Draft changes covering architecture decisions,
data models, API contracts, behavioral requirements.

**Criteria:**
- [ ] Spec changes drafted
- [ ] No unresolved contradictions
- [ ] Changes written to .twincode/spec/ files

### scope
Break the agreed work into concrete, implementable tasks. Write tasks
to .twincode/TASKS.md with checkboxes. Each task should be small enough to
implement and test in one iteration.

**Criteria:**
- [ ] .twincode/TASKS.md populated with tasks
- [ ] Each task is concrete and testable

### test-red
For each task: define the interface/contract using language-appropriate
constructs (Go: interface, TypeScript: interface/type, Python: Protocol/ABC,
Rust: trait). Write tests against the interface. Run tests — they MUST fail.

**Criteria:**
- [ ] Interface/contract defined
- [ ] Tests written
- [ ] Tests run and FAIL (red)

### implement
Write minimal implementation to make tests pass. Run tests after
implementation. Do not add functionality beyond what tests require.

**Criteria:**
- [ ] Implementation written
- [ ] All tests PASS (green)

### review
Check: error handling, input validation, no hardcoded secrets,
immutable patterns, function size (<50 lines), file size (<800 lines).
Fix issues found.

**Criteria:**
- [ ] Code review checklist passed
- [ ] Issues fixed

### summarize-changes
Explain the changes you made, why each was necessary, and the reasoning process that led to those decisions.
Describe how these changes fit into the repository's architecture and coding patterns.
Be specific enough that a developer unfamiliar with the context could reproduce both the changes and the problem-solving approach.
Highlight any trade-offs considered or alternatives rejected.
Invite the developer to ask clarifying questions about any part of the implementation.

**Criteria:**
- [ ] Developer is satisfied with the changes

### synthesize
Re-read entire .twincode/spec/ directory. Compare each spec against implementation.
Update specs to reflect code as built. Check for gaps (code not in spec)
and phantom specs (spec not in code). Append dated entry to
.twincode/spec/CHANGELOG.md. Move completed tasks from .twincode/TASKS.md to CHANGELOG.
Clear .twincode/TASKS.md.

**Criteria:**
- [ ] Spec files updated
- [ ] .twincode/spec/CHANGELOG.md entry appended
- [ ] .twincode/TASKS.md cleared

### bootstrap
Explore the codebase: structure, languages, frameworks, dependencies,
major domains. Ask 3-5 questions about purpose, users, key features,
architecture decisions. Generate .twincode/spec/INDEX.md, domain specs,
.twincode/spec/CHANGELOG.md, .twincode/spec/SESSION_STATE.md, .twincode/TASKS.md.

**Criteria:**
- [ ] .twincode/spec/INDEX.md populated
- [ ] At least one domain spec created
- [ ] Developer reviewed and accepted

## Spec Artifacts

| File | Purpose |
|------|---------|
| .twincode/spec/INDEX.md | High-level architecture + spec directory |
| .twincode/spec/<domain>.md | Domain/feature specs (flexible structure) |
| .twincode/spec/CHANGELOG.md | Rolling log of spec changes + completed tasks |
| .twincode/spec/SESSION_STATE.md | Incomplete work, next steps |
| .twincode/TASKS.md | Current iteration task list |

## Explanation Style
- Default: moderate detail — summarize each file, design decisions, and why
- Complex code: scale up to line-level for concurrency, security, algorithms
- Simple changes: brief summary

## Session Protocol

### On Start
1. Read .twincode/spec/INDEX.md
2. Read .twincode/spec/SESSION_STATE.md
3. Read .twincode/TASKS.md
4. Drift detection: compare spec against code, flag discrepancies

### On End
If work is incomplete, write .twincode/spec/SESSION_STATE.md with:
- What's complete, what remains
- Next steps for the next session
