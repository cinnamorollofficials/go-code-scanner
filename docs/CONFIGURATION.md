# Configuration Reference

Go Code Scanner reads strict JSON configuration schema version `1`. Unknown
fields and trailing JSON values are rejected. Version 1 is extended only with
backward-compatible optional fields; incompatible changes require a new schema
and migration guidance.

Validate a file before use:

```sh
security-review config validate security-review.json
```

## Path resolution

When a config file is loaded, `root: "."` (or an empty root) resolves to the
config file's directory. Another relative root resolves from that directory.
Afterward, relative `output`, `rule_files`, `suppression_file`, `baseline_file`,
and `cache.directory` paths resolve from the project root.

Project paths must remain inside `root`; traversal and unsafe symlink targets are
rejected at the relevant read/write boundary.

## Minimal configuration

```json
{
  "version": 1,
  "project": "my-service",
  "root": ".",
  "mode": "full",
  "output": "security_findings.json",
  "fail_on": "CRITICAL",
  "include_extensions": [".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"],
  "exclude_directories": [".git", "node_modules", "vendor", "dist", "build", ".next", "out", "bin"],
  "exclude_files": ["security_findings.json", "package-lock.json"],
  "rule_files": [],
  "suppression_file": ".security-ignore",
  "baseline_file": ".security-baseline.json",
  "workers": 4,
  "pattern_max_file_bytes": 2097152,
  "pattern_max_line_bytes": 1048576,
  "scanners": {}
}
```

Fields without an `omitempty` JSON tag must be present when authoring a complete
configuration. Loading starts from defaults, so omitted fields retain their
default values before the JSON overlay is validated.

## Top-level fields

| Field | Type | Default and validation |
| --- | --- | --- |
| `version` | integer | `1`; any other value is rejected. |
| `project` | string | `security-review`; must not be empty. |
| `root` | string | `.`; normalized to an absolute path after validation. |
| `mode` | string | `full`; one of `full`, `changed`, `staged`. |
| `output` | string | `security_findings.json`; safe project path. |
| `fail_on` | severity | `CRITICAL`; fallback threshold for domains absent from `policy`. |
| `include_extensions` | string array | `.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.yaml`, `.yml`, `.json`. |
| `exclude_directories` | string array | `.git`, `node_modules`, `vendor`, `dist`, `build`, `.next`, `out`, `bin`. |
| `exclude_files` | string array | `security_findings.json`, `package-lock.json`. |
| `rule_files` | string array | Empty; each entry must resolve inside the project. |
| `suppression_file` | string | `.security-ignore`; safe project path. |
| `baseline_file` | string | `.security-baseline.json`; safe project path. |
| `workers` | integer | `GOMAXPROCS`; at least `1`. |
| `pattern_max_file_bytes` | integer | `2097152`; at least `1`. Oversized input makes the pattern scanner partial. |
| `pattern_max_line_bytes` | integer | `1048576`; at least `1`. |
| `quality_max_file_bytes` | integer | `0` (disabled); cannot be negative. Positive values create quality findings. |
| `quality_max_line_length` | integer | `0` (disabled); cannot be negative. Counts Unicode code points. |
| `scanners` | object | Scanner ID to scanner definition. IDs must be non-empty. |
| `profiles` | object | Profile name to unique scanner IDs. |
| `offline_profiles` | string array | `fast`; every name must exist in `profiles` and be unique. |
| `policy` | object | Domain to severity threshold. Domains and severities must be valid. |
| `hooks` | object | Hook configuration described below. |
| `supply_chain` | object | Dependency and license policy. |
| `governance` | object | Required files/headers, ownership, and suppression requirements. |
| `architecture` | object | Go layer boundaries and cycle detection. |
| `cache` | object | Content-addressed scanner cache policy. |
| `frontend` | object | Browser client scanning policy. |

Severities are `CRITICAL`, `HIGH`, `MEDIUM`, and `LOW`. Domains are `quality`,
`reliability`, `hardening`, `security`, `supply_chain`, and `governance`.

## Frontend policy (`frontend`)

Configure built-in browser client scanning, framework detection, and sanitizer overrides:

```json
{
  "frontend": {
    "enabled": true,
    "frameworks": ["react", "nextjs", "vue", "svelte"],
    "client_roots": ["src/client", "app"],
    "server_roots": ["src/server", "server"],
    "shared_roots": ["src/shared"],
    "recognize_sanitizers": ["dompurify", "sanitize-html"],
    "detect_import_cycles": true,
    "detect_client_server_boundaries": true
  }
}
```

| Field | Type | Description |
| --- | --- | --- |
| `enabled` | boolean | Enables built-in native frontend client scanning (`frontend` scanner). |
| `frameworks` | string array | Target frameworks to scan (`react`, `nextjs`, `vue`, `nuxt`, `svelte`, `sveltekit`). |
| `client_roots` | string array | Directory paths categorized as client code. |
| `server_roots` | string array | Directory paths categorized as server code. |
| `shared_roots` | string array | Directory paths categorized as shared code. |
| `include_extensions` | string array | Additional extensions to evaluate for frontend rules. |
| `recognize_sanitizers` | string array | Function names recognized as valid HTML sanitizers. |
| `detect_import_cycles` | boolean | Default `true`; detects circular import dependencies between frontend modules. |
| `detect_client_server_boundaries` | boolean | Default `true`; prevents server-only modules from being imported into client code. |

## Profiles and offline behavior

Default profiles are:

```json
{
  "profiles": {
    "fast": ["pattern"],
    "standard": ["pattern", "govulncheck"],
    "full": ["pattern", "govulncheck"],
    "frontend": ["pattern", "frontend", "tsc", "biome", "eslint", "semgrep"]
  },
  "offline_profiles": ["fast", "frontend"]
}
```

Scanner IDs inside a profile must be non-empty and unique. A profile may refer
to an unavailable optional scanner; its `on_missing` policy determines whether
that is skipped or failed. In an offline profile, a scanner whose definition or
adapter sets `requires_network` is skipped without execution.

The CLI-only selected profile is not serialized; use `scan --profile <name>`.

## Policy

```json
{
  "fail_on": "CRITICAL",
  "policy": {
    "quality": "HIGH",
    "reliability": "MEDIUM",
    "hardening": "HIGH",
    "security": "HIGH",
    "supply_chain": "HIGH",
    "governance": "HIGH"
  }
}
```

For each domain, `policy[domain]` overrides `fail_on`. The CLI `--fail-on`
disables the domain map for that invocation and acts as a global override.

## Scanner definitions

Every `scanners.<id>` object supports these common fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `enabled` | boolean | Execute the scanner when selected. Disabled scanners are skipped. |
| `required` | boolean | Convert scanner failure into an operational run error. |
| `timeout` | duration string | Positive Go duration such as `500ms`, `30s`, or `2m`. Empty uses no scanner-specific deadline. |
| `type` | string | Empty/`pattern`, `command`, or `adapter`. `pattern` is reserved for ID `pattern`. |
| `version` | string | Included in scanner status and cache identity. |
| `requires_network` | boolean | Skip in offline profiles. Adapter defaults may set this automatically. |

The built-in pattern scanner is always registered. A `pattern` entry controls
enabled/required/timeout behavior but cannot be replaced by command or adapter
configuration.

### Command scanners

```json
{
  "scanners": {
    "custom-lint": {
      "enabled": true,
      "required": false,
      "timeout": "30s",
      "type": "command",
      "command": ["custom-lint", "--json", "."],
      "domain": "quality",
      "workspace": "staged",
      "on_missing": "skip",
      "finding_exit_codes": [1],
      "severity": "HIGH",
      "category": "static_analysis",
      "description": "Custom lint reported a finding",
      "version": "1",
      "max_output_bytes": 65536,
      "snapshot_max_files": 100000,
      "snapshot_max_bytes": 2147483648,
      "output_format": "json-lines",
      "environment": ["CUSTOM_LINT_CONFIG"],
      "requires_network": false
    }
  }
}
```

Command-specific fields:

| Field | Default and validation |
| --- | --- |
| `command` | Non-empty argument array. The first item is an executable name or clean absolute path; shell strings are not accepted. |
| `domain` | Required valid domain. |
| `workspace` | `root` by default; `root` or `staged`. Staged scanners execute only in staged mode. |
| `on_missing` | `fail` by default; `skip` or `fail`. |
| `finding_exit_codes` | `[1]`; unique nonzero integers. |
| `severity` | Required valid severity. |
| `category`, `description` | Required non-empty strings. |
| `max_output_bytes` | `65536`; at least `1`. Applies separately to stdout/stderr and structured output files. |
| `snapshot_max_files` | `100000`; at least `1`. |
| `snapshot_max_bytes` | `2147483648`; at least `1`. |
| `output_format` | `exit-code`; one of `exit-code`, `json-lines`, or `paths`. Adapter-specific JSON parsers are selected by adapters. |
| `environment` | Empty; names must be valid uppercase-style environment identifiers and unique. Only default safe variables plus this allowlist are forwarded. |

`options` is accepted as a reserved object for compatibility but is not consumed
by current command or adapter implementations. Prefer the explicit fields.

### Adapter scanners

```json
{
  "scanners": {
    "gitleaks": {
      "enabled": true,
      "required": false,
      "type": "adapter",
      "adapter": "gitleaks",
      "workspace": "staged",
      "on_missing": "skip",
      "timeout": "45s",
      "max_output_bytes": 1048576
    }
  }
}
```

`adapter` must be one of:

| Adapter | Domain | Default behavior |
| --- | --- | --- |
| `gofmt` | Quality | Reports paths printed by `gofmt -l`. |
| `go-vet` | Reliability | Runs `go vet ./...`. |
| `go-test` | Reliability | Runs `go test ./...`. |
| `govulncheck` | Supply chain | Parses streaming JSON; requires network. |
| `gosec` | Security | Parses gosec JSON. |
| `gitleaks` | Security | Uses redacted JSON output file. |
| `trivy` | Supply chain | Parses filesystem JSON; requires network. |
| `osv-scanner` | Supply chain | Parses recursive source JSON; requires network. |
| `semgrep` | Security | Parses Semgrep JSON. |

`args` replaces an adapter's arguments after the executable. `workspace`,
`on_missing`, `environment`, `max_output_bytes`, `snapshot_max_files`, and
`snapshot_max_bytes` override adapter execution options. Parser, domain,
severity, finding exit codes, and network behavior come from the preset.

## Hooks

```json
{
  "hooks": {
    "pre_commit": {
      "enabled": true,
      "profile": "fast",
      "staged_only": true,
      "new_only": true
    },
    "commit_msg": {
      "enabled": true,
      "message_pattern": "^(feat|fix|docs|test|refactor)(\\([^)]+\\))?: .+",
      "max_subject_length": 72
    },
    "pre_push": {
      "enabled": true,
      "profile": "standard",
      "staged_only": false,
      "new_only": false
    }
  }
}
```

Enabled pre-commit/pre-push hooks require an existing profile. `staged_only`
adds staged mode; `new_only` applies baseline-aware policy. Enabled commit-msg
hooks accept an optional valid Go regular expression and a nonnegative maximum
subject length. Defaults enable pre-commit with `fast`, staged-only, and
new-only; commit-msg defaults to length 72 but is disabled; pre-push defaults to
profile `standard` but is disabled.

## Supply-chain policy

`dependency_allowlist`, `dependency_denylist`, `license_allowlist`, and
`license_denylist` are arrays of unique, non-empty `filepath.Match` patterns.
Comparison is case-insensitive for duplicate validation. Allow/deny policies are
evaluated by built-in dependency/license checks; deny policy takes precedence
when a value matches.

```json
{
  "supply_chain": {
    "dependency_allowlist": ["github.com/my-org/*"],
    "dependency_denylist": ["example.invalid/deprecated*"],
    "license_allowlist": ["MIT", "Apache-2.0"],
    "license_denylist": ["AGPL-*"]
  }
}
```

## Governance policy

- `required_files`: unique safe repository-relative paths.
- `required_headers`: objects containing unique `id`, non-empty path patterns,
  a valid Go regex `pattern`, optional nonnegative `max_lines`, optional
  severity, description, and recommendation.
- `ownership_file`: safe repository-relative path; defaults to `CODEOWNERS`.
- `ownership_rules`: unique one-line `path`, non-empty unique owners without
  whitespace, and optional severity.
- `suppression_requirements`: non-empty rule-ID glob patterns plus at least one
  of `require_ticket` or `require_approver`.

```json
{
  "governance": {
    "required_files": ["SECURITY.md", "CODEOWNERS"],
    "required_headers": [{
      "id": "copyright",
      "paths": ["**/*.go"],
      "pattern": "^// Copyright",
      "max_lines": 5,
      "severity": "HIGH",
      "description": "Approved copyright header is required"
    }],
    "ownership_file": ".github/CODEOWNERS",
    "ownership_rules": [{
      "path": "/internal/security/",
      "owners": ["@security-team"],
      "severity": "HIGH"
    }],
    "suppression_requirements": [{
      "rule_ids": ["security/*"],
      "require_ticket": true,
      "require_approver": true
    }]
  }
}
```

## Architecture policy

Layers require unique names and non-empty slash-style path patterns. Forbidden
dependencies must reference defined layers and each direction must be unique.
`detect_cycles` enables deterministic Go import-cycle findings.

```json
{
  "architecture": {
    "layers": [
      {"name": "domain", "paths": ["internal/domain/*.go"]},
      {"name": "infra", "paths": ["internal/infra/*.go"]}
    ],
    "forbidden_dependencies": [
      {"from": "domain", "to": "infra"}
    ],
    "detect_cycles": true
  }
}
```

## Cache policy

```json
{
  "cache": {
    "enabled": true,
    "directory": ".go-code-scanner-cache",
    "max_age": "168h",
    "max_bytes": 268435456
  }
}
```

When enabled, `directory` must be a safe non-symlink project path, `max_age`
must be a positive Go duration, and `max_bytes` must be at least `1`. Defaults
are seven days and 256 MiB. Cache entries use mode `0600` where supported and do
not retain finding snippets.

## Frontend policy

Frontend policy configures browser-client scanning scope, framework detection, and security rules.

```json
{
  "frontend": {
    "enabled": true,
    "frameworks": ["vanilla", "react", "next", "vue", "nuxt", "svelte", "sveltekit"],
    "client_roots": ["src/client", "components"],
    "server_roots": ["src/server", "api"],
    "shared_roots": ["src/shared", "lib"],
    "include_extensions": [".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts", ".html", ".vue", ".svelte"],
    "recognize_sanitizers": ["DOMPurify.sanitize", "sanitizeHtml"],
    "detect_import_cycles": true,
    "detect_client_server_boundaries": true
  }
}
```

`client_roots`, `server_roots`, and `shared_roots` must be safe project-relative paths. `frameworks` accepts `vanilla`, `react`, `next`, `vue`, `nuxt`, `svelte`, and `sveltekit`. `detect_import_cycles` and `detect_client_server_boundaries` default to `true` when frontend scanning is enabled. An omitted `frontend` block preserves default non-frontend behavior.

## Full example

```json
{
  "version": 1,
  "project": "payments-service",
  "root": ".",
  "mode": "full",
  "output": "artifacts/security-review.json",
  "fail_on": "CRITICAL",
  "include_extensions": [".go", ".yaml", ".yml", ".json"],
  "exclude_directories": [".git", "vendor", "dist", "build"],
  "exclude_files": ["security-review.json"],
  "rule_files": ["config/custom-rules.json"],
  "suppression_file": ".security-ignore",
  "baseline_file": ".security-baseline.json",
  "workers": 4,
  "pattern_max_file_bytes": 2097152,
  "pattern_max_line_bytes": 1048576,
  "quality_max_file_bytes": 1048576,
  "quality_max_line_length": 120,
  "scanners": {
    "pattern": {"enabled": true, "required": true, "timeout": "10s"},
    "go-vet": {"enabled": true, "type": "adapter", "adapter": "go-vet", "workspace": "staged", "on_missing": "fail", "timeout": "30s"},
    "govulncheck": {"enabled": true, "type": "adapter", "adapter": "govulncheck", "on_missing": "skip", "timeout": "2m"},
    "gitleaks": {"enabled": true, "type": "adapter", "adapter": "gitleaks", "workspace": "staged", "on_missing": "skip", "timeout": "45s", "max_output_bytes": 1048576}
  },
  "profiles": {
    "fast": ["pattern", "go-vet", "gitleaks"],
    "standard": ["pattern", "go-vet", "govulncheck"],
    "full": ["pattern", "go-vet", "govulncheck", "gitleaks"]
  },
  "offline_profiles": ["fast"],
  "policy": {
    "quality": "HIGH",
    "reliability": "MEDIUM",
    "hardening": "HIGH",
    "security": "HIGH",
    "supply_chain": "HIGH",
    "governance": "HIGH"
  },
  "hooks": {
    "pre_commit": {"enabled": true, "profile": "fast", "staged_only": true, "new_only": true},
    "commit_msg": {"enabled": true, "message_pattern": "^(feat|fix|docs|test): .+", "max_subject_length": 72},
    "pre_push": {"enabled": true, "profile": "standard"}
  },
  "supply_chain": {
    "dependency_denylist": ["example.invalid/deprecated*"],
    "license_allowlist": ["MIT", "Apache-2.0"]
  },
  "governance": {
    "required_files": ["SECURITY.md", ".github/CODEOWNERS"],
    "ownership_file": ".github/CODEOWNERS",
    "ownership_rules": [{"path": "/internal/security/", "owners": ["@security-team"], "severity": "HIGH"}],
    "suppression_requirements": [{"rule_ids": ["security/*"], "require_ticket": true, "require_approver": true}]
  },
  "architecture": {
    "layers": [
      {"name": "domain", "paths": ["internal/domain/*.go"]},
      {"name": "infra", "paths": ["internal/infra/*.go"]}
    ],
    "forbidden_dependencies": [{"from": "domain", "to": "infra"}],
    "detect_cycles": true
  },
  "cache": {"enabled": true, "directory": ".go-code-scanner-cache", "max_age": "168h", "max_bytes": 268435456}
}
```

Validate this structure with the CLI after adapting paths and scanner choices.
External tool versions and installation are intentionally managed outside the
configuration file.
