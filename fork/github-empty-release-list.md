# Find GitHub releases when the release list comes back empty

- **Status:** carried
- **Upstream:** the issue [creativeprojects/go-selfupdate#64](https://github.com/creativeprojects/go-selfupdate/issues/64),
  open; the maintainer answered there that the unreleased `v2` branch (module path
  `github.com/creativeprojects/go-selfupdate/v2`) fixes it by asking the source for the latest release directly when no
  version is requested, and v1 has no fix. The pull request: to file — prepared in the fork: branch
  `upstream/github-empty-release-list` (one commit on upstream `main` @ `81c5a15`, merging cleanly with upstream `main`
  at v1.7.1). creativeprojects/go-selfupdate searched 2026-11-14, open and closed (`empty`, `latest release`,
  `GetLatestRelease`, `ListReleases`): no pull request covers it; the closed
  [creativeprojects/go-selfupdate#4](https://github.com/creativeprojects/go-selfupdate/issues/4) asked for
  `/releases/latest` to save an API call.
- **Our commit:** `686a6b27` `fix: find GitHub releases when the release list comes back empty`, rebase-merged with its
  ledger entry in [giantswarm/go-selfupdate#22](https://github.com/giantswarm/go-selfupdate/pull/22), which closed
  [giantswarm/go-selfupdate#20](https://github.com/giantswarm/go-selfupdate/issues/20); first released in v1.6.7. The
  prepared branch's commit `9871011b` is the same patch (`git patch-id`).

## Change

`github_source.go`: `GitHubSource.ListReleases` asks again when the release list comes back empty, twice, waiting
0.5 s and 1 s, and asks `GET /repos/{owner}/{repo}/releases/latest` as well, adding that release when the list lacks
it, so a list that stays empty still finds a release. A 404 from the latest lookup means no release yet; a failed latest
lookup is an error only when the list is empty too. `findReleaseAndAsset` still picks the newest release with a
matching asset, so a latest release whose assets are still uploading falls back to the newest listed one that has
them. `github_source_test.go` covers a list empty once, a list that stays empty, a cancelled retry, a latest release
without assets and a list that already holds the latest release.

## Reason

GitHub intermittently answers `GET /repos/{owner}/{repo}/releases` with `200` and an empty array (no `Link` header),
authenticated or anonymous, so `DetectLatest` finds no release and every Giant Swarm CLI refuses to update, while
`/releases/latest` answers reliably. The latest release is the most recently published one, not necessarily the
highest version, which is why the list is asked again before the latest stands alone. The patch drops at a re-pin onto
an upstream release that finds a release when the list comes back empty.
