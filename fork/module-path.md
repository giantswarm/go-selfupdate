# Module path github.com/giantswarm/go-selfupdate

- **Status:** carried, Giant Swarm-specific: kept for as long as the line exists
- **Upstream:** never; a carried patch goes upstream without this commit, on upstream's module path
  ([Contributing](../FORK.md#contributing))
- **Our commit:** `44247f35` `chore: module path github.com/giantswarm/go-selfupdate`, rebase-merged in
  [giantswarm/go-selfupdate#3](https://github.com/giantswarm/go-selfupdate/pull/3), first released in v1.6.1; the
  regrouped imports of `update.go` and `internal/path_test.go` came with `2646ffa`
  ([giantswarm/go-selfupdate#9](https://github.com/giantswarm/go-selfupdate/pull/9))

## Change

The `module` line of `go.mod` is `github.com/giantswarm/go-selfupdate`, and every import of the module's own packages
names it: `path.go`, `update.go`, `updater.go`, `update/apply.go`, `update/doc.go`, `cmd/source.go`, the
`cmd/detect-latest-release` and `cmd/get-release` examples with their READMEs, and `internal/path_test.go`. In
`update.go` and `internal/path_test.go` the line's own packages sit in an import group of their own, as the line's
`goimports -local` check wants ([Module path](../FORK.md#module-path)).

## Reason

A consumer imports the line directly, in place of upstream's path, with no `replace`: `go install <module>@<version>`
ignores every `replace` of the installed module, and a library cannot pass one on to its importers. Under its own path
the line reaches every CLI through selfupdate-cosign ([Consumers](../FORK.md#consumers)), and its own semver tags live
in its own version space ([Releases](../FORK.md#releases)). The imports go back to upstream's path, and this change
goes with the line, at the first upstream release that contains every other carried patch.
