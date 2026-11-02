# Maintainer release procedure

Release preparation is a reviewed PR. Publishing is a separate action after
that PR has merged and CI has passed on the exact commit to be tagged.

## Prepare the version

1. Move the completed CHANGELOG.md Unreleased entries into the version section,
   set the intended release date, and update the version/compare links. Keep an
   empty Unreleased section for later changes.
2. Add `docs/releases/<tag>.md` with the release summary and compatibility notes.
   Use absolute, tag-pinned links in that file because it becomes the GitHub
   release body. Keep migration guides, README, and translated guide links in
   sync with the version's approved numerical contract.
3. Review the PR, including `@codex review`, and wait for Test and both
   Correctness jobs. These cover workflow lint, race, BSON, mapstructure,
   reference regeneration, independent numerical checks, and bounded fuzzing.
   For local equivalents, see [CONTRIBUTING.md](../CONTRIBUTING.md) and the
   [correctness gates](correctness-gates.md).

For v0.8.0, the intended date is 2026-11-02. If publication happens on a different
date, adjust its changelog date before tagging. The [release notes](releases/v0.8.0.md)
and [upgrade guide](numeric-contract-migration.md) describe changes from v0.7.0.
No go.mod version field or `/v0` import suffix is needed; the existing module
path and Go 1.26.0 minimum are unchanged.

## Publish after merge

Use a clean main checkout without local-only commits. Fetch and fast-forward it:

```sh
git switch main
git fetch origin --prune
git merge --ff-only origin/main
git status --short
git rev-parse HEAD
```

Confirm that HEAD includes the release preparation and that GitHub's **main
branch CI on this exact SHA** has passed, including Test and Correctness
(amd64/386). PR checks on a different SHA do not establish this. Check that the
version is not already tagged and its reviewed notes exist:

```sh
git tag --list v0.8.0
test -s docs/releases/v0.8.0.md
```

The tag listing must be empty. When publication is authorized, these commands
create the annotated tag at HEAD and trigger the release workflow:

```sh
git tag -a v0.8.0 -m "release v0.8.0"
git push origin v0.8.0
```

The workflow validates a nonempty `docs/releases/<tag>.md`, runs the default,
BSON, and mapstructure tests, then publishes using the action's
[body_path input](https://github.com/softprops/action-gh-release/tree/v2#external-release-notes)
with generated change notes enabled. Check the completed Release workflow and
the resulting GitHub release body, tag target, and upgrade links. It publishes
source releases; no prebuilt library binaries are required.

Once the tag is available to the configured Go proxy, verify a consumer without
the integration module's local replace directive can resolve it:

```sh
go list -m -json github.com/exc-works/decimal@v0.8.0
```

If publishing fails after the tag push, inspect the failed run before retrying.
Do not move or reuse a published version tag; a source correction after
publication requires a new version.
