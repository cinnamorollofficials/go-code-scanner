# security-review

Reusable Go library and command-line scanner for source-code security review.

The project is currently being extracted from HINT Core. The public library
owns orchestration and normalized findings, while repository-specific rules,
suppressions, paths, and CI policy remain in the consuming repository.

## Usage

```sh
go run ./cmd/security-review scan --root /path/to/project
go run ./cmd/security-review scan --root /path/to/project --changed
go run ./cmd/security-review scan --root /path/to/project --staged --ci
go run ./cmd/security-review config validate security-review.json
```

Configuration files use JSON in the initial release. The library performs the
native pattern scan itself and writes a versioned, redacted JSON report.

## Development

```sh
go test ./...
go vet ./...
```

The module intentionally uses only the Go standard library during its initial
foundation phase.
