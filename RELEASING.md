# Releasing

This document describes how to publish a new version of `lettr-go`.

Go modules are published via git tags — there is no package registry to push to. Once a tag is pushed to GitHub, Go's module proxy (`proxy.golang.org`) picks it up on first fetch, and users can run `go get github.com/lettr-com/lettr-go@vX.Y.Z`.

## Versioning Policy

This project follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (`vX.0.0`) — incompatible API changes.
- **MINOR** (`v0.X.0`) — backward-compatible feature additions.
- **PATCH** (`v0.0.X`) — backward-compatible bug fixes.

While the SDK is on `v0.x`, minor versions may contain breaking changes. Once `v1.0.0` is released, the public API becomes stable and breaking changes require a new major version.

### Major version ≥ 2

Go enforces the "import compatibility rule": a module at `v2` or higher must change its module path, e.g. `module github.com/lettr-com/lettr-go/v2` in `go.mod`. `v0` → `v1` does **not** require a path change.

## Release Checklist

1. **Update the `Version` const** in `lettr.go` (used in the `User-Agent` header).
2. **Update `CHANGELOG.md`**:
   - Move entries from `[Unreleased]` into a new `[X.Y.Z] - YYYY-MM-DD` section.
   - Summarize user-visible changes under `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.
   - Clearly label any breaking changes.
3. **Run tests and static analysis**:
   ```bash
   go build ./...
   go vet ./...
   go test ./...
   ```
4. **Commit** the version bump and changelog:
   ```bash
   git add lettr.go CHANGELOG.md
   git commit -m "Release vX.Y.Z"
   ```
5. **Tag the release** (annotated tag, `v` prefix required by Go):
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   ```
6. **Push the branch and tag**:
   ```bash
   git push origin main
   git push origin vX.Y.Z
   ```
7. **Create a GitHub release** (optional but recommended):
   ```bash
   gh release create vX.Y.Z --notes-file <(sed -n '/^## \[X.Y.Z\]/,/^## \[/p' CHANGELOG.md | sed '$d')
   ```
   Or use `--generate-notes` to auto-generate release notes from commit messages.
8. **Verify the release is resolvable**:
   ```bash
   GOPROXY=proxy.golang.org go list -m github.com/lettr-com/lettr-go@vX.Y.Z
   ```

## Fixing a bad release

Tags in the Go module proxy are **immutable** — once published, a version cannot be changed. If a release is broken:

1. Publish a new patch (`vX.Y.Z+1`) with the fix.
2. Optionally retract the bad version by adding a `retract` directive to `go.mod`:
   ```go
   retract vX.Y.Z // broken: <reason>
   ```
   Then release a new version including the retraction. Users running `go get` will see a warning about the retracted version.

Do **not** delete or force-push git tags — the proxy has already cached them and users depending on the version will break.
