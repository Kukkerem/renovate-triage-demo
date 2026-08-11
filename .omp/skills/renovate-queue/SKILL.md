---
name: renovate-queue
description: Use when processing Renovate dependency PRs in this repo — merging or implementing triage-labeled PRs one by one, when the user says "go through the renovate queue" or "check dependency PRs", or when open Renovate PRs lack triage labels.
---

# Renovate queue

Work through Renovate PRs in `Kukkerem/renovate-triage-demo` one at a time,
using the triage labels produced by the `/wf-renovate-triage` slash command
(`.omp/commands/wf-renovate-triage.md`). Renovate itself labels every new PR
`needs-triage` (renovate.json `labels`); the triage command replaces that
with a verdict label.

## Labels

| Label | Meaning | Default action |
|---|---|---|
| `ready-to-merge` | No config/migration impact detected | Merge |
| `config-change-needed` | Our go.mod, Makefile pins, or generated CRDs must change with this bump | Implement plan, then merge |
| `migration-needed` | Upstream requires data/service migration steps | Implement plan carefully, then merge |
| `needs-human` | Risky/undeterminable — triage could not decide | Investigate with the user |

Triage artifacts (local, gitignored): `.scratch/renovate/pr-<N>/analysis.md`
and, for the two action labels, `.scratch/renovate/pr-<N>/plan.md`.

## Flow

1. **Ensure triage.** `gh pr list --author app/renovate --state open --json number,title,labels`.
   Untriaged PRs present? Run `/wf-renovate-triage` first (or triage inline
   following the same rules from that command) — never guess a label ad hoc.
2. **Queue order:** `ready-to-merge` → `config-change-needed` →
   `migration-needed` → `needs-human`.
3. **Per PR — staleness gate first:** if the PR has commits newer than its
   `analysis.md` (Renovate force-pushes on rebase), the analysis is stale —
   re-triage before acting.
4. **Present** title, verdict, evidence summary (and plan if present), then
   `ask` the user: Merge / Implement / Skip / Close.
   - **Merge:** `gh pr merge <N> --squash --delete-branch`.
   - **Implement:** apply `plan.md` in a branch or directly per user
     preference; verify per the go-verify skill BEFORE merging the
     Renovate PR; land the config/CRD change and the PR together (order per
     plan).
   - **Close:** `gh pr close <N> --comment "<reason>"` and let the user
     decide whether to pin/ignore the dep in renovate.json.
5. **After the queue:** remind the user to cut a release (provider image
   build + CRD/xpkg package publish) — user-run, never the agent.

## Common mistakes

- Merging a `config-change-needed` PR that touches a CRD-defining type
  before `make generate` / `make reviewable` regenerates it and the diff is
  reviewed — the next `kubectl apply` rejects the CRD or silently drops
  fields. Regenerate, review the diff, then merge.
- Trusting `go build ./...` alone when the break is behavioural, not
  compile-time — a renamed default, a relaxed validation, or a changed CRD
  field still compiles clean. Only the targeted grep, `go test ./...`, or
  `govulncheck` surfaces it.
- Skipping `govulncheck` because the CVE looked unreachable from the
  changelog prose alone — reachability is a call-graph fact the scanner
  computes, not something you can eyeball from a description.
- Trusting `analysis.md` after Renovate force-pushed a rebase (step 3 gate)
  — the diff it describes no longer matches HEAD.
- Applying a triage label manually without an evidence URL — labels must
  trace to a changelog/release/govulncheck URL, else use `needs-human`.
