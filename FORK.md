# The Giant Swarm line of go-selfupdate

This repository is Team Bumblebee's fork of [creativeprojects/go-selfupdate](https://github.com/creativeprojects/go-selfupdate),
the library every Giant Swarm command-line tool installs its releases with, behind the cosign bundle check of
[giantswarm/selfupdate-cosign](https://github.com/giantswarm/selfupdate-cosign). It carries an upstream fix the CLIs
need before upstream releases it, and answers one question from one place: **which go-selfupdate do the CLIs run, and
why does it differ from upstream?**

Tracking (this fork has issues disabled): [giantswarm/selfupdate-cosign#17](https://github.com/giantswarm/selfupdate-cosign/issues/17).

## Branches

| Branch | What it is |
|---|---|
| `giantswarm` (default) | **The line**: the upstream release tag the CLIs run ("the pin") + the carried patches + this fork's own files (`FORK.md`, `renovate.json5`, and upstream's `build.yml` running on this branch too). Every change is a pull request against it. |
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
alone, on upstream's module path; a change to it is cherry-picked onto upstream `main` and pushed there, never taken
from the line with the module path commit.

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
the minor or major of the upstream release a pin move takes. A tag is pushed on the merged commit of `giantswarm` once
its Build workflow is green; a Go module needs no build or publish step, so the Go proxy serves the tag as it is.
Upstream's tags on the line (v1.6.0 and older) name upstream's module path and are not consumable from this one.

## Carried patches

| Patch | Why the CLIs need it | Upstream | State |
|---|---|---|---|
| `fix: replace unmaintained golang.org/x/crypto/openpgp with ProtonMail fork` | `validate.go` imports `golang.org/x/crypto/openpgp` unconditionally for `PGPValidator`, so every CLI that embeds the module reaches it (`selfupdate.init` → `openpgp.init`) and `govulncheck` reports GO-2026-5932 (unmaintained, unsafe by design, no fixed version) in each of them, although none verifies a PGP signature: they verify cosign bundles. The patch switches to `github.com/ProtonMail/go-crypto/openpgp`, the maintained drop-in fork the advisory recommends, with `cloudflare/circl` at v1.6.3. | [creativeprojects/go-selfupdate#58](https://github.com/creativeprojects/go-selfupdate/pull/58), open since 2026-07-24; the maintainer holds it for a v2 ([#57](https://github.com/creativeprojects/go-selfupdate/issues/57)) | carried, [giantswarm/go-selfupdate#1](https://github.com/giantswarm/go-selfupdate/pull/1) |
| `chore: module path github.com/giantswarm/go-selfupdate` | Lets consumers import the line directly, so a library passes it on and `go install` users get it (a `replace` reaches neither). | never: Giant Swarm-specific | carried, see [Module path](#module-path) |
