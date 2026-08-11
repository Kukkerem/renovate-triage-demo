---
name: go-verify
description: How to correctly verify a change in a Crossplane/upjet Go provider — the codegen-drift trap, the canonical build/test/lint gate, and the never-merge rule. Use before claiming a Go or Crossplane provider change builds.
---
# Verifying changes in this provider

`make reviewable` regenerates code (CRDs, deepcopy, conversion, docs) before
it lints and tests. `go build ./...` / `go test ./...` alone verify against
whatever generated code already happens to be checked out — if a dependency
bump changes a struct tag, an API type, or an upjet schema, stale generated
code can build clean and pass tests while being silently wrong. The signal
that matters is `make reviewable`'s **diff**, not its exit code: a clean
`git diff --exit-code` afterward means generation did not drift.

## Verify
`go build ./... && go vet ./... && go test ./... && make reviewable && govulncheck ./...`

- `go build ./...` / `go vet ./...` — does it compile, does vet see anything
  the compiler doesn't.
- `go test ./...` — unit and controller tests.
- `make reviewable` — regenerate (CRDs/deepcopy/conversion), lint, test;
  then `git diff --exit-code` to confirm generation produced no diff you
  have not reviewed and committed.
- `govulncheck ./...` — reachability, not mere presence. A flagged CVE with
  no reachable call path is not a blocker; report it, do not gate on it.

## Do NOT merge
Verification is the agent's job; merging is the human's. The agent runs the
gate above, reports pass/fail plus the `govulncheck` reachability verdict,
and stops there. `git commit`, `gh pr merge`, and pushing to the PR branch
are the human's to run, never the agent's.

## Common mistakes
- Trusting `go build ./...` alone on a bump that touches a CRD-defining
  type — it happily compiles against stale generated code; only
  `make reviewable`'s diff catches the drift.
- Treating any `govulncheck` hit as an automatic blocker without checking
  reachability — an unreachable CVE is not our problem, and calling it one
  trains people to ignore the report next time it matters.
- Running `make reviewable` and merging on a green exit code without
  reading its diff — an uncommitted diff after generation is not
  "reviewable", it is untested.
- Bumping go.mod without checking `TERRAFORM_PROVIDER_VERSION` (or any
  other dual-pinned tool version) in the Makefile — upjet needs both moved
  together, and nothing enforces that for you.
