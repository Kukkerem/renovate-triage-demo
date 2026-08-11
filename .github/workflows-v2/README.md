# Not a live workflow

`renovate-triage.yaml` in this directory is a **design sketch**, not an
active workflow. It lives in `workflows-v2/` rather than `workflows/` on
purpose:

- GitHub only executes files under `.github/workflows/`, so nothing here
  ever runs. It has no Anthropic credentials configured and would fail if
  it did run.
- Renovate's `github-actions` manager only matches `.github/workflows/`,
  so the action versions pinned here do not generate dependency PRs and
  cannot pollute the demo queue.

It is here to be *read*: it shows the two-job split (`analyze` holds no
GitHub write scope; `publish` makes zero model calls and validates the
verdict against `../../.omp/verdict.schema.json` before applying a label).

The thing that actually runs today is `.omp/commands/wf-renovate-triage.md`,
driven locally.
