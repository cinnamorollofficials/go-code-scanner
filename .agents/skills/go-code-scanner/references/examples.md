# Integration Examples

## 1. Local Pre-commit Hook Setup

```sh
security-review hook install pre-commit --root .
```

## 2. GitHub Actions Workflow Integration

```yaml
name: Security Review Gate
on: [push, pull_request]

jobs:
  security-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Build & Run Security Review
        run: |
          go build -o /tmp/security-review ./cmd/security-review
          /tmp/security-review scan --profile fast --ci --format sarif --output security.sarif
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: security.sarif
```
