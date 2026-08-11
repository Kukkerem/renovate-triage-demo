# renovate-triage-demo

A miniature stand-in for [`SAP/crossplane-provider-btp`](https://github.com/SAP/crossplane-provider-btp)
(Go, Crossplane/Upjet, ~350 LOC instead of tens of thousands), built to
demonstrate an AI-agent Renovate PR triage workflow. It is **not** a real
provider — the `internal/*` packages exist to give real dependency APIs a
real call site, so that `go build`, `go vet`, `go test` and `govulncheck`
all say something true about them.

**Every direct dependency here is deliberately out of date.** Base images,
pinned tool versions and Go modules are all held at an old value on purpose,
so that pointing Renovate at this repo produces a queue of genuine update
PRs — a patch bump, a major bump that really does break compilation, a
deprecated module, a mutable action tag, a database major version, and a
pre-release stability change.

**What each of those bumps *means* is deliberately not written down in this
repo.** No answer key, no expected-verdict table, no per-seed notes. Working
that out from the diff, the upstream changelog and this repo's own call
sites is precisely the job being demonstrated — an agent that could look up
the answer in `README.md` would prove nothing. The verdicts live in the
presenter's notes, outside this repository, and are compared against the
agent's output after the fact.

## Layout

```
cmd/provider/main.go          wires every internal/ package into one real call graph
internal/auth/                JWT signing/parsing        (github.com/golang-jwt/jwt)
internal/config/              provider config timestamps (github.com/golang/protobuf)
internal/cfclient/            Cloud Foundry API client   (github.com/cloudfoundry/go-cfclient/v3)
internal/htmlscan/            HTML title extraction      (golang.org/x/net/html), reachable from main
internal/format/              display-name casing        (golang.org/x/text)
package/Dockerfile            Terraform provider ARG + tag+digest base image
Makefile                      TERRAFORM_PROVIDER_VERSION, pinned alongside go.mod
test/e2e/docker-compose.yaml  Postgres service backing the (future) e2e suite
.github/workflows/ci.yaml     build/vet/test pipeline
.github/workflows/e2e.yaml    CROSSPLANE_CHART_VERSION/_SHA256 for the e2e cluster
.github/workflows-v2/         a design sketch, NOT a live workflow — see its README
renovate.json                 SAP/crossplane-provider-btp's real config, retargeted
.omp/                         the triage tooling itself: one command, two skills
```

## The triage tooling

The point of this repo is what lives in `.omp/`:

| File | Role |
|---|---|
| `.omp/commands/wf-renovate-triage.md` | writes verdict labels — picker → analyzer → actor, up to 10 PRs |
| `.omp/skills/renovate-queue/SKILL.md` | reads those labels back and walks a human through the queue |
| `.omp/skills/go-verify/SKILL.md` | the Go/Crossplane verification gate both rely on |
| `.omp/verdict.schema.json` | the closed 4-value verdict enum |

Roughly 240 lines of markdown and a JSON schema. There is no service, no
database, and no state anywhere except GitHub labels and a gitignored
`.scratch/` directory. **The label is the API.**

Neither half ever merges, closes, or pushes anything. Labels, one comment
per PR, and local files are their only writes.

## Running it

```console
$ go build ./...
$ go vet ./...
$ go test ./...
$ go run ./cmd/provider
```

No network access or credentials are required for any of the above — the
Cloud Foundry client (`internal/cfclient`) is only dialed when `CF_API_ROOT`
is set, which none of the commands above do.
