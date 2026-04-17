# Specification: Agentic RAG Flow (ReAct)

> **Status:** Design (unimplemented). Target architecture; nothing under
> `internal/agent/` exists yet. Update this doc as the phases below land so
> it stays the source of truth.

## 1. Purpose

Replace the fixed three-stage `DoContextualSearch` pipeline with a
**ReAct-style agent** that reasons and acts in alternating steps: it writes
its own Cypher queries against the graph schema, sees the rows, thinks
about what to do next, and either issues another query or composes a final
grounded answer. The conversation is multi-turn; follow-ups ("which of
those is in the Book of Mormon?") reuse prior reasoning traces via
`chat.Session`.

Existing retrieval primitives (`xxxSimilar` kNN, per-label traversal) don't
disappear — they become first-class **actions** the agent can emit. Hand-
crafted non-agentic search (`app.App.DoContextualSearch`) stays as a
fallback path.

## 2. Architecture choice: ReAct

We evaluated three patterns:

- **(A) Single-shot Cypher** — LLM writes one query, we run it, LLM answers from rows. Fast and cheap, but dies on any question requiring more than one pass of data.
- **(B) ReAct loop** — LLM iterates (*Thought → Action → Observation → Thought → …*), up to `maxIters` turns. Robust on compound questions; every turn is a full model inference. **This is what we're building.**
- **(C) Tool-use with primitives** — LLM picks and parameterises a fixed tool vocabulary via structured JSON. Less flexible than ReAct; better suited to models with native function-calling APIs.

**Why ReAct for this project:** we have a richly connected graph where
compound questions are common ("topics Paul mentions that also appear in
Alma, ranked by how often") and a mid-size local model (`gemma4:31b`). The
iterative loop matters more than token efficiency — the user isn't paying
per call, only per second, and a 1-minute answer is fine.

**The cost we accept:** each `Run` does up to N full model inferences. On a
31B local model that's seconds-to-tens-of-seconds per turn. `maxIters` is
a hard wall-clock cap, not an aspirational budget.

## 3. The ReAct loop

Each turn, the agent sends the running trace to the chat model and stops
generation at the literal token `Observation:`. Whatever the model emitted
before that is parsed as a *Thought + Action + Action Input*. The agent
runs the action, and the tool's result becomes the next `Observation:`
appended to the trace for the next turn.

```text
User: Which topical-guide topics show up most in verses about faith?

Thought: I should start by finding verses semantically about "faith".
Action: semantic_search
Action Input: {"label": "VerseGroup", "query": "faith", "k": 20}
Observation: [{"id":"vg/nt/heb/11/1-5","snippet":"..."}, ...]

Thought: Good, 20 seeds. Now I'll expand to the TG topics those verses
footnote, and aggregate.
Action: cypher_query
Action Input: {"query":"UNWIND $ids AS gid MATCH (g:VerseGroup {id: gid})-[:INCLUDES]->(v:Verse)-[:TG_FOOTNOTE]->(t:TopicalGuideEntry) RETURN t.name AS topic, count(*) AS hits ORDER BY hits DESC LIMIT 10","params":{"ids":["vg/nt/heb/11/1-5", ...]}}
Observation: [{"topic":"Faith","hits":47},{"topic":"Believe","hits":23},...]

Thought: I have enough to answer.
Action: final_answer
Action Input: {"text":"The topics that appear most in verses about faith are…","citations":["tg/Faith","tg/Believe",...]}
```

The loop terminates on `final_answer` or when `maxIters` is hit (returns a
"ran out of steps" answer with the best evidence gathered so far).

## 4. Action vocabulary

The agent can emit **exactly one** of these as the `Action:` value. Tight
vocabulary keeps the parser simple and the safety story coherent; inside
`cypher_query` the LLM still has full Cypher expressiveness.

```text
semantic_search(label, query, k=10)
    → [{id, name, distance, snippet}]
    Embeds `query`, runs the existing per-label xxxSimilar kNN. Exists
    because pure Cypher has no way to do semantic similarity without a
    pre-computed embedding, and the LLM can't embed text itself.

cypher_query(query, params={})
    → [{...row}]  (rows capped at 500 after LIMIT injection)
    Arbitrary read-only Cypher. The agent's main workhorse. Passes
    through cypher_guard.Validate before execution. Rejected queries
    surface as an `error` observation so the agent can revise and retry
    in the next turn.

final_answer(text, citations=[])
    → terminates the loop
    `citations` are node ids the agent used. The caller surfaces them
    alongside the text.
```

Note the intentional absence of a separate `expand` action: one-hop
traversal is just a short Cypher query the LLM can write directly.

## 5. Prompt structure

```text
You are a scripture-study assistant grounded in the LDS standard works,
topical guide, bible dictionary, triple combination index, and Joseph
Smith Translation. Answer only from data you retrieve via Actions.

You must follow the ReAct format for every turn:
  Thought: <one or two sentences of reasoning>
  Action: <one of: semantic_search | cypher_query | final_answer>
  Action Input: <single-line JSON>

Stop after Action Input. The system will append an Observation for you.

## Graph schema
{{auto-generated from internal/falkor/schema.graphql — node labels, their
properties, each relationship with direction and target label, which
properties are vector-indexed (embedding / summaryEmbedding)}}

## Actions
{{action signatures from §4}}

## Cypher rules
- Read-only only: no CREATE, MERGE, SET, DELETE, REMOVE.
- Allowed CALLs: db.idx.vector.queryNodes, db.indexes(), db.labels(),
  db.relationshipTypes(). All others rejected.
- Always parameterise user input.
- If you omit LIMIT, 500 is imposed.
- Every query has a 5s server-side timeout.

## Examples
{{2–3 few-shot traces: each a full Thought → Action → Action Input →
Observation → ... → final_answer exchange. Pick examples that cover
both semantic_search and cypher_query.}}
```

The schema summary is regenerated once per agent boot from `schema.graphql`
so it always matches live code.

## 6. Key Components (planned)

| File | Role |
|---|---|
| `internal/agent/agent.go` | `Agent` struct; `Run(ctx, session, userQ) (Answer, error)` drives the ReAct loop |
| `internal/agent/actions.go` | Action dispatch table + `semantic_search`/`cypher_query`/`final_answer` implementations wrapping existing retrieval code |
| `internal/agent/parser.go` | Extract `Thought:` / `Action:` / `Action Input:` blocks from model output; classify parse failures for retry |
| `internal/agent/cypher_guard.go` | Read-only Cypher validator + auto-`LIMIT` injection |
| `internal/agent/schema_prompt.go` | Generates the compact schema summary injected into the system prompt from `internal/falkor/schema.graphql` |
| `internal/agent/prompt.go` | System prompt template + few-shot trace exemplars |
| `internal/agent/agent_test.go` | Loop-behaviour tests with a mock `chat.Client` returning canned traces |
| `internal/agent/parser_test.go` | Golden-style tests against representative model outputs (well-formed, whitespace-noisy, malformed) |
| `internal/agent/cypher_guard_test.go` | Table-driven: destructive statements rejected, missing LIMIT injected, allowed CALLs pass |
| `cmd/chat/main.go` | Interactive stdin REPL holding a `chat.Session` + `Agent`, with `/reset` and `/history` meta-commands |

## 7. Data Models

- **`Agent`** — `{chatClient chat.Client, fc *falkor.Client, embed embedding.Client, systemPrompt string, maxIters int, stopAt string}`. `stopAt` is `"Observation:"`; the generation halts there so the model can't hallucinate its own tool results.
- **`Step`** — `{Thought string, Action string, ActionInput json.RawMessage}`. One parsed assistant turn.
- **`Observation`** — `{Action string, Rows any, Error string}`. Serialised back into the trace as an `Observation:` JSON blob.
- **`Answer`** — `{Text string, Citations []string, Steps []Step, TurnsUsed int}`. Full provenance — the UI can show the trace in an expandable panel.
- **`parseError`** — distinguished type so the loop can append a corrective message ("Please respond in the Thought/Action/Action Input format.") and retry **without** consuming an iteration. Capped at 2 parse retries per turn to avoid infinite loops on a model that genuinely cannot comply.

## 8. Interfaces

- **`Agent.Run(ctx, session *chat.Session, userQ string) (Answer, error)`** — appends `userQ` to the session, runs up to `maxIters` turns, returns `Answer` on a `final_answer` action or on the cap (with `Text` explaining the cap and `Citations` from whatever rows were seen). Errors only on transport/ctx issues; parse / validation failures are fed back into the trace and don't bubble up.
- **Trace shape in `chat.Session`** — every ReAct iteration extends the session: the assistant turn is `"Thought: ...\nAction: ...\nAction Input: ..."`, and the Observation is appended as a fresh `user`-role message whose content is `"Observation: {...json...}"`. That way a follow-up `Run` on the same session sees the entire previous reasoning as context.
- **Stop tokens** — `chat.Client` gains an optional `Stop []string` option. Ollama supports this via `options.stop`. Essential — without it the model will keep going and invent Observations.

## 9. Safety rails on `cypher_query`

All four apply to every invocation:

1. **Parse-time allowlist.** Statements must be built only from `MATCH`, `OPTIONAL MATCH`, `WHERE`, `WITH`, `RETURN`, `ORDER BY`, `LIMIT`, `SKIP`, `UNWIND`, and restricted `CALL`.
2. **`CALL` allowlist.** Only `db.idx.vector.queryNodes`, `db.indexes()`, `db.labels()`, `db.relationshipTypes()`. Anything else rejected.
3. **LIMIT injection.** If the parsed query lacks a top-level `LIMIT`, wrap as `CALL { <original> } WITH * LIMIT 500`. The agent may request a smaller limit; never larger.
4. **Server-side timeout.** `GRAPH.CONFIG SET TIMEOUT_DEFAULT 5000` (5 s), applied once by the agent constructor alongside the existing `RESULTSET_SIZE` lift in `Migrate`. Bounds runtime even if the guard misses something.

Rejected queries do NOT abort the turn — they come back to the model as
`Observation: {"error":"validator: <reason>"}` so it can revise. This is
the ReAct trait we care about: the model gets to fix its own mistakes
inline instead of surfacing errors to the user.

## 10. Trace-length management

Every turn appends a Thought + Action + Observation to the trace. Ten
turns of 1 KB each is already 10 KB of context on top of the system
prompt (which is itself non-trivial once the schema is dumped in).
`gemma3:27b` has a 128K context — plenty in absolute terms — but latency
scales roughly with context length, so keeping the trace lean matters.

Policy:
- Summarise Observations longer than **2 KB** (truncate row arrays, replace with `"rows_truncated": N`).
- Hard cap: if the trace exceeds **32K chars** mid-run, replace all Observations except the most recent two with `"(summarised)"` and a one-line synopsis emitted by a cheap summarisation call (or just `"Observation: <N rows returned, first id: ...>"` without any LLM help).
- Surface hitting this cap to the caller via `Answer.Notes` so the user knows their query walked far.

## 11. Dependencies

- **Depends on**: `internal/chat` (session + `/api/chat` wrapper, extended with `Stop []string`), `internal/embedding` (for `semantic_search`), `internal/falkor` (`Raw()` for `cypher_query`, + its indexes/timeouts), `internal/falkor/schema.graphql` (read at boot for the prompt).
- **Depended on by**: `cmd/chat` (interactive REPL), eventually an HTTP `/chat` handler consumed by `frontend/`.

## 12. Testing strategy

**Unit** (no Ollama, no FalkorDB):
- `parser_test.go` — golden table: well-formed output parses correctly, trailing whitespace tolerated, missing `Action Input` raises `parseError`, stray prose before `Thought:` is tolerated (model over-explains).
- `agent_test.go` — mock `chat.Client` returns a canned ReAct trace (multi-turn) and asserts: each Action was dispatched in order, Observations were injected correctly, loop terminates on `final_answer`, `maxIters` cap returns the expected "ran out of steps" `Answer`.
- `cypher_guard_test.go` — table-driven: destructive keywords rejected, unsafe `CALL` rejected, missing `LIMIT` injected, params preserved, existing `LIMIT` not duplicated.
- `schema_prompt_test.go` — golden-file snapshot of the generated prompt; fails when `schema.graphql` drifts (forces a conscious update).

**Integration** (testcontainer FalkorDB + mock chat):
- Seed a small fixture graph, feed the agent a canned ReAct trace exercising both `semantic_search` and `cypher_query`, assert `Answer.Citations` point at real node ids.
- Test a trace that includes a deliberately invalid Cypher → expect the `Observation` error, a corrective Thought, a revised Cypher, successful termination.

**End-to-end** (opt-in via env flag, skipped by default):
- Real Ollama + real FalkorDB, a dozen hand-picked questions with expected citation *categories* (not exact ids — LLM output is non-deterministic). Regression check, not a correctness oracle.

## 13. Phased implementation

1. **Parser + schema prompt.** Before any loop, prove we can turn model output into `Step`s reliably and that the schema summary lands well in the prompt. Manual harness: paste prompt into Ollama UI, eyeball output quality across ~20 hand-crafted questions.
2. **Agent skeleton + `final_answer` only.** Simplest possible single-turn loop. `Run` sends prompt + user question, parses one Step, if it's `final_answer` we return, else error. Validates the stop-token + parsing path before layering on any real actions.
3. **`semantic_search` action.** Adds the simplest real action. Two-turn traces start working (one search + one answer).
4. **`cypher_query` + guard.** The big step. Full ReAct becomes possible. Land with the guard and its tests in the same PR — safety and capability together.
5. **Trace-length management.** Once we see real traces, implement the truncation policy from §10. Don't build speculatively.
6. **CLI REPL (`cmd/chat`).** `chat.Session` + `Agent` behind a stdin loop with `/reset`, `/history`, `/trace` meta-commands. This is the first daily-driver surface.
7. **HTTP endpoint.** Only once the frontend is ready to consume it; stream Thought + Action deltas SSE-style so the UI can show the agent "thinking" live.

## 14. Open questions

1. **Format compliance of `gemma4:31b`.** ReAct lives or dies on the model emitting `Thought:` / `Action:` / `Action Input:` headers consistently. First-week experiment: in phase 1 run the schema prompt + 20 hand-crafted questions, measure the malformed-output rate. If >10%, either (a) retry with a stricter corrective message, (b) switch to a ReAct-trained model (e.g. `qwen2.5:32b`, `llama3.1:70b`), or (c) use llama.cpp grammar-constrained decoding to force the shape.
2. **Stop-token reliability.** Ollama honours `options.stop` but some builds strip the token and keep generating. Verify with a short probe in phase 2; fall back to post-hoc truncation at the first `Observation:` we see in output if the server-side stop is unreliable.
3. **Citation format.** Node ids like `vg/ot/gen/1/1-5` aren't human-readable. Either the agent emits display strings alongside ids in `final_answer`, or the frontend hydrates via a lookup query. Prefer the latter — keeps the LLM's job small.
4. **Max iterations.** Start at `5`. On a 31B local model that's roughly a 30–90 s wall-clock cap per question. Tune after end-to-end baseline.
5. **Cancellation.** Ctx cancellation must abort mid-generation cleanly, not wait for the current turn to finish. Requires forwarding the `ctx` through the HTTP client without hacks and handling partial responses in the parser. Build this in phase 2 — easier to get right small than to retrofit.
