---
title: Installation
description: Install Go Code Scanner from pre-compiled binaries, Go toolchain, or source.
---

# Installation

`security-review` is distributed as a self-contained, single binary with zero external runtime dependencies.

## Prerequisites

- **Operating System**: Linux (amd64/arm64), macOS (amd64/arm64), or Windows (amd64/arm64).
- **Go Toolchain** (Optional): Go 1.25+ if compiling from source or installing via `go install`.

## Option 1: Install via Go Toolchain

If you have Go 1.25+ installed, install `security-review` directly:

```sh
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest
```

Ensure `$GOPATH/bin` (or `~/go/bin`) is in your system `PATH`:

```sh
security-review version
```

## Option 2: Build from Source

Clone the repository and build the binary locally:

```sh
git clone https://github.com/cinnamorollofficials/go-code-scanner.git
cd go-code-scanner
go build -o security-review ./cmd/security-review
./security-review version
```

## Option 3: Pre-compiled Release Binaries

Pre-compiled release archives and Ed25519-signed provenance manifests will be available from the [GitHub Releases](https://github.com/cinnamorollofficials/go-code-scanner/releases) page upon tagged releases.

Verify release integrity using checksums:

```sh
# Verify release archive checksums
security-review release checksums verify --manifest SHA256SUMS --directory dist/
```

The command returns exit code `0` only when every archive listed in the manifest
matches its SHA-256 digest. A mismatch returns exit code `1`; invalid arguments
or an unreadable manifest return exit code `2`.
