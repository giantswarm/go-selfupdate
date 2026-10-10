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
(`git merge --no-ff v<version>`), resolving conflicts so the carried patches stay in effect on the new upstream code and
the fork's own files stay as the line has them; upstream's tests and `go vet ./...` pass before it opens. It lands with
a merge commit, so the new upstream tag is an ancestor of the line and `git describe`, like the pseudo-version
consumers pin, counts from it.

## Consumers

A consumer takes the line through a `replace` directive, keeping upstream's module path and release in its `require`,
pinned to a commit of `giantswarm` by its pseudo-version (`go list -m github.com/giantswarm/go-selfupdate@<commit>`
prints it):

```
replace github.com/creativeprojects/go-selfupdate v1.6.0 => github.com/giantswarm/go-selfupdate v1.6.1-0.<date>-<commit>
```

The module path in `go.mod` stays `github.com/creativeprojects/go-selfupdate`, so the line is only usable through
`replace`, and a library that takes it (selfupdate-cosign) cannot pass it on: each CLI carries the same line in its own
`go.mod`. The `replace` names the upstream version on its left, so a Renovate bump of go-selfupdate in a consumer
leaves the line, and the consumer's `govulncheck` shows whether upstream carries the patch. The `replace` goes away,
and this line retires, at the first upstream release that contains every carried patch; upstream plans the openpgp
change for a v2, whose module path (`…/go-selfupdate/v2`) the consumers then import.

## Releases

None. A Go module needs no build or publish step, and a pseudo-version names exactly the commit of the protected
`giantswarm` branch a consumer runs, so the line cuts no tags of its own (an own tag would sit in upstream's `v1`
version space, since a module's major is part of its import path). Upstream tags on the line are upstream's, kept for
`git describe`.

## Carried patches

| Patch | Why the CLIs need it | Upstream | State |
|---|---|---|---|
| `fix: replace unmaintained golang.org/x/crypto/openpgp with ProtonMail fork` | `validate.go` imports `golang.org/x/crypto/openpgp` unconditionally for `PGPValidator`, so every CLI that embeds the module reaches it (`selfupdate.init` → `openpgp.init`) and `govulncheck` reports GO-2026-5932 (unmaintained, unsafe by design, no fixed version) in each of them, although none verifies a PGP signature: they verify cosign bundles. The patch switches to `github.com/ProtonMail/go-crypto/openpgp`, the maintained drop-in fork the advisory recommends, with `cloudflare/circl` at v1.6.3. | [creativeprojects/go-selfupdate#58](https://github.com/creativeprojects/go-selfupdate/pull/58), open since 2026-07-24; the maintainer holds it for a v2 ([#57](https://github.com/creativeprojects/go-selfupdate/issues/57)) | carried, [giantswarm/go-selfupdate#1](https://github.com/giantswarm/go-selfupdate/pull/1) |
