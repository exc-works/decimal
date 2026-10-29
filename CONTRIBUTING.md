# Contributing

Thanks for your interest in improving `decimal`.

## Development Setup

1. Install Go matching the version in `go.mod`.
2. Clone the repository.
3. Run tests:

```bash
go test ./...
```

## Local Quality Checks

Run these before opening a PR:

```bash
export GOCACHE="$PWD/.cache/go-build"
gofmt -w ./*.go ./integration/mapstructure/*.go
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
go vet ./...
go test ./...
go test -race ./...
go test -cover ./...
go vet -tags bson ./...
go test -race -tags bson ./...
```

The mapstructure integration is a separate Go module. Root-level `./...`
commands do not include it, so also run:

```bash
cd integration/mapstructure
go vet ./...
go test -race ./...
```

The project-local cache above is ignored by Git and avoids dependence on the
shared Go build cache. If a race build reports an unexpected toolchain error
such as `runtime/race: package testmain: cannot find package`, first retry
with a fresh cache, using the same Go version:

```bash
GOCACHE="$(mktemp -d "${TMPDIR:-/tmp}/decimal-go-build.XXXXXX")" go test -race ./...
```

Run this command from the repository root. If it succeeds, the shared cache is
implicated; no library concurrency fix is needed for that build failure. If it
still fails, include `go version`, `go env GOROOT GOCACHE GOTOOLCHAIN`, and the
complete error in the issue report. CI uses a runner-local cache and checks
the default build, BSON build, and mapstructure integration separately.

## Pull Request Guidelines

1. Keep PRs focused and small.
2. Add or update tests for behavior changes.
3. Update `README.md` and `CHANGELOG.md` for user-visible changes.
4. Use clear commit messages and PR descriptions.

## Release Process

1. Update `CHANGELOG.md` under a new version heading.
2. Commit and merge to the default branch.
3. Create an annotated tag, for example:

```bash
git tag -a v0.1.0 -m "release v0.1.0"
git push origin v0.1.0
```

4. The `release.yml` workflow publishes a GitHub release from the tag.
