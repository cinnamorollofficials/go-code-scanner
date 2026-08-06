---
title: Installation
description: Install Go Code Scanner from pre-compiled binaries, Go toolchain, or source.
---

# Installation

`security-review` is distributed as a self-contained, single binary with zero external runtime dependencies.

## Prerequisites

- **Operating System**: Linux (amd64/arm64), macOS (amd64/arm64), or Windows (amd64/arm64).
- **Go Toolchain** (Optional): Go 1.22+ if compiling from source or installing via `go install`.

## Option 1: Pre-compiled Release Binary

Download the latest pre-compiled binary from the [GitHub Releases](https://github.com/cinnamorollofficials/go-code-scanner/releases) page.

### Linux & macOS

```sh
# Download archive for your platform
curl -sSL https://github.com/cinnamorollofficials/go-code-scanner/releases/latest/download/security-review-linux-amd64.tar.gz | tar -xz

# Move binary to system PATH
sudo mv security-review /usr/local/bin/
security-review version
```

### Windows (PowerShell)

```powershell
# Download zip archive and extract
Invoke-WebRequest -Uri "https://github.com/cinnamorollofficials/go-code-scanner/releases/latest/download/security-review-windows-amd64.zip" -OutFile "security-review.zip"
Expand-Archive -Path "security-review.zip" -DestinationPath "C:\Program Files\security-review"

# Verify execution
security-review.exe version
```

## Option 2: Install via Go Toolchain

If you have Go installed, install `security-review` directly to your `$GOPATH/bin`:

```sh
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest
```

## Option 3: Build from Source

Clone the repository and build the binary locally:

```sh
git clone https://github.com/cinnamorollofficials/go-code-scanner.git
cd go-code-scanner
go build ./cmd/security-review
```
