---
description: Triage up to 10 open Renovate PRs one at a time (pick, analyze, act)
---
Renovate PR triage (writes labels + comments + local plans). This is a
sequential pipeline: each iteration triages exactly ONE open, untriaged
Renovate PR through three ordered stages (picker -> analyzer -> actor). Up
to 10 iterations run, but the loop stops as soon as the queue is empty —
never force all 10.

Labels applied (created in the repo, see .omp/skills/renovate-queue):
- `ready-to-merge` — no config/migration impact detected
- `config-change-needed` — our go.mod, Makefile pins, or generated CRDs
  must change with this bump
- `migration-needed` — upstream requires data/service migration steps
- `needs-human` — risky or undeterminable, review manually

Models resolve from the launching profile's default role — no pins here.

This is workflowz: run it from a single Python `eval` cell using the
eval-kernel's `agent(...)` helper, looping up to 10 times. For each
iteration, run the three stages strictly in order (never parallelize them —
`analyzer` depends on what `picker` wrote, `actor` depends on what
`analyzer` wrote), and read `.scratch/renovate/current.txt` between stages
so an empty queue costs one short iteration instead of ten:

```python
for i in range(10):
    agent(
        """
        Pick the next Renovate PR to triage.
        1. mkdir -p .scratch/renovate; touch .scratch/renovate/processed.txt
        2. List candidates:
           gh pr list --author app/renovate --state open \\
             --json number,title,labels
        3. Keep PRs labeled needs-triage (renovate.json labels new PRs with
           it), plus any PR carrying NO triage label at all (pre-dates the
           renovate label config). Discard PRs that already carry a verdict
           label (ready-to-merge, config-change-needed, migration-needed,
           needs-human) and PRs whose number is in
           .scratch/renovate/processed.txt.
        4. If none remain: write the single word NONE to
           .scratch/renovate/current.txt and stop.
        5. Otherwise write the lowest PR number to
           .scratch/renovate/current.txt and immediately append that same
           number to .scratch/renovate/processed.txt (claims it for this
           iteration even if later steps fail; failed PRs get needs-human
           on a future manual pass).
        """,
        agent="sonic",
    )

    current = read(".scratch/renovate/current.txt").strip()
    if current == "NONE":
        break
    n = current

    agent(
        f"""
        Analyze Renovate PR #{n} in this repo.
        1. gh pr view {n} and gh pr diff {n} - identify the dependency, old
           and new version, and which file(s) in this repo pin it (go.mod,
           Makefile, Dockerfile, workflow YAML, docker-compose).
        2. Research what changed between the two versions: the Renovate PR
           body's release notes first, then upstream changelog/releases via
           web_search or deepwiki. For digest-only bumps of the same tag,
           research is unnecessary - note "digest refresh" and move on.
        3. Check OUR usage - Go-native evidence, not vibes. Every check
           below is DIFFERENTIAL: a failure is evidence against this PR
           only if it does NOT already happen without it. Work in a
           scratch copy, never the live checkout:
             cp -r . /tmp/triage-{n} && cd /tmp/triage-{n}
           BASE = the repo as it stands. HEAD = BASE with this PR's pin
           applied (go get <module>@<new> for go.mod bumps, otherwise
           edit the pinned file the PR touches, then go mod tidy).
           a. go build ./... and go vet ./... on HEAD - does it fail? If
              it does, re-run on BASE and confirm BASE passes before
              blaming this PR. A repo that was already broken is a
              separate problem: say so once and move on.
           b. go mod why <module> - direct or transitive? A transitive-only
              breaking change is usually not our problem - say so
              explicitly; if transitive, go mod graph to show which direct
              dependency pulls it in.
           c. govulncheck ./... on BOTH states - reachability, and
              differential. Collect the GO-YYYY-NNNN ids from each:
                introduced (in HEAD, not BASE) - evidence against merging
                fixed      (in BASE, not HEAD) - evidence FOR merging,
                             and worth stating plainly: this PR reduces
                             exposure, which usually argues ready-to-merge
                baseline   (in both) - NOT evidence. Never cite a
                             pre-existing vulnerability as a reason to
                             hold this PR; the repo is already in that
                             state and this PR does not change it.
              govulncheck exits 3 only for vulns reachable from our code.
              The trailing "modules you require, but your code doesn't
              appear to call" section is informational, never a blocker.
           d. grep this repo for the changed symbol, flag, or env var
              across internal/, apis/, cmd/, Makefile, .github/workflows/.
              A breaking change we do not call is NOT a blocker - say so
              explicitly.
           e. Crossplane/upjet bumps specifically: does the CRD schema need
              regeneration (make generate / make reviewable)? If the bump
              touches the Terraform provider, does
              TERRAFORM_PROVIDER_VERSION in the Makefile need to move in
              lockstep with this go.mod bump?
        4. Decide exactly one verdict:
           ready-to-merge       - digest/patch bump, or changes do not touch
                                  our usage, or the only break is
                                  transitive and unreachable
           config-change-needed - renamed/removed flags, changed defaults,
                                  new required settings, or a go.mod bump
                                  that needs a matching Makefile/CRD change
           migration-needed     - DB schema/data migrations, manual upgrade
                                  steps, or ordered multi-step rollouts
           needs-human          - major bump with unclear impact, upstream
                                  changelog missing, or anything security
                                  sensitive you cannot verify
        5. Write .scratch/renovate/pr-{n}/analysis.md: verdict on line 1 as
           "verdict: <label>", then dependency, versions, evidence bullets
           (link every claim to a changelog/release/govulncheck URL), and
           our affected files with line references. State the differential
           result explicitly, even when it is empty - "introduced: none,
           fixed: none, baseline: 11 (pre-existing, not this PR's)" is a
           complete and useful line. If BASE is already red, say so once,
           near the top, so no reader mistakes it for this PR's doing.
        """,
        agent="task",
    )

    agent(
        f"""
        Act on the triage verdict for Renovate PR #{n}.
        1. Read .scratch/renovate/pr-{n}/analysis.md; take the verdict label.
        2. Apply it: gh pr edit {n} --add-label <verdict>. Remove
           needs-triage and any other verdict label if present
           (--remove-label).
        3. If the verdict is NOT ready-to-merge: post ONE comment on PR #{n}
           unless a comment containing the marker <!-- renovate-triage -->
           already exists (check gh pr view {n} --comments). The comment:
           the marker line, the verdict, a 3-8 line summary of what blocks
           the merge, and the evidence links from the analysis.
        4. If the verdict is config-change-needed or migration-needed:
           write .scratch/renovate/pr-{n}/plan.md - concrete steps to adapt
           this repo (files to edit, flags/CRD fields to rename, migration
           commands, verification per .omp/skills/go-verify/SKILL.md), so
           the PR can be implemented later via the renovate-queue skill.
        5. Never merge, close, or push anything. Labels, one comment, and
           .scratch/ files are your only writes.
        """,
        agent="task",
    )
```

Report, after the loop, how many PRs were actually triaged (iterations
before the `NONE` break) and their verdicts.
