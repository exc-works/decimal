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
gofmt -w *.go
go vet ./...
go test ./...
go test -race ./...
go test -cover ./...
```

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
