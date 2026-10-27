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
| `giantswarm` (default) | **The line**: the upstream release tag the CLIs run ("the pin") + the carried patches + this fork's own files (`FORK.md`, `renovate.json5`, `.circleci/config.yml`, the generated `zz_generated.*` workflows, and upstream's `build.yml` running on this branch too). Every change is a pull request against it: a carried patch by rebase merge (below), the fork's own files by squash merge. |
| `main` | upstream `main` at the time of the fork; not consumed, not synced. |
| `fork/<topic>` | pull-request branches against `giantswarm` |

## Pin

Upstream **v1.6.0** (`27a5f170`, 2026-07-08). The line moves to the next upstream release the CLIs take, with every
carried patch kept in effect on it or dropped once upstream contains it.

## Changing the line

A carried patch lands by rebase merge: one upstream-ready commit, author and message kept, so the line's history is
upstream's plus the patches and the patch is ready to go upstream as it is. The protected branch is never force-pushed.

A pull request from `fork/v<version>` against `giantswarm` moves the pin: it merges the upstream tag into the line
(`git merge --no-ff v<version>`), resolving conflicts so the carried patches stay in effect on the new upstream code,
every import keeps this line's module path and the fork's own files stay as the line has them; upstream's tests and
`go vet ./...` pass before it opens. It lands with a merge commit, so the new upstream tag is an ancestor of the line.

## Module path

The line's module path is `github.com/giantswarm/go-selfupdate`: the `module` line in `go.mod` and every internal
import name it. This is the one Giant Swarm-specific commit on the line (`chore: module path github.com/giantswarm/go-selfupdate`),
never offered upstream. The upstream pull request of a carried patch stays what it was opened from: the patch commit
alone, on upstream's module path; a change to it is cherry-picked onto upstream `main` as an `upstream/<slug>` branch on this repository,
queued as a row on [giantswarm/giantswarm#37742](https://github.com/giantswarm/giantswarm/issues/37742), never taken from the line with the module path commit. A human
decides what goes upstream and opens it; an agent never opens, comments on, reviews or closes anything on the
upstream repository. The line's pipeline (below) checks the tree with `goimports -local
github.com/giantswarm/go-selfupdate`, which wants the line's own packages in an import group of their own and every
file in gofmt's form. The line therefore carries, as Giant Swarm-only formatting: the regrouped imports of `update.go`
and `internal/path_test.go` (in the lines the module path commit rewrites anyway), gofmt's doc-comment form of
`doc.go` and `update/doc.go`, and the sorted imports of `mockdata_test.go` and `validate_test.go` (the carried
openpgp patch's files). The last two are gofmt's output and upstream's to take as they are; a re-pin keeps upstream's
version wherever upstream has formatted them.

## CI

The line's own pipeline is the hand-written `.circleci/config.yml`: the architect orb's stock `go-test` job
(`go test ./...`, gosec, and nancy, the dependency vulnerability scan) on every branch and tag, next to upstream's
`build.yml`, whose Go matrix the line keeps at the Go its `go` directive names (`GOTOOLCHAIN=local` there refuses an
older one, and upstream's pinned golangci-lint cannot read a newer one's export data), while the orb's job runs the
architect image's current Go. It is Giant Swarm CI only, never offered upstream, and never generated: the line's entry in
giantswarm/github declares `gen.ci.generate: false`, so the file is edited here directly.

nancy audits the modules of the built packages against OSS Index, by CVE id and module version. A finding with a
fixed version is a bump of that module in `go.mod` on the line, by hand (Renovate stays off, `renovate.json5`), a
`fix:` pull request of the fork's own; the next pin move keeps the higher version, so the bump never conflicts with
upstream for long. A finding without a fix is a time-boxed `.nancy-ignore` entry, `<id> until=<date> # <module>@<version>`
with its tracking issue in a comment line, removed when the issue closes or the date passes; the scan is never
bypassed, and no entry outlives its finding. GO-2026-5932 (`golang.org/x/crypto/openpgp`, unmaintained) is no entry:
the carried patch took the package out of the line's imports, and nancy knows no such id. The line's x/crypto is
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
if something else requires it. The import goes back to upstream's path (then its v2, `…/go-selfupdate/v2`, where
upstream plans the openpgp change) at the first upstream release that contains every carried patch, and this line
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

| Patch | Why the CLIs need it | Upstream | State |
|---|---|---|---|
| `fix: replace unmaintained golang.org/x/crypto/openpgp with ProtonMail fork` | `validate.go` imports `golang.org/x/crypto/openpgp` unconditionally for `PGPValidator`, so every CLI that embeds the module reaches it (`selfupdate.init` → `openpgp.init`) and `govulncheck` reports GO-2026-5932 (unmaintained, unsafe by design, no fixed version) in each of them, although none verifies a PGP signature: they verify cosign bundles. The patch switches to `github.com/ProtonMail/go-crypto/openpgp`, the maintained drop-in fork the advisory recommends, with `cloudflare/circl` at v1.6.3. | [creativeprojects/go-selfupdate#58](https://github.com/creativeprojects/go-selfupdate/pull/58), open since 2026-07-24; the maintainer holds it for a v2 ([#57](https://github.com/creativeprojects/go-selfupdate/issues/57)) | carried, [giantswarm/go-selfupdate#1](https://github.com/giantswarm/go-selfupdate/pull/1) |
| `fix: find GitHub releases when the release list comes back empty` | GitHub intermittently answers the release list `200 []`, so `DetectLatest` finds nothing and every CLI refuses to update; `GitHubSource.ListReleases` asks an empty list again and adds `/releases/latest`, which answers reliably. | to file — prepared in the fork: branch `upstream/github-empty-release-list` | carried, [giantswarm/go-selfupdate#20](https://github.com/giantswarm/go-selfupdate/issues/20) |
| `fix: move go-github to v93` | The release lookup's errors are go-github types (`*github.RateLimitError`), which a CLI matches with `errors.As`; another major's type never matches, so a CLI on a newer go-github reads a spent rate limit as a network failure. v93 `NewClient` takes options and returns an error. | [creativeprojects/go-selfupdate#63](https://github.com/creativeprojects/go-selfupdate/pull/63), from branch `upstream/go-github-v93` | carried |
| `chore: module path github.com/giantswarm/go-selfupdate` | Lets consumers import the line directly, so a library passes it on and `go install` users get it (a `replace` reaches neither). | never: Giant Swarm-specific | carried, see [Module path](#module-path) |
