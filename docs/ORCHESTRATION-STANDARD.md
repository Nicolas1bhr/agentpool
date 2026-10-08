# Orchestration Standard — how delegated and multi-agent work is run

**Version 2026-10-08** · synced from TradeAgent `docs/HOW-WE-BUILD.md` + `docs/FLEET.md` at `bac86ab1` — the owner made the
orchestrator seat the standard on 2026-10-06 and asked that the orchestration documents be upgraded to carry it.
**Read-gate:** before ANY delegated or multi-agent work in this repository — one agent or a fleet, a feature or a one-line fix —
read this file and the project's `CLAUDE.md` / `AGENTS.md`. Every brief tells its agent to read both first.
**Scope:** how the BUILD is run — seats, briefs, gates, landing, honesty. It is not a design for any product's own agents. It
outranks no protection in the project's instruction files; where those are stricter, they win.

This file is identical in every project that carries it. Project facts — the gate commands, the protected surfaces, the machines,
the record file — live in the project's instruction file under **Orchestration**. If that section names no gate, establishing
one is the first unit. To change this standard, change TradeAgent's two source documents, then re-sync every copy.

## Why it is shaped this way

A round-based triad — build, verify, cross-model review, repeated per round — produced 82 commits of which 9 touched product
code, 16,000 lines of briefs and records, and nothing landed, when it was ported to a small deliverable. It is priced for a
multi-tenant SaaS with fifty migrations. What follows keeps the hierarchy and the honesty and deletes the passes. A project whose
blast radius needs more (multi-tenant data, policies that leak silently) adds review legs on that surface in its own file — it
never removes a rule below.

## 1. Seats

| Seat | Decides | Never |
|---|---|---|
| **Orchestrator** — the main session, holding the owner's own seat over the build (§2) | what sits ABOVE the seats: which seats exist (at most two at a time) and their lanes; priorities across seats; builder allotments; cross-seat conflicts; the channel to the owner; the heartbeat; memory and the resume block | product code; running a builder; landing a unit; deciding inside a manager's lane |
| **Top-level manager** — one per seat, a fresh agent | everything inside its lane: which units to brief, run and land, in what order within the session's priorities; how each is briefed; every report and deviation judged; surveys, fixers, landing, the record, CI per sha | product code (its builders write it); another seat's lane |
| **Builder · fixer · survey leg** — fresh, spawned by a manager | how to build its one unit: one brief, one worktree, one pass | anything outside its brief; `main` |

- **Every agent is Opus**, with `model: "opus"` set explicitly on every spawn — the owner's choice. Codex legs are not Claude agents.
- **Capacity is set by the machine and the 5-hour usage window, not by the weekly budget** (the owner, 2026-10-06: "No need to
  budget the weekly rate limit"). Reference machine, an M3 Pro with 18 GB: at most four builders fleet-wide. Two managers, four
  builders and the orchestrator empty a window in ~75–110 min, so spend each window on the highest-value units first.
- **The seats are a ceiling, not a quota.** Small work may run with no manager: the orchestrator dispatches one builder and lands
  it itself under §6.
- A manager runs its builders itself and talks to the orchestrator by `SendMessage` to `main`. A sub-agent is never woken by its
  background child finishing — that notice goes to the orchestrator.

## 2. The orchestrator — the owner's own seat over the build

- **What it is.** It decides sequencing, allotments and cross-seat questions on the owner's behalf — conservative, compliant,
  reversible, each decision written down with its reason, his to overrule — and it is the one channel to him. Its managers keep
  their full authority, freedom and judgement inside their lanes. It does **not** hold his authority over money, credentials,
  legal status, paid commitments, releases or anything sent in his name: each still needs his explicit yes.
- **Build session or meeting — his words decide.** In a meeting or planning session the orchestrator lands its conclusions as
  docs and arms no heartbeat, opens no seat, dispatches no builder. "Implementation can start whenever" means allowed later, not now.
- **Starting a build session:** (1) arm the heartbeat first — crons die with the session that set them; (2) check the machine —
  power, lid state, free disk (a closed lid on battery sleeps the fleet); (3) read usage — the 5-hour window and its reset;
  (4) read the state from git, worktrees, locks, CI runs and the status and handoff files — dated text is not proof of state;
  (5) open FRESH seats (agent ids die with their session) whose prompts name the charter, handoff and status files, the session's
  rules and its priorities; each manager plans its lane on its own judgement and says so when it sees a better order. Write the board.
- **Each wake.** A builder's completion notice reaches the orchestrator, which wakes the manager with the report's facts — tips,
  gate counts, CI per platform, deviations, NOT VERIFIED — and, where it helps, its view. The manager judges the report itself;
  the orchestrator rules only on what sits above a seat and says which is which. A message that changes nothing for a running
  agent waits for its next wake. The orchestrator reads reports, the board and status files — never code.
- **A red that no diff can reach is a FIRST SIGHTING**, recorded with its run id and a one-line reading — never "a flake" by
  assertion; platform-only reds have been real defects. One that threatens a protection is surveyed at once.
- **Waiting is detached.** Waiting in the foreground costs a whole context per call: every seat and builder waits with a
  background command that re-invokes it on exit (or `nohup` for durable ledgers). A builder ends a waiting turn with one line,
  `WAITING: …`, which is not a report and wakes no one.
- **Stalls.** A seat whose transcript shows no progress for 15 minutes is stalled: stop it and resume it by `SendMessage` with its
  state. Run commands plainly, with `< /dev/null`, never nested quoting — a nested `zsh -c` once sat on stdin for 38 minutes.
  Board times come from `date`; agents' self-reported clocks have run 4–25 minutes fast.
- **The heartbeat.** Session crons carry the fleet across usage stops: a one-shot wake at each window's reset + 4 min, re-armed at
  every wake, and a recurring two-hourly backstop. A cron fires only while the session is idle, which a usage stop leaves it. On
  the wake: read usage, the machine and the state, and resume every leg that died at the stop with one `SendMessage` naming its
  branch state and the CI runs that finished meanwhile. Nothing is lost, because the branch is the handoff. Crons are
  session-only; a recurring one expires after seven days.
- **The morning report and the wrap-up.** The report: what landed (record shas, gate counts, CI per platform), what is building,
  what only the owner can give, and every NOT VERIFIED. The wrap-up, when he asks or the session has been fruitful enough: seats
  land what is in flight and start nothing new but a protection's fix; each writes its status and handoff; the orchestrator
  checkpoints the resume block, the board, its own handoff and memory, removes its crons, and reports.

## 3. Honesty

Every claim is "verified by running X → output" or "NOT VERIFIED", with what was tried. There is no third category. Banned: should
work, looks correct, probably, I believe, minor, trivial, static-verified. **A real machine that is available is used:** a run
that could have been made and was not is NOT VERIFIED. The project's record file keeps this rule.

## 4. Safety

The project's instruction file names its protected surfaces — money, credentials, authorization, data isolation, evidence,
recovery, or whatever its blast radius is. A change on a protected surface ships with a test that was RED before the fix, and the
builder watches ONE mutant of the guard go red and quotes it. That is the whole proof burden; there is no separate mutant sweep.

## 5. Two passes, then it lands

**Pass 1 — build.** One fresh builder, one brief of at most 40 lines at `docs/briefs/<unit>.md`, one branch, its own worktree at
`~/Projects/<repo>-worktrees/<branch>` (use `git -C`, never `cd` into one inside a compound command). The builder rebases onto
`main` first and resolves any conflict itself; builds the whole unit; writes red-first tests on protected surfaces; runs the
project's gate; commits per item with one-sentence messages; and appends a `## Report` of at most 20 lines to its own brief: tip
sha, gate counts pasted, CI run and verdicts, one line per item, and what it did NOT do. The report is the record. A builder may
push its own branch to run CI — never `main`, never a merge — and never uses the app's built-in browser pane (a permission prompt
only the owner can answer once hung a builder for 80 minutes); it reads web sources with `curl` or WebFetch.

**Pass 2 — land.** The manager runs §6 on the reported tip, writes a record section of at most 40 lines from the report, and
deletes the brief. `docs/briefs/` holds only work in flight; empty means nothing is. `docs/queue/` holds READY briefs not yet
dispatched: before dispatch, re-check each pointer and dependency against `main` — a disagreement is fixed in the docs first,
never in a builder's prompt — then `git mv` the brief into `docs/briefs/`. Built once, landed once.

## 6. The landing checklist

1. `git status --porcelain` clean in the builder's worktree; the tip equals the reported sha.
2. `git rebase main` if `main` moved; a conflict goes back to a builder — the manager does not resolve it.
3. The project's full gate on the landed tree, counts pasted. One full gate at a time per machine (overlapping suites flake
   timing tests). Two units each green alone can be red together: that is a one-item fixer on the landing branch, and neither
   unit's builder is reopened. Only docs moved since the unit's last gate: write "GATE CARRIES".
4. Test-name diff against `main` → nothing removed. A deleted test cannot fail; it has happened three times. The one exception —
   a test that existed only for something the project has deliberately dropped — is named in the report.
5. Secret scan of the whole diff against `main`, as a gate, not a neighbouring command. A judged false positive is excluded by
   name in that one call, never by loosening the pattern.
6. `git merge --ff-only` with its exit status checked — never behind a pipe — then `git rev-list --count <branch>..main` must print
   0 before any record is written; push; CI green on every platform at the merge sha. Red CI in the product: undo the merge —
   reset and `--force-with-lease`, or a revert commit where pushing `main` deploys or others pull from it (the project file says
   which) — then a fixer on the branch. Red only on a hosted runner, in a test the target platform passes: a fresh fixer on top of
   `main`, and the sha is recorded red until it lands. **A `Timing` category is the one place a second attempt exists:** a test
   joins it only when its verdict needs the runner to keep a wall clock, membership is argued at the test with measured numbers,
   and an assertion is never loosened to get in. A red twice in a row is still a red.
7. Record section written; brief deleted; worktree removed; memory updated.

## 7. The fresh-fixer rule

A builder gets one attempt per item. When the gate fails on an item, or the report says NOT FIXED, PARTIAL or "I could not",
the manager does not ask that builder to try again. It stops the leg — the per-item commits are on the branch — and dispatches a
fresh agent with a fixer brief: the item, the failing command and its output, the file and line, and nothing of the previous
builder's reasoning. A builder that has failed carries the wrong model of the problem, and more context makes that worse. Two
fresh fixers failing on the same item means the item is mis-stated or structural: the manager rewrites it as a class fix. It
does not send a third fixer. Two findings with one root cause get one structural fix.

A rate-limit or process kill is not a failure: resume a live leg with one message; otherwise re-brief from the file on disk,
reading the branch first, because the fixes may already be there.

## 8. The milestone review

Once per milestone, as the last step before a release: one fresh Opus reviewer told to break the protected surfaces on `main` at a
named sha, and Codex read-only on the same sha in its own worktree, in parallel. Findings go to `docs/REVIEW-<date>.md` as one
table, one line per finding: severity, `file:line`, the check that settles it. Each HIGH becomes a fix unit before the release;
MED and LOW together one batch unit. Fix units go through the two passes; a red-first test is their proof; no re-review.

## 9. Context preservation

- **Status file per seat** (`fleet/status/<seat>.md`, ≤ 40 lines), rewritten — not appended — at every state change: units and
  their state, branch tips, live agent ids, what is in flight and owed, the last CI verdicts. With git, enough to resume the seat.
- **Hand-off** (`fleet/handoff/<seat>.md`, ≤ 40 lines: state, open judgements, traps met) when a seat's scope ends or its context
  passes ~60 %. The orchestrator rotates a seat — around 400–550k tokens, when each wake costs more than a fresh seat's grounding —
  and opens the fresh seat from the handoff and status files.
- The `fleet/` directory lives at `~/Projects/<repo>-worktrees/fleet/` — outside the repo, and outside `/tmp`, which the OS empties.
  Nothing durable lives in a scratchpad: the orchestrator copies what a seat will need out of it.
- **The repo is the checkpoint:** a record per landing, the resume block at each wave's end and at every stop, the memory files.
  Keep contexts lean: scripts print summaries; read long files by range; never print a whole log.

## 10. Escalation

Builder → its manager (the final report, a blocker named in it). Manager → orchestrator by `SendMessage`, first line
self-contained: a landing, a blocker, a decision outside the seat, an owner question, a budget or machine problem — routine
progress goes in the status file. **A question for the owner travels through the orchestrator,** which carries it to him whole,
for what only he can give: money, credentials, a paid commitment, a release, a product decision the docs leave open.

## 11. Mechanics every seat carries

- Checkpoint into the repo, never the scratchpad. External facts — vendor APIs, prices, rules — are read from the official source
  on the day and dated.
- A shared resource (a test machine, a seeded database, a deploy target) is one leg's at a time, by explicit grant, and its tree
  is proven by hash before a figure from it counts. Locks for the full suite, the landing and commits to the main checkout.
- Secret scan as a gate before every commit and push. `--ff-only` into `main`.
- **No `Co-Authored-By` or "Generated with" lines in commits** — the owner's rule, which overrides any tool default. Credit goes
  in the README and the build docs.

## 12. What is gone

Round numbers. Bounce briefs, verify briefs and Codex prompts as files. Verify records with per-round sections, the manager log,
Codex transcripts in the repo. Tiers as leg allocation. The combination verify. The integration scribe. Foreground polling.
Retrying a builder that has failed. A project may keep an extra review leg on its catastrophic-if-wrong surface; it says so in
its own file.

## 13. Sizes, so that this stays true

Brief ≤ 40 lines. Report ≤ 20. Record section ≤ 40. Status and handoff files ≤ 40. Review table one line per finding. Anything
that needs more is two units.
