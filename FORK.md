# The Giant Swarm line of go-selfupdate

This repository is Team Bumblebee's fork of [creativeprojects/go-selfupdate](https://github.com/creativeprojects/go-selfupdate),
the library every Giant Swarm command-line tool installs its releases with, behind the cosign bundle check of
[giantswarm/selfupdate-cosign](https://github.com/giantswarm/selfupdate-cosign). It carries upstream fixes the CLIs
need before upstream releases them, and answers one question from one place: **which go-selfupdate do the CLIs run, and
why does it differ from upstream?**

Tracking: [giantswarm/selfupdate-cosign#17](https://github.com/giantswarm/selfupdate-cosign/issues/17) for the CLIs' side;
the line's own issues live in this repository.

## Branches

| Branch | What it is |
|---|---|
| `giantswarm` (default) | **The line**: the upstream release tag the CLIs run ("the pin") + the carried patches + this fork's own files (`FORK.md`, `fork/`, `renovate.json5`, `.circleci/config.yml`, the generated `zz_generated.*` workflows, and upstream's `build.yml` running on this branch too). Every change is a pull request against it: a carried patch by rebase merge, the fork's own files by squash merge ([Contributing](#contributing)). |
| `main` | upstream `main` at the time of the fork; not consumed, not synced. |
| `upstream/<slug>` | a carried patch prepared for upstream: its one commit on upstream `main`, on upstream's module path |
| `fork/<topic>` | pull-request branches against `giantswarm` |

## Pin

Upstream **v1.7.1** (`5dd040ba`). The line moves to the next upstream release the CLIs take, with every
carried patch kept in effect on it or dropped once upstream contains it.

## Moving the pin

A pull request from `fork/v<version>` against `giantswarm` moves the pin: it merges the upstream tag into the line
(`git merge --no-ff v<version>`), resolving conflicts so the carried patches stay in effect on the new upstream code,
every import keeps this line's module path and the fork's own files stay as the line has them; upstream's tests and
`go vet ./...` pass before it opens. It lands with a merge commit, so the new upstream tag is an ancestor of the line.
The protected branch is never force-pushed.

## Module path

The line's module path is `github.com/giantswarm/go-selfupdate`: the `module` line in `go.mod` and every internal
import name it. This is the one Giant Swarm-specific commit on the line (`chore: module path github.com/giantswarm/go-selfupdate`),
never offered upstream; its ledger file is [`fork/module-path.md`](fork/module-path.md). A carried patch goes upstream
without it, on upstream's module path ([Contributing](#contributing)). The line's pipeline (below) checks the tree with `goimports -local
github.com/giantswarm/go-selfupdate`, which wants the line's own packages in an import group of their own and every
file in gofmt's form. The line therefore carries, as Giant Swarm-only formatting, the regrouped imports of `update.go`
and `internal/path_test.go`, in the lines the module path commit rewrites anyway.

## CI

The line's own pipeline is the hand-written `.circleci/config.yml`: the architect orb's stock `go-test` job
(`go test ./...`, gosec, and nancy, the dependency vulnerability scan) on every branch and tag, next to upstream's
`build.yml`, run as upstream has it (its Go matrix and golangci-lint version) with `giantswarm` added to its branches,
while the orb's job runs the architect image's current Go. It is Giant Swarm CI only, never offered upstream, and never generated: the line's entry in
giantswarm/github declares `gen.ci.generate: false`, so the file is edited here directly.

nancy audits the modules of the built packages against OSS Index, by CVE id and module version. A finding with a
fixed version is a bump of that module in `go.mod` on the line, by hand (Renovate stays off, `renovate.json5`), a
`fix:` pull request of the fork's own; the next pin move keeps the higher version, so the bump never conflicts with
upstream for long. A finding without a fix is a time-boxed `.nancy-ignore` entry, `<id> until=<date> # <module>@<version>`
with its tracking issue in a comment line, removed when the issue closes or the date passes; the scan is never
bypassed, and no entry outlives its finding. GO-2026-5932 (`golang.org/x/crypto/openpgp`, unmaintained) is no entry:
upstream's v1.7.0 took the package out of the module's imports, and nancy knows no such id. The line's x/crypto is
v0.58.0 (CVE-2026-56855, CVE-2026-78662 in `x/crypto/ssh`, fixed in v0.56.0), whose `go` directive sets the
module's to 1.26.0: the Go a consumer of the line needs from the release carrying it on.

## Consumers

A consumer imports this path directly, in place of upstream's, with no `replace`:

```go
import selfupdate "github.com/giantswarm/go-selfupdate"
```

```
require github.com/giantswarm/go-selfupdate v1.6.1
```

A `replace` never reaches `go install <module>@<version>` users, which ignores every `replace` of the installed
module, and a library cannot pass one on to its importers. With the line under its own path, a library that imports it
(selfupdate-cosign) carries it to every CLI transitively: the CLI's graph contains the line, and upstream's module only
if something else requires it. The import goes back to upstream's path at the first upstream release that contains every carried patch, and this line
retires.

## Releases

The line cuts its own semver tags on `giantswarm`, in its own module path's version space, continuing upstream's
numbers from the pin: the first is **v1.6.1** on upstream v1.6.0, then a patch release per change of the line, and
the minor or major of the upstream release a pin move takes. A tag is cut by the Auto Release workflow
(`.github/workflows/zz_generated.auto_release.yaml` and `cliff.toml`, generated from the line's entry in
giantswarm/github) from the conventional subject of the commit merged into `giantswarm`: every releasable merge becomes
the next release candidate `vX.Y.Z-rc.N` and a pre-release on GitHub, and the stable `vX.Y.Z` is that workflow run by
hand with `release-type: stable`, which tags the latest candidate's commit. A pin move keeps the rule above: its merge commit's subject names the bump (`feat:` for an upstream minor, a
breaking marker for an upstream major, `fix:` otherwise), so the tag follows the number of upstream's release.
A Go module needs no build or publish step, so the Go proxy serves
the stable tag as it is (a release candidate resolves only when a consumer asks for it by version).
Upstream's tags on the line (v1.6.0 and older) name upstream's module path and are not consumable from this one.

## Carried patches

Each change the line carries has its own ledger file in [`fork/`](fork/); the directory listing is the set of carried
changes, and a file's name is the change's stable key. A change adds or edits only its own file, never this one. On
this line a carried patch is rebase-merged, its ledger file in a `docs(fork):` commit of its own next to the
upstream-ready commit; the fork's own files are squash-merged. A ledger file has a title naming the change and:

- **Status:** `carried`, or `dropped` once a re-pin onto an upstream release that contains it makes it redundant; a
  Giant Swarm-specific change stays `carried` for as long as the line exists
- **Upstream:** the upstream pull request or issue and its state; until a pull request is opened, "to file — prepared
  in the fork: branch `upstream/<slug>`", with the upstream commit the branch is based on and the upstream search;
  `never` for a Giant Swarm-specific change
- **Our commit:** the commit on `giantswarm` and the pull request that landed it
- `## Change`: what the change does
- `## Reason`: why the line carries it

## Contributing

- **Upstream first, prepared here.** Before a change, search upstream's issues and pull requests, open and closed:
  work upstream already has in flight is supported with our evidence, not duplicated. Every behavioural change is one
  upstream-ready commit: upstream's conventions and tests (upstream's `CONTRIBUTING.md`), no fork references, its
  message the upstream pull request's text and its subject conventional (`fix: …`), which upstream welcomes and Auto
  Release reads. Its pull request against `giantswarm`, from a `fork/<topic>` branch, carries that commit and the
  `docs(fork):` commit with its ledger file.
- **Prepared on upstream's module path.** On the line the patch imports this line's module path; upstream's code
  imports `github.com/creativeprojects/go-selfupdate`. The patch commit alone is cherry-picked onto upstream `main` as
  an `upstream/<slug>` branch on this repository, never taken from the line with the [module path](#module-path)
  commit: an import of the module's own packages goes back to upstream's path, and upstream's tests pass there. A
  change to a prepared or opened patch takes the same way. The ledger file's **Upstream** reads "to file — prepared in
  the fork: branch `upstream/<slug>`" until it is opened, and links the upstream pull request from then on.
- **A human opens it upstream.** The maintainers review the prepared branches in batches, decide what is useful to
  upstream and open it; the upstream pull request then names the fork commit. An agent never opens, comments on,
  reviews or closes a pull request or issue on the upstream repository.
- **Public only.** This repository is public: a ledger file, a commit message, a pull request and this file cite
  public issues and pull requests only; where the reasoning lives in a private tracker it is written out in a sentence
  instead. A pull request or commit message here names no upstream issue or pull request, so that nothing on this
  fork shows up on upstream's timeline: the ledger file carries those links.
- **Merging.** A carried patch is rebase-merged: its upstream-ready commit lands as it is, author and message kept, so
  the line's history is upstream's plus the patches. A fork-only change (this file, `renovate.json5`,
  `.circleci/config.yml`, a `go.mod` bump for a nancy finding) is squash-merged; a pin move lands with a merge commit
  ([Moving the pin](#moving-the-pin)).
- **Own tags.** Auto Release cuts the line's semver tags from the conventional subjects of what lands on `giantswarm`
  ([Releases](#releases)): a carried patch's `fix:` or `feat:` cuts the next release candidate; a `docs(fork):` or
  `chore:` change alone cuts none and ships with the next releasable commit.
- **Own CI.** `.circleci/config.yml` is hand-written and edited here directly, Giant Swarm CI only, never offered
  upstream ([CI](#ci)); a nancy finding is a bump or a time-boxed `.nancy-ignore` entry, never a bypassed scan.
- **Renovate is off** (`renovate.json5`): upstream owns the dependency graph, and dependencies move when the line is
  re-pinned, or by hand for a nancy finding.
